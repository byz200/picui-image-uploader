package main

import (
	"context"
	"encoding/base64"
	"fmt"
	"strings"
	"sync"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// App 是暴露给前端的 Wails 绑定结构体，所有前端可调用方法均定义于此。
// 后端具体能力委托给各子模块（config/api/queue/tray/hotkey/screenshot）实现。
type App struct {
	ctx context.Context

	cfg    *ConfigManager
	api    *APIClient
	queue  *UploadQueue
	tray   *TrayManager
	hotkey *HotkeyManager
	ss     *ScreenshotManager

	mu       sync.Mutex
	domReady bool
	quitting bool
}

// NewApp 构造应用实例。
func NewApp() *App {
	cfg := NewConfigManager()
	api := NewAPIClient(cfg)
	a := &App{
		cfg: cfg,
		api: api,
	}
	a.queue = NewUploadQueue(api, cfg, a)
	a.tray = NewTrayManager(a)
	a.hotkey = NewHotkeyManager(a)
	a.ss = NewScreenshotManager(a)
	return a
}

// Startup 在 Wails 启动时调用：加载配置、初始化 API、启动托盘与快捷键。
func (a *App) Startup(ctx context.Context) {
	a.ctx = ctx
	a.cfg.Load()
	a.cfg.ApplyToAPI(a.api)
	a.queue.SetContext(ctx)
	a.tray.Start()
	a.hotkey.Start()

	// 注册窗口拖拽上传：将文件路径转发给前端，由前端按当前上传选项入队
	runtime.OnFileDrop(ctx, func(x, y int, paths []string) {
		imgs := filterImagePaths(paths)
		if len(imgs) > 0 {
			a.emit("drop:files", imgs)
		}
	})
}

// filterImagePaths 过滤出图片文件路径。
func filterImagePaths(paths []string) []string {
	out := make([]string, 0, len(paths))
	for _, p := range paths {
		if isImageFile(p) {
			out = append(out, p)
		}
	}
	return out
}

// DomReady 在前端 DOM 就绪时调用。
func (a *App) DomReady(ctx context.Context) {
	a.mu.Lock()
	a.domReady = true
	a.mu.Unlock()
}

// Shutdown 在应用关闭时调用：保存配置、释放托盘与快捷键。
func (a *App) Shutdown(ctx context.Context) {
	a.quitting = true
	a.hotkey.Stop()
	a.tray.Stop()
	a.cfg.Save()
}

// BeforeClose 拦截窗口关闭：默认最小化到托盘而非退出。
func (a *App) BeforeClose(ctx context.Context) bool {
	if a.quitting {
		return false
	}
	minimize := a.cfg.GetGlobal().MinimizeToTray
	if minimize {
		runtime.WindowHide(ctx)
		return true
	}
	return false
}

// emit 向前端发送事件。
func (a *App) emit(name string, data ...interface{}) {
	if a.ctx == nil {
		return
	}
	runtime.EventsEmit(a.ctx, name, data...)
}

// defaultUploadOpts 基于当前站点设置构造上传选项（供托盘快捷操作使用）。
func (a *App) defaultUploadOpts() UploadOptions {
	s := a.cfg.GetSettings()
	perm := s.DefaultPermission
	if perm == 0 {
		perm = 1
	}
	return UploadOptions{
		AlbumID:         s.DefaultAlbumID,
		StrategyID:      s.DefaultStrategyID,
		Permission:      perm,
		Compress:        s.Compress,
		CompressFormat:  s.CompressFormat,
		CompressQuality: s.CompressQuality,
		MaxWidth:        s.MaxWidth,
	}
}

// ============================ 站点与配置 ============================

// GetAppState 返回应用启动状态（是否首启、当前站点、站点列表、设置）。
func (a *App) GetAppState() (*AppState, error) {
	return a.cfg.GetAppState(), nil
}

// SelectSite 首次启动选择站点。
func (a *App) SelectSite(siteID string) error {
	if err := a.cfg.SelectSite(siteID); err != nil {
		return err
	}
	a.cfg.ApplyToAPI(a.api)
	a.cfg.Save()
	a.emit("site:changed", siteID)
	return nil
}

// SwitchSite 设置页切换站点：清空业务上下文并加载对应站点配置。
func (a *App) SwitchSite(siteID string) error {
	if err := a.cfg.SwitchSite(siteID); err != nil {
		return err
	}
	a.cfg.ApplyToAPI(a.api)
	a.queue.Clear()
	a.cfg.Save()
	a.emit("site:changed", siteID)
	return nil
}

// GetSettings 返回当前站点的设置（合并全局项）。
func (a *App) GetSettings() (*Settings, error) {
	return a.cfg.GetSettings(), nil
}

// SaveSettings 保存设置并刷新 API 客户端与快捷键。
func (a *App) SaveSettings(s Settings) error {
	if err := a.cfg.SaveSettings(s); err != nil {
		return err
	}
	a.cfg.ApplyToAPI(a.api)
	a.queue.Reload()
	a.hotkey.Reload()
	a.cfg.Save()
	a.emit("settings:saved", s)
	return nil
}

// TestToken 校验 Token 连通性，返回用户资料。
func (a *App) TestToken(token string) (*Profile, error) {
	return a.api.TestToken(token)
}

// Login 使用账号密码登录获取 Token。
func (a *App) Login(req LoginRequest) (*LoginResult, error) {
	return a.api.Login(req)
}

// ============================ 相册 ============================

func (a *App) GetAlbums(page int, q string, order string) (*AlbumList, error) {
	return a.api.GetAlbums(page, q, order)
}

func (a *App) CreateAlbum(name string, intro string, isPublic bool) (*CreateAlbumResult, error) {
	return a.api.CreateAlbum(name, intro, isPublic)
}

func (a *App) UpdateAlbum(id int, name string, intro string, isPublic bool) error {
	return a.api.UpdateAlbum(id, name, intro, isPublic)
}

func (a *App) DeleteAlbum(id int) error {
	return a.api.DeleteAlbum(id)
}

// ============================ 储存策略 ============================

func (a *App) GetStrategies() ([]Strategy, error) {
	return a.api.GetStrategies()
}

// ============================ 图片 ============================

func (a *App) GetImages(page int, q string, order string, permission string, albumID string) (*ImageList, error) {
	return a.api.GetImages(page, q, order, permission, albumID)
}

func (a *App) DeleteImage(key int) error {
	return a.api.DeleteImage(key)
}

// ============================ 上传 ============================

// UploadFiles 上传本地文件路径列表（拖拽/选择）。
func (a *App) UploadFiles(paths []string, opts UploadOptions) ([]string, error) {
	ids := make([]string, 0, len(paths))
	for _, p := range paths {
		id := a.queue.EnqueueFile(p, opts)
		if id != "" {
			ids = append(ids, id)
		}
	}
	return ids, nil
}

// UploadClipboard 上传剪贴板中的图片。
func (a *App) UploadClipboard(opts UploadOptions) (string, error) {
	id, err := a.queue.EnqueueClipboard(opts)
	if err != nil {
		return "", err
	}
	return id, nil
}

// UploadBase64 上传 base64(data URL) 图片（截图选区裁剪结果）。
func (a *App) UploadBase64(dataURL string, filename string, opts UploadOptions) (string, error) {
	bytes, mime, err := decodeDataURL(dataURL)
	if err != nil {
		return "", err
	}
	if filename == "" {
		filename = "screenshot.png"
	}
	return a.queue.EnqueueBytes(bytes, filename, mime, opts), nil
}

func (a *App) RetryUpload(id string) error {
	return a.queue.Retry(id)
}

func (a *App) RemoveUploadTask(id string) error {
	return a.queue.Remove(id)
}

func (a *App) ClearUploadTasks() error {
	return a.queue.Clear()
}

func (a *App) GetUploadTasks() ([]UploadTask, error) {
	return a.queue.Snapshot(), nil
}

// ============================ 截图 ============================

// StartScreenshot 捕获所有显示器并向前端发送截图数据，由前端展示选区遮罩。
func (a *App) StartScreenshot() error {
	return a.ss.CaptureAll()
}

// ============================ 历史记录 ============================

func (a *App) GetHistory(page int, pageSize int) (*HistoryPage, error) {
	return a.cfg.GetHistory(page, pageSize), nil
}

func (a *App) DeleteHistory(id string) error {
	a.cfg.DeleteHistory(id)
	a.cfg.Save()
	return nil
}

func (a *App) ClearHistory() error {
	a.cfg.ClearHistory()
	a.cfg.Save()
	return nil
}

// ============================ 系统交互 ============================

func (a *App) MinimizeToTray() error {
	if a.ctx != nil {
		runtime.WindowHide(a.ctx)
	}
	return nil
}

func (a *App) QuitApp() error {
	a.quitting = true
	if a.ctx != nil {
		runtime.Quit(a.ctx)
	}
	return nil
}

func (a *App) BringToFront() error {
	if a.ctx != nil {
		runtime.WindowShow(a.ctx)
	}
	return nil
}

func (a *App) CopyText(text string) error {
	if a.ctx != nil {
		return runtime.ClipboardSetText(a.ctx, text)
	}
	return nil
}

func (a *App) OpenURL(url string) error {
	if a.ctx != nil {
		runtime.BrowserOpenURL(a.ctx, url)
	}
	return nil
}

func (a *App) SetTheme(theme string) error {
	a.cfg.SetTheme(theme)
	a.cfg.Save()
	a.emit("theme:changed", theme)
	return nil
}

// decodeDataURL 解析 data:[<mime>];base64,<data> 格式。
func decodeDataURL(dataURL string) ([]byte, string, error) {
	const prefix = "data:"
	if !strings.HasPrefix(dataURL, prefix) {
		return nil, "", fmt.Errorf("invalid data url")
	}
	rest := dataURL[len(prefix):]
	comma := strings.Index(rest, ",")
	if comma < 0 {
		return nil, "", fmt.Errorf("invalid data url")
	}
	header := rest[:comma]
	data := rest[comma+1:]
	mime := "image/png"
	if strings.Contains(header, ";") {
		mime = strings.Split(header, ";")[0]
	} else if strings.HasPrefix(header, "image/") {
		mime = header
	}
	if !strings.Contains(header, "base64") {
		return nil, "", fmt.Errorf("only base64 data url supported")
	}
	b, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		return nil, "", err
	}
	return b, mime, nil
}
