package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
)

// 站点定义：两套站点完全隔离。
var siteDefs = []SiteInfo{
	{ID: "picui", Name: "Picui 主站", BaseURL: "https://picui.cn"},
	{ID: "v2picui", Name: "Picui V2 站", BaseURL: "https://v2.picui.cn"},
}

func siteBaseURL(id string) string {
	for _, s := range siteDefs {
		if s.ID == id {
			return s.BaseURL
		}
	}
	return "https://picui.cn"
}

func siteName(id string) string {
	for _, s := range siteDefs {
		if s.ID == id {
			return s.Name
		}
	}
	return id
}

// GlobalConfig 全局（跨站点）设置。
type GlobalConfig struct {
	Theme            string `json:"theme"`
	MinimizeToTray   bool   `json:"minimizeToTray"`
	HotkeyShowWindow string `json:"hotkeyShowWindow"`
	HotkeyScreenshot string `json:"hotkeyScreenshot"`
	HotkeyClipboard  string `json:"hotkeyClipboard"`
}

// SiteConfig 单站点配置与历史。
type SiteConfig struct {
	BaseURL            string        `json:"baseUrl"`
	Token              string        `json:"token"`
	MaxConcurrency     int           `json:"maxConcurrency"`
	AutoCopyMarkdown   bool          `json:"autoCopyMarkdown"`
	DefaultPermission  int           `json:"defaultPermission"`
	DefaultStrategyID  string        `json:"defaultStrategyId"`
	DefaultAlbumID     string        `json:"defaultAlbumId"`
	Compress           bool          `json:"compress"`
	CompressFormat     string        `json:"compressFormat"`
	CompressQuality    int           `json:"compressQuality"`
	MaxWidth           int           `json:"maxWidth"`
	History            []HistoryItem `json:"history"`
}

type fileConfig struct {
	Version     int                    `json:"version"`
	FirstRun    bool                   `json:"firstRun"`
	CurrentSite string                 `json:"currentSite"`
	Global      GlobalConfig           `json:"global"`
	Sites       map[string]*SiteConfig `json:"sites"`
}

// ConfigManager 负责配置的加载、保存与站点隔离。
type ConfigManager struct {
	path string
	mu   sync.RWMutex
	data *fileConfig
}

func NewConfigManager() *ConfigManager {
	cm := &ConfigManager{}
	cm.path = configFilePath()
	cm.data = defaultConfig()
	return cm
}

func configFilePath() string {
	dir := os.Getenv("APPDATA")
	if dir == "" {
		if d, err := os.UserConfigDir(); err == nil {
			dir = d
		} else {
			dir = "."
		}
	}
	appDir := filepath.Join(dir, "PicuiUploader")
	_ = os.MkdirAll(appDir, 0755)
	return filepath.Join(appDir, "config.json")
}

func defaultConfig() *fileConfig {
	cfg := &fileConfig{
		Version:     1,
		FirstRun:    true,
		CurrentSite: "",
		Global: GlobalConfig{
			Theme:            "system",
			MinimizeToTray:   true,
			HotkeyShowWindow: "Ctrl+Shift+P",
			HotkeyScreenshot: "Ctrl+Shift+A",
			HotkeyClipboard:  "Ctrl+Shift+C",
		},
		Sites: map[string]*SiteConfig{},
	}
	for _, s := range siteDefs {
		cfg.Sites[s.ID] = defaultSiteConfig(s.BaseURL)
	}
	return cfg
}

func defaultSiteConfig(baseURL string) *SiteConfig {
	return &SiteConfig{
		BaseURL:          baseURL,
		MaxConcurrency:   3,
		AutoCopyMarkdown: true,
		DefaultPermission: 1,
		Compress:         false,
		CompressFormat:   "original",
		CompressQuality:  85,
		MaxWidth:         0,
	}
}

func (cm *ConfigManager) Load() {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	b, err := os.ReadFile(cm.path)
	if err != nil {
		return
	}
	var loaded fileConfig
	if err := json.Unmarshal(b, &loaded); err != nil {
		return
	}
	if loaded.Sites == nil {
		loaded.Sites = map[string]*SiteConfig{}
	}
	for _, s := range siteDefs {
		if loaded.Sites[s.ID] == nil {
			loaded.Sites[s.ID] = defaultSiteConfig(s.BaseURL)
		} else {
			cm.mergeDefaults(loaded.Sites[s.ID], s.BaseURL)
		}
	}
	if loaded.Global.HotkeyShowWindow == "" {
		loaded.Global.HotkeyShowWindow = "Ctrl+Shift+P"
	}
	if loaded.Global.HotkeyScreenshot == "" {
		loaded.Global.HotkeyScreenshot = "Ctrl+Shift+A"
	}
	if loaded.Global.HotkeyClipboard == "" {
		loaded.Global.HotkeyClipboard = "Ctrl+Shift+C"
	}
	if loaded.Global.Theme == "" {
		loaded.Global.Theme = "system"
	}
	cm.data = &loaded
}

func (cm *ConfigManager) mergeDefaults(s *SiteConfig, baseURL string) {
	if s.BaseURL == "" {
		s.BaseURL = baseURL
	}
	if s.MaxConcurrency == 0 {
		s.MaxConcurrency = 3
	}
	if s.CompressFormat == "" {
		s.CompressFormat = "original"
	}
	if s.CompressQuality == 0 {
		s.CompressQuality = 85
	}
}

func (cm *ConfigManager) Save() {
	cm.mu.RLock()
	data, err := json.MarshalIndent(cm.data, "", "  ")
	cm.mu.RUnlock()
	if err != nil {
		return
	}
	_ = os.WriteFile(cm.path, data, 0644)
}

func (cm *ConfigManager) GetGlobal() GlobalConfig {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	return cm.data.Global
}

func (cm *ConfigManager) currentSiteID() string {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	if cm.data.CurrentSite == "" {
		return "picui"
	}
	return cm.data.CurrentSite
}

func (cm *ConfigManager) currentSite() SiteConfig {
	id := cm.currentSiteID()
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	s := cm.data.Sites[id]
	if s == nil {
		return *defaultSiteConfig(siteBaseURL(id))
	}
	return *s
}

func (cm *ConfigManager) GetAppState() *AppState {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	sites := make([]SiteInfo, len(siteDefs))
	for i, s := range siteDefs {
		sc := cm.data.Sites[s.ID]
		sites[i] = SiteInfo{
			ID:       s.ID,
			Name:     s.Name,
			BaseURL:  s.BaseURL,
			HasToken: sc != nil && sc.Token != "",
		}
	}
	return &AppState{
		FirstRun:    cm.data.FirstRun,
		CurrentSite: cm.data.CurrentSite,
		Sites:       sites,
		Settings:    cm.buildSettingsLocked(),
	}
}

func (cm *ConfigManager) SelectSite(siteID string) error {
	if !siteExists(siteID) {
		return errInvalidSite
	}
	cm.mu.Lock()
	cm.data.CurrentSite = siteID
	cm.data.FirstRun = false
	cm.mu.Unlock()
	return nil
}

func (cm *ConfigManager) SwitchSite(siteID string) error {
	return cm.SelectSite(siteID)
}

func siteExists(id string) bool {
	for _, s := range siteDefs {
		if s.ID == id {
			return true
		}
	}
	return false
}

func (cm *ConfigManager) buildSettingsLocked() *Settings {
	id := cm.data.CurrentSite
	if id == "" {
		id = "picui"
	}
	s := cm.data.Sites[id]
	if s == nil {
		s = defaultSiteConfig(siteBaseURL(id))
	}
	return &Settings{
		SiteID:            id,
		Token:             s.Token,
		Theme:             cm.data.Global.Theme,
		MaxConcurrency:    s.MaxConcurrency,
		AutoCopyMarkdown:  s.AutoCopyMarkdown,
		DefaultPermission: s.DefaultPermission,
		DefaultStrategyID: s.DefaultStrategyID,
		DefaultAlbumID:    s.DefaultAlbumID,
		Compress:          s.Compress,
		CompressFormat:    s.CompressFormat,
		CompressQuality:   s.CompressQuality,
		MaxWidth:          s.MaxWidth,
		HotkeyShowWindow:  cm.data.Global.HotkeyShowWindow,
		HotkeyScreenshot:  cm.data.Global.HotkeyScreenshot,
		HotkeyClipboard:   cm.data.Global.HotkeyClipboard,
		MinimizeToTray:    cm.data.Global.MinimizeToTray,
	}
}

func (cm *ConfigManager) GetSettings() *Settings {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	return cm.buildSettingsLocked()
}

func (cm *ConfigManager) SaveSettings(s Settings) error {
	if !siteExists(s.SiteID) {
		return errInvalidSite
	}
	cm.mu.Lock()
	defer cm.mu.Unlock()
	sc := cm.data.Sites[s.SiteID]
	if sc == nil {
		sc = defaultSiteConfig(siteBaseURL(s.SiteID))
		cm.data.Sites[s.SiteID] = sc
	}
	sc.Token = s.Token
	sc.MaxConcurrency = clampInt(s.MaxConcurrency, 1, 10)
	sc.AutoCopyMarkdown = s.AutoCopyMarkdown
	sc.DefaultPermission = s.DefaultPermission
	sc.DefaultStrategyID = s.DefaultStrategyID
	sc.DefaultAlbumID = s.DefaultAlbumID
	sc.Compress = s.Compress
	sc.CompressFormat = s.CompressFormat
	sc.CompressQuality = clampInt(s.CompressQuality, 1, 100)
	sc.MaxWidth = clampInt(s.MaxWidth, 0, 20000)
	cm.data.Global.Theme = s.Theme
	cm.data.Global.MinimizeToTray = s.MinimizeToTray
	cm.data.Global.HotkeyShowWindow = s.HotkeyShowWindow
	cm.data.Global.HotkeyScreenshot = s.HotkeyScreenshot
	cm.data.Global.HotkeyClipboard = s.HotkeyClipboard
	cm.data.CurrentSite = s.SiteID
	cm.data.FirstRun = false
	return nil
}

func (cm *ConfigManager) SetTheme(theme string) {
	cm.mu.Lock()
	cm.data.Global.Theme = theme
	cm.mu.Unlock()
}

func (cm *ConfigManager) ApplyToAPI(api *APIClient) {
	s := cm.currentSite()
	api.SetBaseURL(s.BaseURL)
	api.SetToken(s.Token)
}

// ============================ 历史记录（按站点隔离） ============================

const maxHistory = 500

func (cm *ConfigManager) GetHistory(page, pageSize int) *HistoryPage {
	if pageSize <= 0 {
		pageSize = 20
	}
	if page <= 0 {
		page = 1
	}
	id := cm.currentSiteID()
	cm.mu.RLock()
	sc := cm.data.Sites[id]
	var all []HistoryItem
	if sc != nil {
		all = sc.History
	}
	cm.mu.RUnlock()
	total := len(all)
	start := (page - 1) * pageSize
	if start > total {
		start = total
	}
	end := start + pageSize
	if end > total {
		end = total
	}
	var data []HistoryItem
	if start < end {
		data = append([]HistoryItem{}, all[start:end]...)
	}
	return &HistoryPage{Total: total, Page: page, Data: data}
}

func (cm *ConfigManager) AddHistory(item HistoryItem) {
	if item.ID == "" {
		item.ID = newID()
	}
	id := cm.currentSiteID()
	cm.mu.Lock()
	defer cm.mu.Unlock()
	sc := cm.data.Sites[id]
	if sc == nil {
		return
	}
	sc.History = append([]HistoryItem{item}, sc.History...)
	if len(sc.History) > maxHistory {
		sc.History = sc.History[:maxHistory]
	}
}

func (cm *ConfigManager) DeleteHistory(id string) {
	sid := cm.currentSiteID()
	cm.mu.Lock()
	defer cm.mu.Unlock()
	sc := cm.data.Sites[sid]
	if sc == nil {
		return
	}
	out := sc.History[:0]
	for _, h := range sc.History {
		if h.ID != id {
			out = append(out, h)
		}
	}
	sc.History = out
}

func (cm *ConfigManager) ClearHistory() {
	sid := cm.currentSiteID()
	cm.mu.Lock()
	defer cm.mu.Unlock()
	sc := cm.data.Sites[sid]
	if sc == nil {
		return
	}
	sc.History = nil
}

// currentSiteIDForUpload 供上传队列记录站点归属。
func (cm *ConfigManager) currentSiteIDForUpload() string {
	return cm.currentSiteID()
}

func clampInt(v, lo, hi int) int {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}
