package main

import (
	"bytes"
	"encoding/base64"
	"errors"
	"image"
	"image/png"
	"time"

	"github.com/kbinani/screenshot"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

var errNoDisplay = errors.New("未检测到可用显示器")

// ScreenCapture 单显示器截图数据。
type ScreenCapture struct {
	Index   int    `json:"index"`
	X       int    `json:"x"`
	Y       int    `json:"y"`
	Width   int    `json:"width"`
	Height  int    `json:"height"`
	DataURL string `json:"dataUrl"`
}

// ScreenshotPayload 截图选区所需数据。
type ScreenshotPayload struct {
	MinX        int             `json:"minX"`
	MinY        int             `json:"minY"`
	TotalWidth  int             `json:"totalWidth"`
	TotalHeight int             `json:"totalHeight"`
	Monitors    []ScreenCapture `json:"monitors"`
}

type ScreenshotManager struct {
	app *App
}

func NewScreenshotManager(app *App) *ScreenshotManager {
	return &ScreenshotManager{app: app}
}

// CaptureAll 捕获所有显示器并向前端发送截图数据，由前端展示选区遮罩。
// 捕获前隐藏主窗口，避免遮挡截图内容；捕获后重新显示窗口以承载选区遮罩。
func (m *ScreenshotManager) CaptureAll() error {
	// 隐藏主窗口，等待其完全消失后再截图
	if m.app.ctx != nil {
		runtime.WindowHide(m.app.ctx)
		time.Sleep(250 * time.Millisecond)
	}

	n := screenshot.NumActiveDisplays()
	if n == 0 {
		if m.app.ctx != nil {
			runtime.WindowShow(m.app.ctx)
		}
		return errNoDisplay
	}

	var union image.Rectangle
	monitors := make([]ScreenCapture, 0, n)
	for i := 0; i < n; i++ {
		bounds := screenshot.GetDisplayBounds(i)
		img, err := screenshot.CaptureRect(bounds)
		if err != nil {
			continue
		}
		var buf bytes.Buffer
		if err := png.Encode(&buf, img); err != nil {
			continue
		}
		dataURL := "data:image/png;base64," + base64.StdEncoding.EncodeToString(buf.Bytes())
		monitors = append(monitors, ScreenCapture{
			Index:   i,
			X:       bounds.Min.X,
			Y:       bounds.Min.Y,
			Width:   bounds.Dx(),
			Height:  bounds.Dy(),
			DataURL: dataURL,
		})
		if len(monitors) == 1 {
			union = bounds
		} else {
			union = union.Union(bounds)
		}
	}
	if len(monitors) == 0 {
		if m.app.ctx != nil {
			runtime.WindowShow(m.app.ctx)
		}
		return errNoDisplay
	}

	payload := ScreenshotPayload{
		MinX:        union.Min.X,
		MinY:        union.Min.Y,
		TotalWidth:  union.Dx(),
		TotalHeight: union.Dy(),
		Monitors:    monitors,
	}
	// 先重新显示窗口（承载选区遮罩），再下发截图数据
	if m.app.ctx != nil {
		runtime.WindowShow(m.app.ctx)
	}
	m.app.emit("screenshot:ready", payload)
	return nil
}
