package flathub

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"time"
)

// AppSummary is the normalised app record exposed to the frontend. Field names
// and JSON tags match the Wails-generated TypeScript bindings in
// frontend/wailsjs/go/models.ts.
type AppSummary struct {
	FlatpakAppId string `json:"flatpakAppId"`
	Name         string `json:"name"`
	Summary      string `json:"summary"`
	IconUrl      string `json:"iconUrl"`
	Version      string `json:"version"`
	Developer    string `json:"developer"`
}

type AppDetails struct {
	FlatpakAppId  string   `json:"flatpakAppId"`
	Name          string   `json:"name"`
	Summary       string   `json:"summary"`
	Description   string   `json:"description"`
	HomepageUrl   string   `json:"homepageUrl"`
	BugtrackerUrl string   `json:"bugtrackerUrl"`
	IconUrl       string   `json:"iconUrl"`
	Version       string   `json:"version"`
	Developer     string   `json:"developer"`
	Screenshots   []string `json:"screenshots"`
	ReleaseDate   string   `json:"releaseDate"`
}


// apiHit matches the raw JSON shape of a single record returned by the
// Flathub v2 API (field names differ from AppSummary).
type apiHit struct {
	AppId             string `json:"app_id"`
	Name              string `json:"name"`
	Summary           string `json:"summary"`
	Icon              string `json:"icon"`
	DeveloperName     string `json:"developer_name"`
	InstallsLastMonth int    `json:"installs_last_month"`
}

// apiResponse wraps the top-level {"hits":[...]} envelope used by all
// Flathub v2 collection and search endpoints.
type apiResponse struct {
	Hits []apiHit `json:"hits"`
}

func (h apiHit) toAppSummary() AppSummary {
	return AppSummary{
		FlatpakAppId: h.AppId,
		Name:         h.Name,
		Summary:      h.Summary,
		IconUrl:      h.Icon,
		Developer:    h.DeveloperName,
	}
}

func toAppSummaries(hits []apiHit) []AppSummary {
	apps := make([]AppSummary, len(hits))
	for i, h := range hits {
		apps[i] = h.toAppSummary()
	}
	return apps
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

// SetAPIBase overrides the default Flathub API base URL. This is primarily
// useful for testing purposes, allowing callers outside this package to
// point the client at a mock server or an invalid URL.
func (c *Client) SetAPIBase(base string) {
	c.apiBase = base
}

func (c *Client) FetchDiscoverApps() ([]AppSummary, error) {
	resp, err := c.httpClient.Get(fmt.Sprintf("%s/collection/recently-updated", c.apiBase))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result apiResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return toAppSummaries(result.Hits), nil
}

func (c *Client) FetchByCategory(category string) ([]AppSummary, error) {
	resp, err := c.httpClient.Get(fmt.Sprintf("%s/collection/category/%s", c.apiBase, strings.ToLower(category)))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result apiResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return toAppSummaries(result.Hits), nil
}

func (c *Client) Search(query string) ([]AppSummary, error) {
	body, err := json.Marshal(map[string]string{"query": query})
	if err != nil {
		return nil, err
	}
	resp, err := c.httpClient.Post(
		fmt.Sprintf("%s/search", c.apiBase),
		"application/json",
		bytes.NewReader(body),
	)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result apiResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return toAppSummaries(result.Hits), nil
}

// fetchCategoryHits retrieves raw hit records for a given category so that
// callers can sort or combine them before converting to AppSummary.
func (c *Client) fetchCategoryHits(category string) ([]apiHit, error) {
	resp, err := c.httpClient.Get(fmt.Sprintf("%s/collection/category/%s", c.apiBase, strings.ToLower(category)))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result apiResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return result.Hits, nil
}

// sortByPopularity sorts a slice of apiHit records in descending order of
// installs_last_month so the most popular apps appear first.
func sortByPopularity(hits []apiHit) {
	sort.Slice(hits, func(i, j int) bool {
		return hits[i].InstallsLastMonth > hits[j].InstallsLastMonth
	})
}

// FetchPopularApps returns the overall most-downloaded apps on Flathub over
// the last month, as reported by the /collection/popular endpoint.
func (c *Client) FetchPopularApps() ([]AppSummary, error) {
	resp, err := c.httpClient.Get(fmt.Sprintf("%s/collection/popular", c.apiBase))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result apiResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return toAppSummaries(result.Hits), nil
}

// FetchPopularGames returns the most-downloaded games, sorted by last-month
// install count.
func (c *Client) FetchPopularGames() ([]AppSummary, error) {
	hits, err := c.fetchCategoryHits("game")
	if err != nil {
		return nil, err
	}
	sortByPopularity(hits)
	return toAppSummaries(hits), nil
}

// FetchPopularCreate returns the most-downloaded creative apps (AudioVideo +
// Graphics), sorted by last-month install count with duplicates removed.
func (c *Client) FetchPopularCreate() ([]AppSummary, error) {
	audioHits, err := c.fetchCategoryHits("audiovideo")
	if err != nil {
		return nil, err
	}
	graphicsHits, err := c.fetchCategoryHits("graphics")
	if err != nil {
		return nil, err
	}

	// Merge and deduplicate by app_id.
	seen := make(map[string]struct{}, len(audioHits)+len(graphicsHits))
	merged := make([]apiHit, 0, len(audioHits)+len(graphicsHits))
	for _, h := range append(audioHits, graphicsHits...) {
		if _, ok := seen[h.AppId]; !ok {
			seen[h.AppId] = struct{}{}
			merged = append(merged, h)
		}
	}

	sortByPopularity(merged)
	return toAppSummaries(merged), nil
}
