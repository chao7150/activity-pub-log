package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/go-sql-driver/mysql"
	"github.com/labstack/echo/v4"
)

var db *sql.DB

func main() {
	cfg := mysql.Config{
		User:   os.Getenv("MYSQL_USER"),
		Passwd: os.Getenv("MYSQL_PASSWORD"),
		Net:    "tcp",
		Addr:   os.Getenv("MYSQL_HOST") + ":3306",
		DBName: os.Getenv("MYSQL_DATABASE"),
	}
	fmt.Println(cfg.FormatDSN())
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
		fmt.Println(host)
		return c.String(http.StatusOK, host)
	})

	e.Logger.Fatal(e.Start(":1323"))
}
