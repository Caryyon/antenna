package main

import (
	"embed"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/menu"
	"github.com/wailsapp/wails/v2/pkg/menu/keys"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	app := NewApp()

	// Create menu
	appMenu := menu.NewMenu()
	
	// App menu (macOS)
	appMenu.Append(menu.AppMenu())
	
	// File menu
	fileMenu := appMenu.AddSubmenu("File")
	fileMenu.AddText("Refresh", keys.CmdOrCtrl("r"), func(_ *menu.CallbackData) {
		runtime.EventsEmit(app.ctx, "refresh")
	})
	fileMenu.AddSeparator()
	fileMenu.AddText("Close Window", keys.CmdOrCtrl("w"), func(cd *menu.CallbackData) {
		runtime.Quit(cd.MenuItem.Ctx())
	})
	
	// Help menu
	helpMenu := appMenu.AddSubmenu("Help")
	helpMenu.AddText("GitHub Repository", nil, func(_ *menu.CallbackData) {
		runtime.BrowserOpenURL(app.ctx, "https://github.com/Caryyon/antenna")
	})
	helpMenu.AddText("Report Issue", nil, func(_ *menu.CallbackData) {
		runtime.BrowserOpenURL(app.ctx, "https://github.com/Caryyon/antenna/issues")
	})
	helpMenu.AddSeparator()
	helpMenu.AddText("OpenClaw Documentation", nil, func(_ *menu.CallbackData) {
		runtime.BrowserOpenURL(app.ctx, "https://docs.openclaw.ai")
	})

	err := wails.Run(&options.App{
		Title:     "Antenna",
		Width:     1200,
		Height:    700,
		MinWidth:  800,
		MinHeight: 500,
		Menu:      appMenu,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 5, G: 5, B: 5, A: 1},
		OnStartup:        app.startup,
		Bind: []interface{}{
			app,
		},
		Mac: &options.Mac{
			About: &options.AboutInfo{
				Title:   "Antenna",
				Message: "OpenClaw Session Monitor\n\nVersion 1.0.2\n\nhttps://github.com/Caryyon/antenna\n\n© 2026 Cary Wolff",
			},
		},
	})

	if err != nil {
		println("Error:", err.Error())
	}
}
