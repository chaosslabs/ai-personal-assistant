package main

import (
	"context"
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

var globalApp *App

func createAppMenu() *menu.Menu {
	appMenu := menu.NewMenu()

	// AI Personal Assistant menu
	aiMenu := appMenu.AddSubmenu("AI Personal Assistant")
	aiMenu.AddText("About AI Personal Assistant", keys.CmdOrCtrl(""), nil)
	aiMenu.AddSeparator()
	aiMenu.AddText("Hide AI Personal Assistant", keys.CmdOrCtrl("h"), func(_ *menu.CallbackData) {
		if globalApp != nil {
			globalApp.HideWindow()
		}
	})
	aiMenu.AddText("Hide Others", keys.CmdOrCtrl("alt+h"), nil)
	aiMenu.AddText("Show All", keys.CmdOrCtrl(""), nil)
	aiMenu.AddSeparator()
	aiMenu.AddText("Quit AI Personal Assistant", keys.CmdOrCtrl("q"), func(_ *menu.CallbackData) {
		if globalApp != nil {
			globalApp.QuitApp()
		}
	})

	// Window menu
	windowMenu := appMenu.AddSubmenu("Window")
	windowMenu.AddText("Show Window", keys.CmdOrCtrl("0"), func(_ *menu.CallbackData) {
		if globalApp != nil {
			globalApp.ShowWindow()
		}
	})
	windowMenu.AddSeparator()
	windowMenu.AddText("App Status", keys.CmdOrCtrl("s"), func(_ *menu.CallbackData) {
		if globalApp != nil {
			result := globalApp.GetAppStatus()
			if globalApp.ctx != nil {
				go func() {
					runtime.MessageDialog(globalApp.ctx, runtime.MessageDialogOptions{
						Type:    runtime.InfoDialog,
						Title:   "App Status",
						Message: result,
					})
				}()
			}
		}
	})

	return appMenu
}

func main() {
	// Create an instance of the app structure
	app := NewApp()
	globalApp = app

	// Create app menu
	appMenu := createAppMenu()

	// Create application with options
	err := wails.Run(&options.App{
		Title:  "AI Personal Assistant",
		Width:  800,
		Height: 600,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 27, G: 38, B: 54, A: 1},
		OnStartup: func(ctx context.Context) {
			app.startup(ctx)
		},
		Bind: []interface{}{
			app,
		},
		Menu:              appMenu,
		MinWidth:          400,
		MinHeight:         300,
		HideWindowOnClose: false,
		StartHidden:       false,
		WindowStartState:  options.Normal,
	})

	if err != nil {
		println("Error:", err.Error())
	}
}
