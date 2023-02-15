package activitypublog

import (
	"database/sql"
	"fmt"
)

func dSelectAppByHost(host string) (App, error) {
	var app App

	row := db.QueryRow("SELECT * FROM app where host = ?", host)
	if err := row.Scan(&app.Host, &app.ClientId, &app.ClientSecret); err != nil {
		if err == sql.ErrNoRows {
			return app, fmt.Errorf("getAppByHost %s: no such app", host)
		}
		return app, fmt.Errorf("getAppByHost %s: %v", host, err)
	}
	return app, nil
}

func dInsertApp(app App) error {
	_, err := db.Exec("INSERT INTO app (host, client_id, client_secret) VALUES (?, ?, ?)", app.Host, app.ClientId, app.ClientSecret)
	if err != nil {
		return fmt.Errorf("createApp: %v", err)
	}
	return nil
}
