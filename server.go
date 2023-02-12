package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"

	"github.com/go-sql-driver/mysql"
	"github.com/labstack/echo/v4"
)

var db *sql.DB

type App struct {
	Host         string `json:"host"`
	ClientId     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
}

func getAppByHost(host string) (App, error) {
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

func createApp(app App) error {
	_, err := db.Exec("INSERT INTO app (host, client_id, client_secret) VALUES (?, ?, ?)", app.Host, app.ClientId, app.ClientSecret)
	if err != nil {
		return fmt.Errorf("createApp: %v", err)
	}
	return nil
}

func main() {
	cfg := mysql.Config{
		User:   os.Getenv("MYSQL_USER"),
		Passwd: os.Getenv("MYSQL_PASSWORD"),
		Net:    "tcp",
		Addr:   os.Getenv("MYSQL_HOST") + ":3306",
		DBName: os.Getenv("MYSQL_DATABASE"),
	}

	var err error
	db, err = sql.Open("mysql", cfg.FormatDSN())
	if err != nil {
		log.Fatal(err)
	}
	pingErr := db.Ping()
	if pingErr != nil {
		log.Fatal(pingErr)
	}
	fmt.Println("datebase connection established.")
	_, err = db.Exec("CREATE TABLE IF NOT EXISTS app (host VARCHAR(255), client_id VARCHAR(255), client_secret VARCHAR(255), PRIMARY KEY(`host`));")
	if err != nil {
		fmt.Printf("failed to initialize db table: %v", err)
	}

	e := echo.New()
	e.GET("/", func(c echo.Context) error {
		cookie, err := c.Cookie("user_session")
		if err != nil {
			return c.Redirect(302, "/login")
		}
		return c.String(http.StatusOK, fmt.Sprintf("name: %v\nvalue: %v", cookie.Name, cookie.Value))
	})
	e.File("/login", "static/login.html")
	e.POST("/sign_in", func(c echo.Context) error {
		host := c.FormValue("host")
		app, err := getAppByHost(host)
		if err != nil {
			fmt.Printf("app data was not found in db. fetch it.")
			path := "https://" + host + "/api/v1/apps"
			// TODO: redirect_uris
			resp, err := http.PostForm(path, url.Values{"client_name": {"chao-mastodon-log"}, "redirect_uris": {"example.com"}})
			if err != nil {
				return c.String(http.StatusBadRequest, fmt.Sprintf("failed to create app for the host: %v", err))
			}
			defer resp.Body.Close()
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				return c.String(http.StatusInternalServerError, fmt.Sprintf("failed to read response from server: %v", err))
			}
			if err := json.Unmarshal(body, &app); err != nil {
				return c.String(http.StatusInternalServerError, fmt.Sprintf("failed to parse response from server: %v", err))
			}
			app.Host = host
			err = createApp(app)
			if err != nil {
				return c.String(http.StatusInternalServerError, fmt.Sprintf("failed to insert app to db: %v", err))
			}
		}
		return c.String(http.StatusOK, app.ClientSecret)
	})

	e.Logger.Fatal(e.Start(":1323"))
}
