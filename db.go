package activitypublog

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/uptrace/bun"
)

type DAccount struct {
	bun.BaseModel `bun:"table:account"`
	Id            string `bun:",pk"`
	Username      string
	Host          string `bun:",pk"`
	AllFetched    bool   `bun:",default:true"`
}

type DStatus struct {
	bun.BaseModel `bun:"table:status"`
	Id            string `bun:",pk"`
	Host          string `bun:",pk"`
	AccountId     string
	Text          string `bun:"type:VARCHAR(10000)"`
	Url           string
	CreatedAt     time.Time
}

func ConvertCreatedAtToTokyo(statuses []Status) []Status {
	location, _ := time.LoadLocation("Asia/Tokyo")
	for i, v := range statuses {
		statuses[i].CreatedAt = v.CreatedAt.In(location)
	}
	return statuses
}

func dSelectAppByHost(host string) (App, error) {
	var app App
	err := bundb.NewSelect().Model(&app).Where("host = ?", host).Scan(ctx)
	if err != nil {
		if err == sql.ErrNoRows {
			return app, fmt.Errorf("no app for hostname: %s", host)
		}
		return app, fmt.Errorf("unknown db error: %v", err)
	}
	return app, nil
}

func dInsertApp(app App) error {
	_, err := bundb.NewInsert().Model(&app).Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to create app: %v", err)
	}
	return nil
}

func dInsertStatuses(statuses []Status, accountId string, host string) (int64, error) {
	if len(statuses) == 0 {
		return 0, nil
	}
	baseQuery := "INSERT INTO status (id, host, account_id, text, url, created_at) VALUES "
	var dataQueries []string
	vals := []interface{}{}
	for _, v := range statuses {
		dataQueries = append(dataQueries, "(?, ?, ?, ?, ?, ?)")
		vals = append(vals, v.Id, host, accountId, v.Text, v.Url, v.CreatedAt)
	}
	q := baseQuery + strings.Join(dataQueries, ",")
	res, err := db.Exec(q, vals...)
	if err != nil {
		return 0, fmt.Errorf("failed to insert statuses: %v", err)
	}
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("failed to get insert result: %v", err)
	}
	return rowsAffected, nil
}

func execSelectSingleStatusId(query string, accountId string) (string, error) {
	var id string
	row := db.QueryRow(query, accountId)
	if err := row.Scan(&id); err != nil {
		if err == sql.ErrNoRows {
			return "", nil
		}
		return "", fmt.Errorf("failed to select status: %v", err)
	}
	return id, nil
}

func dSelectNewestStatusIdByAccount(accoutId string) (string, error) {
	return execSelectSingleStatusId("SELECT id FROM status WHERE account_id = ? ORDER BY id DESC LIMIT 1", accoutId)
}

func dSelectOldestStatusIdByAccount(accoutId string) (string, error) {
	return execSelectSingleStatusId("SELECT id FROM status WHERE account_id = ? ORDER BY id ASC LIMIT 1", accoutId)
}

func dSelectStatusesByAccountAndText(accountId string, includedText string) ([]Status, error) {
	var res []Status

	rows, err := db.Query("SELECT id, text, url, created_at FROM status WHERE account_id = ? AND text LIKE CONCAT('%', ?, '%') ORDER BY id DESC", accountId, includedText)
	if err != nil {
		return nil, fmt.Errorf("query failed: %v", err)
	}
	defer rows.Close()
	for rows.Next() {
		var status Status
		if err := rows.Scan(&status.Id, &status.Text, &status.Url, &status.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan failed: %v", err)
		}
		res = append(res, status)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows included error: %v", err)
	}

	statuses := ConvertCreatedAtToTokyo(res)
	return statuses, nil
}

func dInsertAccountIfNotExists(id string, username string, host string) (int64, error) {
	res, err := db.Exec("INSERT INTO account SELECT * FROM (SELECT ? as c1, ? as c2, ? as c3, false) AS tmp WHERE NOT EXISTS (SELECT id FROM account WHERE id = ?) LIMIT 1", id, username, host, id)
	if err != nil {
		return 0, fmt.Errorf("failed to insert account: %v", err)
	}
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("failed to get insert result: %v", err)
	}
	return rowsAffected, nil
}

func dSelectAccountAllFetchedById(accountId string, host string) (bool, error) {
	var allFetched bool
	row := db.QueryRow("SELECT all_fetched FROM account WHERE id = ? AND host = ?", accountId, host)
	if err := row.Scan(&allFetched); err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, fmt.Errorf("dSelectAccountAllFetchedById: %v", err)
	}
	return allFetched, nil
}

func dUpdateAccountAllFetched(accountId string) error {
	_, err := db.Exec("UPDATE account set all_fetched = true where id = ?", accountId)
	if err != nil {
		return err
	}
	return nil
}

// get statuses from db by account acct joined with account table
func dSelectStatusesByAccount(username string, host string) ([]Status, error) {
	var res []Status

	rows, err := db.Query("SELECT status.text, status.created_at FROM status INNER JOIN account ON status.account_id = account.id WHERE account.username = ? AND account.host = ? ORDER BY status.id DESC", username, host)
	if err != nil {
		return nil, fmt.Errorf("query failed: %v", err)
	}
	defer rows.Close()
	for rows.Next() {
		var status Status
		if err := rows.Scan(&status.Text, &status.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan failed: %v", err)
		}
		res = append(res, status)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows included error: %v", err)
	}
	return ConvertCreatedAtToTokyo(res), nil
}
