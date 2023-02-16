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
		client := &http.Client{}
		req, err := http.NewRequest("GET", "https://"+host+"/api/v1/accounts/verify_credentials", nil)
		if err != nil {
			return c.String(http.StatusInternalServerError, fmt.Sprintf("failed to verify credentials: %v", err))
		}
		req.Header.Add("Authorization", "Bearer "+token)
		resp, err := client.Do(req)
		if err != nil {
			return c.String(http.StatusBadRequest, fmt.Sprintf("failed to fetch user infomation: %v", err))
		}
		defer resp.Body.Close()
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return c.String(http.StatusInternalServerError, fmt.Sprintf("failed to read response from server: %v", err))
		}
		return c.Render(http.StatusOK, "top", string(body))
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
		cookie := new(http.Cookie)
		cookie.Name = "authentication-ongoing-instance-name"
		cookie.Value = host
		cookie.Expires = time.Now().Add(5 * time.Minute)
		cookie.Path = "/authorize"
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
		fmt.Println(q)
		resp, err := http.PostForm(u.String(), q)
		if err != nil {
			return c.String(http.StatusBadRequest, fmt.Sprintf("failed to create app for the host: %v", err))
		}
		defer resp.Body.Close()
		body, err := io.ReadAll(resp.Body)
		fmt.Printf("body %s", body)
		if err != nil {
			return c.String(http.StatusInternalServerError, fmt.Sprintf("failed to read response from server: %v", err))
		}
		var r PostOauthTokenResponse
		if err := json.Unmarshal(body, &r); err != nil {
			return c.String(http.StatusInternalServerError, fmt.Sprintf("failed to parse response from server: %v", err))
		}
		fmt.Printf("token %v", r)
		cookie = new(http.Cookie)
		cookie.Name = "token"
		cookie.Value = r.AccessToken
		cookie.Expires = time.Now().Add(24 * 7 * time.Hour)
		c.SetCookie(cookie)
		hostCookie := http.Cookie{}
		hostCookie.Name = "host"
		hostCookie.Value = host
		hostCookie.Expires = time.Now().Add(24 * 7 * time.Hour)
		c.SetCookie(&hostCookie)
		return c.Redirect(302, "/")
	})

	e.Logger.Fatal(e.Start(":1323"))
}
