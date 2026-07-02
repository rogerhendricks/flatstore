package flathub

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type AppSummary struct {
	FlatpakAppId string `json:"flatpakAppId"`
	Name         string `json:"name"`
	Summary      string `json:"summary"`
	IconUrl      string `json:"iconUrl"`
	Version      string `json:"version"`
	Developer    string `json:"developer"`
}

type Client struct {
	httpClient *http.Client
	apiBase    string
}

func NewClient() *Client {
	return &Client{
		httpClient: &http.Client{Timeout: 10 * time.Second},
		apiBase:    "https://flathub.org/api/v2",
	}
}

func (c *Client) FetchDiscoverApps() ([]AppSummary, error) {
	// Pulls the recently updated apps from flathub endpoint
	resp, err := c.httpClient.Get(fmt.Sprintf("%s/collection/recently-updated", c.apiBase))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var apps []AppSummary
	if err := json.NewDecoder(resp.Body).Decode(&apps); err != nil {
		return nil, err
	}
	return apps, nil
}

func (c *Client) FetchByCategory(category string) ([]AppSummary, error) {
	resp, err := c.httpClient.Get(fmt.Sprintf("%s/category/%s", c.apiBase, category))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var apps []AppSummary
	if err := json.NewDecoder(resp.Body).Decode(&apps); err != nil {
		return nil, err
	}
	return apps, nil
}

func (c *Client) Search(query string) ([]AppSummary, error) {
	resp, err := c.httpClient.Get(fmt.Sprintf("%s/search?q=%s", c.apiBase, query))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var apps []AppSummary
	if err := json.NewDecoder(resp.Body).Decode(&apps); err != nil {
		return nil, err
	}
	return apps, nil
}
