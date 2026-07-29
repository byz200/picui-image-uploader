package main

import (
	"context"
	"embed"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/windows"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

//go:embed all:frontend/dist
var assets embed.FS

//go:embed build/windows/icon.ico
var iconBytes []byte

func main() {
	app := NewApp()

	err := wails.Run(&options.App{
		Title:            "Picui 图床上传工具",
		Width:            1120,
		Height:           740,
		MinWidth:         900,
		MinHeight:        600,
		DisableResize:    false,
		Fullscreen:       false,
		WindowStartState: options.Normal,
		BackgroundColour: &options.RGBA{R: 255, G: 255, B: 255, A: 255},
		Assets:            assets,
		DragAndDrop: &options.DragAndDrop{
			EnableFileDrop:     true,
			DisableWebViewDrop: true,
			CSSDropProperty:    "--wails-drop-target",
			CSSDropValue:       "drop",
		},
		OnStartup: func(ctx context.Context) {
			app.Startup(ctx)
		},
		OnDomReady: func(ctx context.Context) {
			app.DomReady(ctx)
		},
		OnShutdown: func(ctx context.Context) {
			app.Shutdown(ctx)
		},
		OnBeforeClose: func(ctx context.Context) (prevent bool) {
			return app.BeforeClose(ctx)
		},
		Bind: []interface{}{
			app,
		},
		Windows: &windows.Options{
			WebviewIsTransparent: false,
			WindowIsTranslucent:  false,
		},
	})
	if err != nil {
		println("Error:", err.Error())
	}
}

// keep runtime import referenced（OnFileDrop 在 app.go 中使用）
var _ = runtime.OnFileDrop
