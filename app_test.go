package main

import (
	"context"
	"testing"

	"flatstore/internal/appstream"
	"flatstore/internal/flathub"
)

func TestAppDiscoverApps(t *testing.T) {
	// Initialize our real structs (in a complex app you might use interfaces
	// here to mock them entirely, but since we can override the API base,
	// this works perfectly).
	flathubClient := flathub.NewClient()

	// Intentionally point to a bad URL to test error handling,
	// or point to an httptest server like in the flathub tests.
	flathubClient.SetAPIBase("http://127.0.0.1:0")

	catalogManager := appstream.NewManager()

	app := NewApp(flathubClient, catalogManager)

	// Simulate Wails startup
	app.startup(context.Background())

	// Test the bound method
	_, err := app.GetDiscoverApps()
	if err == nil {
		t.Error("Expected an error because we pointed to a dead URL, got nil")
	}
}
