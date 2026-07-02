package appstream

import (
	"testing"
)

// We can test the internal indexing logic without needing an HTTP request
// by directly feeding a dummy Catalog struct into buildIndexes.
func TestManagerIndexing(t *testing.T) {
	// 1. Setup mock catalog data
	mockCatalog := &Catalog{
		Components: []Component{
			{
				Type: "desktop-application",
				ID:   "org.mozilla.firefox",
				Name: "Firefox",
				Categories: Categories{
					List: []string{"Network", "WebBrowser"},
				},
			},
			{
				Type: "desktop-application",
				ID:   "com.visualstudio.code",
				Name: "VS Code",
				Categories: Categories{
					List: []string{"Development"},
				},
			},
			{
				Type: "runtime", // Should be ignored by our indexer
				ID:   "org.freedesktop.Platform",
				Name: "Freedesktop Platform",
			},
		},
	}

	manager := NewManager()

	// 2. Build the indexes manually for testing
	manager.buildIndexes(mockCatalog)

	// 3. Test GetApp (ByID)
	t.Run("GetApp", func(t *testing.T) {
		app, exists := manager.GetApp("org.mozilla.firefox")
		if !exists {
			t.Fatalf("Expected Firefox to exist in index")
		}
		if app.Name != "Firefox" {
			t.Errorf("Expected name Firefox, got %s", app.Name)
		}

		_, exists = manager.GetApp("org.freedesktop.Platform")
		if exists {
			t.Errorf("Expected runtime to be ignored and not exist in index")
		}
	})

	// 4. Test GetAppsByCategory
	t.Run("GetAppsByCategory", func(t *testing.T) {
		devApps := manager.GetAppsByCategory("Development")
		if len(devApps) != 1 {
			t.Fatalf("Expected 1 dev app, got %d", len(devApps))
		}
		if devApps[0].Name != "VS Code" {
			t.Errorf("Expected VS Code, got %s", devApps[0].Name)
		}

		networkApps := manager.GetAppsByCategory("Network")
		if len(networkApps) != 1 {
			t.Fatalf("Expected 1 network app, got %d", len(networkApps))
		}
	})
}
