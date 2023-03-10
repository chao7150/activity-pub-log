package activitypublog

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

func hPostApp(host string, myHost string) (App, error) {
	var app App
	path := "https://" + host + "/api/v1/apps"
	resp, err := http.PostForm(path, url.Values{"client_name": {"chao-activitypublog"}, "redirect_uris": {"https://" + myHost + "/authorize"}})
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
		return account, fmt.Errorf("failed to parse account data: %v", err)
	}
	return account, nil
}

type hGetAccountStatusesResponse []struct {
	Id        string
	Account   Account
	Text      string
	Url       string
	CreatedAt string `json:"created_at"`
	Tags      []Tag
}

func hGetAccountStatusesNewerThan(host string, token string, id string, newestStatusId string) ([]Status, error) {
	var statuses []Status
	client := &http.Client{}
	req, err := http.NewRequest("GET", "https://"+host+"/api/v1/accounts/"+id+"/statuses?min_id="+newestStatusId, nil)
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
		return statuses, fmt.Errorf("failed to parse account data: %v", err)
	}

	location, _ := time.LoadLocation("Asia/Tokyo")
	for _, v := range res {
		ca, err := time.Parse(time.RFC3339, v.CreatedAt)
		if err != nil {
			continue
		}
		s := Status{
			Id:        v.Id,
			Account:   v.Account,
			Text:      v.Text,
			Url:       v.Url,
			CreatedAt: ca.In(location),
			Tags:      v.Tags,
		}
		statuses = append(statuses, s)
	}
	return statuses, nil
}

func hGetAccountStatusesOlderThan(host string, token string, id string, oldestStatusId string) ([]Status, error) {
	var statuses []Status
	client := &http.Client{}
	req, err := http.NewRequest("GET", "https://"+host+"/api/v1/accounts/"+id+"/statuses?max_id="+oldestStatusId, nil)
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
		return statuses, fmt.Errorf("failed to parse account data: %v", err)
	}

	location, _ := time.LoadLocation("Asia/Tokyo")
	for _, v := range res {
		ca, err := time.Parse(time.RFC3339, v.CreatedAt)
		if err != nil {
			continue
		}
		s := Status{
			Id:        v.Id,
			Account:   v.Account,
			Text:      v.Text,
			Url:       v.Url,
			CreatedAt: ca.In(location),
			Tags:      v.Tags,
		}
		statuses = append(statuses, s)
	}
	return statuses, nil
}
