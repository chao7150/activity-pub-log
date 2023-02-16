package activitypublog

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

func hPostApp(host string) (App, error) {
	var app App
	path := "https://" + host + "/api/v1/apps"
	resp, err := http.PostForm(path, url.Values{"client_name": {"chao-activitypublog"}, "redirect_uris": {"http://localhost:1323/authorize"}})
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
