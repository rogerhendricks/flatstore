package flathub

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestFetchDiscoverApps(t *testing.T) {
	// 1. Create mock data
	mockApps := []AppSummary{
		{
			FlatpakAppId: "org.gimp.GIMP",
			Name:         "GIMP",
			Summary:      "GNU Image Manipulation Program",
		},
	}

	// 2. Set up a local mock HTTP server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/collection/recently-updated" {
			t.Errorf("Expected path /collection/recently-updated, got %s", r.URL.Path)
		}
		json.NewEncoder(w).Encode(mockApps)
	}))
	defer server.Close()

	// 3. Initialize the client and point it to the mock server
	client := NewClient()
	client.apiBase = server.URL // Override the live URL

	// 4. Execute the method
	apps, err := client.FetchDiscoverApps()
	// 5. Assert the results
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(apps) != 1 {
		t.Fatalf("Expected 1 app, got %d", len(apps))
	}

	if apps[0].Name != "GIMP" {
		t.Errorf("Expected app name 'GIMP', got '%s'", apps[0].Name)
	}
}
