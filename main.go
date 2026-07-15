package main

import (
	"context"
	"embed"
	"os"

	"catscope/internal/update"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	handled, err := update.HandleCommandLine(os.Args[1:])
	if handled {
		if err != nil {
			println("CatScope update failed:", err.Error())
			os.Exit(1)
		}
		return
	}
	update.ScheduleCleanup(os.Args[1:])

	app := NewApp()

	err = wails.Run(&options.App{
		Title:  "CatScope",
		Width:  1280,
		Height: 820,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		OnStartup: app.startup,
		OnShutdown: func(ctx context.Context) {
			_ = app.StopLogcat()
		},
		Bind: []interface{}{
			app,
		},
	})
	if err != nil {
		println("Error:", err.Error())
	}
}
