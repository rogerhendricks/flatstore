package main

import (
	"context"
	"fmt"

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

// GetInstalledApps returns the system's installed Flatpaks
func (a *App) GetInstalledApps() ([]flathub.InstalledApp, error) {
	return a.systemManager.ListInstalledApps(a.ctx)
}

func (a *App) UninstallApp(appID string, asSystem bool) error {
	args := []string{"uninstall", appID}
	if asSystem {
		args = append(args, "--system")
	} else {
		args = append(args, "--user")
	}
	go func() {
		err := a.systemManager.ExecuteWithProgress(a.ctx, appID, args...)
		if err != nil {
			println("Async uninstall failed for", appID, ":", err.Error())
		}
	}()
	return nil
}

func (a *App) UpdateApp(appID string, asSystem bool) error {
	args := []string{"update", appID}
	if asSystem {
		args = append(args, "--system")
	} else {
		args = append(args, "--user")
	}
	go func() {
		err := a.systemManager.ExecuteWithProgress(a.ctx, appID, args...)
		if err != nil {
			println("Async update failed for", appID, ":", err.Error())
		}
	}()
	return nil
}

// --- Popular / Discovery Methods ---

func (a *App) GetPopularApps() ([]flathub.AppSummary, error) {
	return a.flathub.FetchPopularApps()
}

func (a *App) GetPopularGames() ([]flathub.AppSummary, error) {
	return a.flathub.FetchPopularGames()
}

func (a *App) GetPopularCreate() ([]flathub.AppSummary, error) {
	return a.flathub.FetchPopularCreate()
}

func (a *App) GetAppDetails(appID string) (*flathub.AppDetails, error) {
	comp, exists := a.catalog.GetApp(appID)
	if !exists {
		return nil, fmt.Errorf("application %s not found in local catalog", appID)
	}

	var homepageURL string
	var bugtrackerURL string
	for _, u := range comp.URLs {
		if u.Type == "homepage" {
			homepageURL = u.Value
		} else if u.Type == "bugtracker" {
			bugtrackerURL = u.Value
		}
	}

	var iconURL string
	for _, icon := range comp.Icons {
		if icon.Type == "remote" {
			iconURL = icon.Value
			break
		}
	}
	if iconURL == "" {
		iconURL = "https://dl.flathub.org/assets/default/settings.svg"
	}

	var screenshots []string
	for _, s := range comp.Screenshots {
		for _, img := range s.Images {
			if img.Type == "source" && img.Value != "" {
				screenshots = append(screenshots, img.Value)
			}
		}
	}

	var version string
	var releaseDate string
	if len(comp.Releases.List) > 0 {
		version = comp.Releases.List[0].Version
		releaseDate = comp.Releases.List[0].Date
	}

	details := &flathub.AppDetails{
		FlatpakAppId:  comp.ID,
		Name:          comp.Name.Value,
		Summary:       comp.Summary.Value,
		Description:   comp.Description.Raw,
		HomepageUrl:   homepageURL,
		BugtrackerUrl: bugtrackerURL,
		IconUrl:       iconURL,
		Version:       version,
		Developer:     comp.Developer,
		Screenshots:   screenshots,
		ReleaseDate:   releaseDate,
		AgeRating:     comp.ContentRating.GetAgeRating(),
		License:       comp.ProjectLicense,
	}

	return details, nil
}

func (a *App) OpenApp(appID string) error {
	return a.systemManager.RunApp(appID)
}

// IsCatalogReady reports whether the AppStream catalog has finished syncing.
// Useful for the frontend to check state on mount, in case it missed the
// "catalog:ready"/"catalog:error" events emitted during startup().
func (a *App) IsCatalogReady() bool {
	return a.catalog.IsReady()
}
