package main

import (
	"runtime"
	"sync"

	"fyne.io/systray"
)

// TrayManager 系统托盘常驻管理。
type TrayManager struct {
	app     *App
	mu      sync.Mutex
	started bool
}

func NewTrayManager(app *App) *TrayManager {
	return &TrayManager{app: app}
}

func (t *TrayManager) Start() {
	t.mu.Lock()
	if t.started {
		t.mu.Unlock()
		return
	}
	t.started = true
	t.mu.Unlock()
	go t.run()
}

// run 在独立 OS 线程上运行托盘消息循环（Windows 允许任意线程拥有消息循环）。
func (t *TrayManager) run() {
	runtime.LockOSThread()
	systray.Run(t.onReady, t.onExit)
}

func (t *TrayManager) onReady() {
	systray.SetIcon(iconBytes)
	systray.SetTitle("")
	systray.SetTooltip("Picui 图床上传工具")

	mOpen := systray.AddMenuItem("打开主窗口", "显示主窗口")
	mClip := systray.AddMenuItem("剪贴板快速上传", "上传剪贴板中的图片")
	mShot := systray.AddMenuItem("截图上传", "屏幕选区截图并上传")
	systray.AddSeparator()
	mQuit := systray.AddMenuItem("退出程序", "退出 Picui 上传工具")

	go func() {
		for {
			select {
			case <-mOpen.ClickedCh:
				t.app.BringToFront()
			case <-mClip.ClickedCh:
				_, _ = t.app.queue.EnqueueClipboard(t.app.defaultUploadOpts())
			case <-mShot.ClickedCh:
				t.app.BringToFront()
				_ = t.app.ss.CaptureAll()
			case <-mQuit.ClickedCh:
				t.app.QuitApp()
				return
			}
		}
	}()
}

func (t *TrayManager) onExit() {
	// 清理资源（如有）
}

func (t *TrayManager) Stop() {
	systray.Quit()
}
