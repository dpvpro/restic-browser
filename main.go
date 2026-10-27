package main

import (
	"embed"
	"flag"
	"log"

	"github.com/wailsapp/wails/v3/pkg/application"
	"github.com/wailsapp/wails/v3/pkg/events"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	// parse command line args
	resticPath := flag.String("restic", "", "Optional path to the restic binary")
	flag.Parse()

	// Create an instance of the restic app
	resticBrowser := NewResticBrowser(resticPath)

	// Create wails application with options and bind the app to the frontend
	app := application.New(application.Options{
		Name:        "Restic Browser",
		Description: "Restic Backup Browser GUI",
		Assets: application.AssetOptions{
			Handler: application.AssetFileServerFS(assets),
		},
		Mac: application.MacOptions{
			ApplicationShouldTerminateAfterLastWindowClosed: true,
		},
		Services: []application.Service{
			application.NewService(resticBrowser),
		},
	})

	window := app.Window.NewWithOptions(application.WebviewWindowOptions{
		Title:            "Restic Browser",
		Width:            1024,
		Height:           768,
		BackgroundColour: application.NewRGB(38, 41, 44),
		URL:              "/",
	})

	// warn about a missing restic program once the frontend is ready:
	// this is the first time we can show a dialog.
	window.OnWindowEvent(events.Common.WindowRuntimeReady, func(*application.WindowEvent) {
		resticBrowser.DomReady()
	})

	// Run the application. This blocks until the application has been exited.
	if err := app.Run(); err != nil {
		log.Fatal(err)
	}
}
