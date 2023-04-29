package activitypublog

import (
	"time"

	"github.com/uptrace/bun"
)

type App struct {
	bun.BaseModel `bun:"table:app"`
	Host          string `json:"host" bun:",pk"`
	ClientId      string `json:"client_id"`
	ClientSecret  string `json:"client_secret"`
}

type Account struct {
	Id          string
	Acct        string
	Avatar      string
	DisplayName string `json:"display_name"`
	Url         string
	UserName    string
}

type Tag struct {
	Name string
	Url  string
}

type Status struct {
	Id        string
	Account   Account
	Text      string
	Url       string
	CreatedAt time.Time
	Tags      []Tag
}
