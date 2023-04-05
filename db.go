package activitypublog

import (
	"database/sql"
	"fmt"
	"strings"
)

func dSelectAppByHost(host string) (App, error) {
	var app App

	row := db.QueryRow("SELECT * FROM app where host = ?", host)
	if err := row.Scan(&app.Host, &app.ClientId, &app.ClientSecret); err != nil {
		if err == sql.ErrNoRows {
			return app, fmt.Errorf("no app for hostname: %s", host)
		}
		return app, fmt.Errorf("unknown db error: %v", err)
	}
	return app, nil
}

func dInsertApp(app App) error {
	_, err := db.Exec("INSERT INTO app (host, client_id, client_secret) VALUES (?, ?, ?)", app.Host, app.ClientId, app.ClientSecret)
	if err != nil {
		return fmt.Errorf("failed to create app: %v", err)
	}
	return nil
}

func dInsertStatuses(statuses []Status, accountId string) (int64, error) {
	if len(statuses) == 0 {
		return 0, nil
	}
	baseQuery := "INSERT INTO status (id, accountId, text, url, created_at) VALUES "
	var dataQueries []string
	vals := []interface{}{}
	for _, v := range statuses {
		dataQueries = append(dataQueries, "(?, ?, ?, ?, ?)")
		vals = append(vals, v.Id, accountId, v.Text, v.Url, v.CreatedAt)
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
	return execSelectSingleStatusId("SELECT id FROM status WHERE accountId = ? ORDER BY id DESC LIMIT 1", accoutId)
}

func dSelectOldestStatusIdByAccount(accoutId string) (string, error) {
	return execSelectSingleStatusId("SELECT id FROM status WHERE accountId = ? ORDER BY id ASC LIMIT 1", accoutId)
}

func dSelectStatusesByAccountAndText(accountId string, includedText string) ([]Status, error) {
	var statuses []Status

	rows, err := db.Query("SELECT * FROM status WHERE accountId = ? AND text LIKE CONCAT('%', ?, '%') ORDER BY id DESC", accountId, includedText)
	if err != nil {
		return nil, fmt.Errorf("query failed: %v", err)
	}
	defer rows.Close()
	var discard string
	for rows.Next() {
		var status Status
		if err := rows.Scan(&status.Id, &discard, &status.Text, &status.Url, &status.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan failed: %v", err)
		}
		statuses = append(statuses, status)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows included error: %v", err)
	}
	return statuses, nil
}

func dInsertAccountIfNotExists(accountId string) (int64, error) {
	res, err := db.Exec("INSERT INTO account SELECT * FROM (SELECT ?, false) AS tmp WHERE NOT EXISTS (SELECT id FROM account WHERE id = ?) LIMIT 1", accountId, accountId)
	if err != nil {
		return 0, fmt.Errorf("failed to insert account: %v", err)
	}
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("failed to get insert result: %v", err)
	}
	return rowsAffected, nil
}

func dSelectAccountAllFetchedById(accountId string) (bool, error) {
	var allFetched bool
	row := db.QueryRow("SELECT all_fetched FROM account WHERE id = ?", accountId)
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
