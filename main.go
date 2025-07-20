package main

import (
	"better-mh/adb"
	"better-mh/logs"
	"better-mh/message"
	"context"
	"embed"
	"github.com/wailsapp/wails/v2/pkg/options/mac"

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
	log := &logs.Log{}

	// Create application with options
	err := wails.Run(&options.App{
		Title:  "BetterMh",
		Width:  1024,
		Height: 768,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 27, G: 38, B: 54, A: 1},
		OnStartup: func(ctx context.Context) {
			log.Startup(ctx)
			app.startup(ctx)
			msg.Startup(ctx)
			adbClient.Startup(ctx, msg, log)
			game.Startup(ctx, adbClient, log)
		},
		Bind: []interface{}{
			app, adbClient, msg, game,
		},
		Mac: &mac.Options{
			TitleBar: &mac.TitleBar{
				// 使标题栏透明。 这具有隐藏标题栏和内容填充窗口的效果。
				TitlebarAppearsTransparent: false,
				// 隐藏窗口的标题。
				HideTitle: true,
				//从 style mask 中删除 NSWindowStyleMaskTitled
				HideTitleBar: false,
				// 使 webview 填满整个窗口。
				FullSizeContent: false,
				// 向窗口添加默认工具栏。
				UseToolbar: false,
				// 删除工具栏下方的线条。
				HideToolbarSeparator: false,
				//OnFileOpen:                 app.onFileOpen,
				//OnUrlOpen:                  app.onUrlOpen,
			},
			Appearance: mac.NSAppearanceNameDarkAqua,
			// WebView 透明
			WebviewIsTransparent: false,
			// 窗口半透明
			WindowIsTranslucent: true,
			//About: &mac.AboutInfo{
			//	Title:   "My Application",
			//	Message: "© 2021 Me",
			//	Icon:    icon,
			//},
		},
	})

	if err != nil {
		println("Error:", err.Error())
	}
}
