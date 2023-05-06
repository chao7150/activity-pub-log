package activitypublog

import (
	"database/sql"
	"fmt"
	"time"
)

func ConvertCreatedAtToTokyo(statuses []Status) []Status {
	location, _ := time.LoadLocation("Asia/Tokyo")
	for i, v := range statuses {
		statuses[i].CreatedAt = v.CreatedAt.In(location)
	}
	return statuses
}

func ConvertCreatedAtToUTC(statuses []Status) []Status {
	for i, v := range statuses {
		statuses[i].CreatedAt = v.CreatedAt.UTC()
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
	statuses = ConvertCreatedAtToUTC(statuses)
	res, err := bundb.NewInsert().Model(&statuses).Exec(ctx)
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
	res, err := db.Exec("INSERT INTO account SELECT * FROM (SELECT ? as c1, ? as c2, ? as c3, false) AS tmp WHERE NOT EXISTS (SELECT id FROM account WHERE id = ?) LIMIT 1", id, host, username, id)
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
	var account Account
	err := bundb.NewSelect().Model(&account).Column("all_fetched").Where("id = ? AND host = ?", accountId, host).Scan(ctx)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, fmt.Errorf("dSelectAccountAllFetchedById: %v", err)
	}
	return account.AllFetched, nil
}

func dUpdateAccountAllFetched(accountId string) error {
	_, err := bundb.NewUpdate().Model(&Account{AllFetched: true}).Column("all_fetched").Where("id = ?", accountId).Exec(ctx)
	if err != nil {
		return err
	}
	return nil
}

func dSelectStatusesByAccount(username string, host string) ([]Status, error) {
	var statuses []Status

	err := bundb.NewSelect().
		Model(&statuses).
		Column("status.text", "status.created_at").
		Join("INNER JOIN account").
		JoinOn("status.account_id = account.id").
		Where("account.user_name = ? AND account.host = ?", username, host).
		Order("status.id DESC").
		Scan(ctx)
	if err != nil {
		return nil, fmt.Errorf("query failed: %v", err)
	}
	return ConvertCreatedAtToTokyo(statuses), nil
}
