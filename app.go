package main

import (
	"context"

	"flatstore/internal/appstream"
	"flatstore/internal/flathub"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type App struct {
	ctx           context.Context
	flathub       *flathub.Client
	catalog       *appstream.Manager
	systemManager *flathub.SystemManager
}

func NewApp(client *flathub.Client, catalog *appstream.Manager) *App {
	return &App{
		flathub: client,
		catalog: catalog,
		// SystemManager will be fully bound once startup injection provides context
	}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx

	// Bind the native Wails runtime event system to our Manager package securely
	eventEmitter := func(eventName string, optionalData ...interface{}) {
		runtime.EventsEmit(ctx, eventName, optionalData...)
	}
	a.systemManager = flathub.NewSystemManager(eventEmitter)

	go func() {
		if err := a.catalog.Sync(); err != nil {
			runtime.EventsEmit(ctx, "catalog:error", err.Error())
		} else {
			runtime.EventsEmit(ctx, "catalog:ready", true)
		}
	}()
}

// --- Data Fetching Methods (Bound to Frontend) ---

func (a *App) GetDiscoverApps() ([]flathub.AppSummary, error) {
	return a.flathub.FetchDiscoverApps()
}

func (a *App) GetAppsByCategory(category string) ([]flathub.AppSummary, error) {
	return a.flathub.FetchByCategory(category)
}

func (a *App) SearchApps(query string) ([]flathub.AppSummary, error) {
	return a.flathub.Search(query)
}

// InstallApp handles async tasks natively now
func (a *App) InstallApp(appID string, installAsSystem bool) error {
	args := []string{"install", "flathub", appID}
	if installAsSystem {
		args = append(args, "--system")
	} else {
		args = append(args, "--user")
	}

	// Run inside a background goroutine so the bound Wails frontend thread
	// does not experience UI lockups/freezes during large downloads
	go func() {
		err := a.systemManager.ExecuteWithProgress(a.ctx, appID, args...)
		if err != nil {
			println("Async installation failed for", appID, ":", err.Error())
		}
	}()

	return nil
}
