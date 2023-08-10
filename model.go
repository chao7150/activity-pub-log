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
	bun.BaseModel `bun:"table:account"`
	Id            string `bun:",pk"`
	Host          string `bun:",pk"`
	Acct          string `bun:"-"`
	Avatar        string `bun:"-"`
	DisplayName   string `json:"display_name" bun:"-"`
	Url           string `bun:"-"`
	UserName      string
	AllFetched    bool `bun:",default:true"`
	Public        bool `bun:",default:false"`
	ShowUnlisted  bool
	ShowPrivate   bool
	ShowDirect    bool
}

type Tag struct {
	Name string
	Url  string
}

type Status struct {
	bun.BaseModel `bun:"table:status"`
	Id            string `bun:",pk"`
	Host          string `bun:",pk"`
	AccountId     string
	Account       Account `bun:"-"`
	Text          string  `bun:"type:VARCHAR(10000)"`
	Url           string
	CreatedAt     time.Time
	Tags          []Tag `bun:"-"`
	Visibility    string
}
