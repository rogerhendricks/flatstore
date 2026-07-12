package main

import (
	"embed"

	"flatstore/internal/appstream"
	"flatstore/internal/flathub"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

var assets embed.FS

func main() {
	// Initialize core services
	flathubClient := flathub.NewClient()
	catalogManager := appstream.NewManager()
	app := NewApp(flathubClient, catalogManager)

	err := wails.Run(&options.App{
		Title:     "FlatStore",
		Width:     1024,
		Height:    768,
		MinWidth:  960,
		MinHeight: 600,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 255, G: 255, B: 255, A: 1},
		OnStartup:        app.startup,
		Bind: []interface{}{
			app,
		},
	})
	if err != nil {
		println("Error:", err.Error())
	}
}
