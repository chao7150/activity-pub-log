package activitypublog

import "time"

type App struct {
	Host         string `json:"host"`
	ClientId     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
}

type Account struct {
	Id          string
	Acct        string
	Avatar      string
	DisplayName string `json:"display_name"`
	Url         string
}

type Tag struct {
	Name string
	Url  string
}

type Status struct {
	Account   Account
	Text      string
	Url       string
	CreatedAt time.Time
	Tags      []Tag
}
