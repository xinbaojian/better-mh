package main

import (
	"better-mh/adb"
	"better-mh/message"
	"context"
	"embed"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed all:frontend/dist
var assets embed.FS

//go:embed all:images
var images embed.FS

func main() {
	// Create an instance of the app structure
	app := NewApp()
	msg := &message.Message{}
	adbClient := &adb.Adb{}
	game := &Game{}

	// Create application with options
	err := wails.Run(&options.App{
		Title:  "better-mh",
		Width:  1024,
		Height: 768,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 27, G: 38, B: 54, A: 1},
		OnStartup: func(ctx context.Context) {
			app.startup(ctx)
			msg.Startup(ctx)
			adbClient.Startup(ctx, msg)
			game.Startup(ctx, adbClient)
		},
		Bind: []interface{}{
			app, adbClient, msg, game,
		},
	})

	if err != nil {
		println("Error:", err.Error())
	}
}
