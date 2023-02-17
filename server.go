package activitypublog

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"text/template"
	"time"

	"github.com/go-sql-driver/mysql"
	"github.com/labstack/echo/v4"
)

var db *sql.DB

type PostOauthTokenResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	Scope        string `json:"scope"`
	CreatedAt    int64  `json:"created_at"`
	RefreshToken string `json:"refresh_token"`
}

func StartServer() {
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

	t := &Template{
		templates: template.Must(template.ParseGlob("public/views/*.html")),
	}

	e := echo.New()
	e.Renderer = t
	e.Static("/static", "assets")
	e.GET("/", func(c echo.Context) error {
		tokenCookie, err := c.Cookie("token")
		if err != nil {
			return c.Redirect(302, "/login")
		}
		token := tokenCookie.Value
		hostCookie, err := c.Cookie("host")
		if err != nil {
			return c.Redirect(302, "/login")
		}
		host := hostCookie.Value
		account, nil := hGetVerifyCredentials(host, token)
		if err != nil {
			return c.String(http.StatusInternalServerError, fmt.Sprintf("error GET /: %v", err))
		}
		return c.Render(http.StatusOK, "top", account)
	})
	e.File("/login", "static/login.html")
	e.POST("/sign_in", func(c echo.Context) error {
		host := c.FormValue("host")
		app, err := dSelectAppByHost(host)
		if err != nil {
			fmt.Printf("app data was not found in db. fetch it.")
			app, err = hPostApp(host)
			if err != nil {
				return c.String(http.StatusInternalServerError, fmt.Sprintf("post app failed: %v", err))
			}
			err = dInsertApp(app)
			if err != nil {
				return c.String(http.StatusInternalServerError, fmt.Sprintf("failed to insert app to db: %v", err))
			}
		}
		u := url.URL{}
		u.Scheme = "https"
		u.Host = host
		u.Path = "/oauth/authorize"
		q := url.Values{"response_type": {"code"}, "client_id": {app.ClientId}, "redirect_uri": {"http://localhost:1323/authorize"}}
		u.RawQuery = q.Encode()
		cookie := &http.Cookie{
			Name:    "authentication-ongoing-instance-name",
			Value:   host,
			Expires: time.Now().Add(5 * time.Minute),
			Path:    "/authorize",
		}
		c.SetCookie(cookie)
		return c.Redirect(302, u.String())
	})
	e.GET("/authorize", func(c echo.Context) error {
		cookie, err := c.Cookie("authentication-ongoing-instance-name")
		if err != nil {
			return c.Redirect(302, "/")
		}
		host := cookie.Value
		code := c.QueryParam("code")
		u := url.URL{}
		u.Scheme = "https"
		u.Host = host
		u.Path = "/oauth/token"
		app, err := dSelectAppByHost(host)
		if err != nil {
			return c.String(http.StatusInternalServerError, fmt.Sprintf("cannot obtain app: %v", err))
		}
		q := url.Values{"grant_type": {"authorization_code"}, "code": {code}, "client_id": {app.ClientId}, "client_secret": {app.ClientSecret}, "redirect_uri": {"http://localhost:1323/authorize"}}
		resp, err := http.PostForm(u.String(), q)
		if err != nil {
			return c.String(http.StatusBadRequest, fmt.Sprintf("failed to create app for the host: %v", err))
		}
		defer resp.Body.Close()
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return c.String(http.StatusInternalServerError, fmt.Sprintf("failed to read response from server: %v", err))
		}
		var r PostOauthTokenResponse
		if err := json.Unmarshal(body, &r); err != nil {
			return c.String(http.StatusInternalServerError, fmt.Sprintf("failed to parse response from server: %v", err))
		}
		tokenCookie := &http.Cookie{
			Name:    "token",
			Value:   r.AccessToken,
			Expires: time.Now().Add(24 * 7 * time.Hour),
		}
		c.SetCookie(tokenCookie)
		hostCookie := &http.Cookie{
			Name:    "host",
			Value:   host,
			Expires: time.Now().Add(24 * 7 * time.Hour),
		}
		c.SetCookie(hostCookie)
		return c.Redirect(302, "/")
	})

	e.Logger.Fatal(e.Start(":1323"))
}
