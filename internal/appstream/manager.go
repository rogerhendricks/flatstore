package appstream

import (
	"compress/gzip"
	"encoding/xml"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"
)

const appstreamURL = "https://dl.flathub.org/repo/appstream/x86_64/appstream.xml.gz"

type Manager struct {
	mu         sync.RWMutex
	appsByID   map[string]*Component
	appsByCat  map[string][]*Component
	httpClient *http.Client
	isLoaded   bool
}

func NewManager() *Manager {
	return &Manager{
		appsByID:  make(map[string]*Component),
		appsByCat: make(map[string][]*Component),
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// Sync downloads and parses the latest AppStream catalog
func (m *Manager) Sync() error {
	log.Println("Downloading AppStream catalog...")

	req, err := http.NewRequest(http.MethodGet, appstreamURL, nil)
	if err != nil {
		return err
	}

	resp, err := m.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to download catalog: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	// Decompress the gzip stream on the fly
	gzReader, err := gzip.NewReader(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to create gzip reader: %w", err)
	}
	defer gzReader.Close()

	log.Println("Parsing AppStream XML...")
	var catalog Catalog
	decoder := xml.NewDecoder(gzReader)

	if err := decoder.Decode(&catalog); err != nil {
		return fmt.Errorf("failed to parse xml: %w", err)
	}

	m.buildIndexes(&catalog)
	m.isLoaded = true
	log.Printf("Successfully loaded %d applications into memory", len(m.appsByID))

	return nil
}

// buildIndexes populates our fast-lookup maps
func (m *Manager) buildIndexes(catalog *Catalog) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Clear existing data for fresh sync
	m.appsByID = make(map[string]*Component)
	m.appsByCat = make(map[string][]*Component)

	for i := range catalog.Components {
		comp := &catalog.Components[i]

		// We only want desktop applications, not runtimes or localized strings
		if comp.Type != "desktop-application" {
			continue
		}

		m.appsByID[comp.ID] = comp

		for _, cat := range comp.Categories.List {
			m.appsByCat[cat] = append(m.appsByCat[cat], comp)
		}
	}
}

// GetApp returns a single app by its Flatpak ID
func (m *Manager) GetApp(id string) (*Component, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if app, exists := m.appsByID[id]; exists {
		return app, true
	}

	// Try lookup fallback variations (with or without .desktop suffix)
	var altID string
	if strings.HasSuffix(id, ".desktop") {
		altID = strings.TrimSuffix(id, ".desktop")
	} else {
		altID = id + ".desktop"
	}

	app, exists := m.appsByID[altID]
	return app, exists
}

// GetAppsByCategory returns all apps matching a category
func (m *Manager) GetAppsByCategory(category string) []*Component {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Return a copy of the slice to prevent external mutation
	apps := m.appsByCat[category]
	result := make([]*Component, len(apps))
	copy(result, apps)

	return result
}

// IsReady checks if the catalog has been downloaded and parsed
func (m *Manager) IsReady() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.isLoaded
}
