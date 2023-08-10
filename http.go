package activitypublog

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

func hPostApp(host string, baseUrl string) (App, error) {
	var app App
	path := "https://" + host + "/api/v1/apps"
	resp, err := http.PostForm(path, url.Values{"client_name": {"chao-activitypublog"}, "redirect_uris": {baseUrl + "/authorize"}})
	if err != nil {
		return app, fmt.Errorf("failed to create app for the host: %v", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return app, fmt.Errorf("failed to read response from server: %v", err)
	}

	if err := json.Unmarshal(body, &app); err != nil {
		return app, fmt.Errorf("failed to parse response from server: %v", err)
	}
	app.Host = host
	return app, nil
}

func hGetVerifyCredentials(host string, token string) (Account, error) {
	var account Account
	client := &http.Client{}
	req, err := http.NewRequest("GET", "https://"+host+"/api/v1/accounts/verify_credentials", nil)
	if err != nil {
		return account, fmt.Errorf("failed to create request: %v", err)
	}
	req.Header.Add("Authorization", "Bearer "+token)
	resp, err := client.Do(req)
	if err != nil {
		return account, fmt.Errorf("failed to GET verify_credentials: %v", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return account, fmt.Errorf("failed to read response body: %v", err)
	}
	if err := json.Unmarshal(body, &account); err != nil {
		fmt.Printf("failed to parse account data: raw data: %v", body)
		return account, fmt.Errorf("failed to parse account data: %v", err)
	}
	return account, nil
}

type hGetAccountStatusesResponse []struct {
	Id         string
	Account    Account
	Text       string
	Url        string
	CreatedAt  string `json:"created_at"`
	Tags       []Tag
	Visibility string
}

func hGetAccountStatuses(host string, token string, id string, params string) ([]Status, error) {
	var statuses []Status
	client := &http.Client{}
	req, err := http.NewRequest("GET", "https://"+host+"/api/v1/accounts/"+id+"/statuses?"+params, nil)
	if err != nil {
		return statuses, fmt.Errorf("failed to create request: %v", err)
	}
	req.Header.Add("Authorization", "Bearer "+token)
	resp, err := client.Do(req)
	if err != nil {
		return statuses, fmt.Errorf("failed to GET accounts/:id/statuses: %v", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return statuses, fmt.Errorf("failed to read response body: %v", err)
	}
	var res hGetAccountStatusesResponse
	if err := json.Unmarshal(body, &res); err != nil {
		fmt.Printf("failed to parse account data: raw data: %v", body)
		return statuses, fmt.Errorf("failed to parse account data: %v", err)
	}

	location, _ := time.LoadLocation("Asia/Tokyo")
	for _, v := range res {
		ca, err := time.Parse(time.RFC3339, v.CreatedAt)
		if err != nil {
			continue
		}
		s := Status{
			Id:         v.Id,
			Account:    v.Account,
			Text:       v.Text,
			Url:        v.Url,
			CreatedAt:  ca.In(location),
			Tags:       v.Tags,
			Host:       host,
			AccountId:  id,
			Visibility: v.Visibility,
		}
		statuses = append(statuses, s)
	}
	return statuses, nil
}

func hGetAccountStatusesOlderThan(host string, token string, id string, maxId string) ([]Status, error) {
	return hGetAccountStatuses(host, token, id, "max_id="+maxId)
}

func hGetAccountStatusesAll(host string, token string, id string, minId string, maxId string) ([]Status, error) {
	var statuses []Status
	for {
		s, err := hGetAccountStatuses(host, token, id, "max_id="+maxId+"&min_id="+minId)
		if err != nil {
			return statuses, err
		}
		if len(s) == 0 {
			break
		}
		statuses = append(statuses, s...)
		maxId = s[len(s)-1].Id
		time.Sleep(time.Second * 2)
	}
	return statuses, nil
}
