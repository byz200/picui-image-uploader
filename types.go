package main

// ============================ 应用状态与设置 ============================

// AppState 启动时返回给前端的应用状态。
type AppState struct {
	FirstRun    bool       `json:"firstRun"`
	CurrentSite string     `json:"currentSite"`
	Sites       []SiteInfo `json:"sites"`
	Settings    *Settings  `json:"settings"`
}

// SiteInfo 站点元信息。
type SiteInfo struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	BaseURL  string `json:"baseUrl"`
	HasToken bool   `json:"hasToken"`
}

// Settings 暴露给前端的设置视图（合并全局与当前站点）。
type Settings struct {
	SiteID            string `json:"siteId"`
	Token             string `json:"token"`
	Theme             string `json:"theme"`
	MaxConcurrency    int    `json:"maxConcurrency"`
	AutoCopyMarkdown  bool   `json:"autoCopyMarkdown"`
	DefaultPermission int    `json:"defaultPermission"`
	DefaultStrategyID string `json:"defaultStrategyId"`
	DefaultAlbumID    string `json:"defaultAlbumId"`
	Compress          bool   `json:"compress"`
	CompressFormat    string `json:"compressFormat"` // jpg/png/original
	CompressQuality   int    `json:"compressQuality"`
	MaxWidth          int    `json:"maxWidth"`
	HotkeyShowWindow  string `json:"hotkeyShowWindow"`
	HotkeyScreenshot  string `json:"hotkeyScreenshot"`
	HotkeyClipboard   string `json:"hotkeyClipboard"`
	MinimizeToTray    bool   `json:"minimizeToTray"`
}

// ============================ 登录 ============================

type LoginRequest struct {
	LoginType   string `json:"loginType"` // username/email/phone
	Username    string `json:"username"`
	Password    string `json:"password"`
	Remember    bool   `json:"remember"`
	CountryCode string `json:"countryCode"`
}

type LoginResult struct {
	Name  string `json:"name"`
	Token string `json:"token"`
}

// ============================ 用户资料 ============================

type Profile struct {
	Avatar       string  `json:"avatar"`
	Name         string  `json:"name"`
	Username     string  `json:"username"`
	Email        string  `json:"email"`
	ImageNum     int     `json:"imageNum"`
	AlbumNum     int     `json:"albumNum"`
	RegisteredIP string  `json:"registeredIp"`
	URL          string  `json:"url"`
	Capacity     int64   `json:"capacity"`
	Size         float64 `json:"size"`
}

// ============================ 相册 ============================

type Album struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Intro    string `json:"intro"`
	ImageNum int    `json:"imageNum"`
}

type AlbumList struct {
	CurrentPage int     `json:"currentPage"`
	LastPage    int     `json:"lastPage"`
	PerPage     int     `json:"perPage"`
	Total       int     `json:"total"`
	Data        []Album `json:"data"`
}

type CreateAlbumResult struct {
	ID int `json:"id"`
}

// ============================ 储存策略 ============================

type Strategy struct {
	ID      interface{} `json:"id"`
	Name    string      `json:"name"`
	Intro   string      `json:"intro"`
	IsDefault bool      `json:"isDefault"`
}

// ============================ 图片 ============================

type ImageLinks struct {
	URL             string `json:"url"`
	HTML            string `json:"html"`
	Bbcode          string `json:"bbcode"`
	Markdown        string `json:"markdown"`
	MarkdownWithLink string `json:"markdownWithLink"`
	ThumbnailURL    string `json:"thumbnailUrl"`
	DeleteURL       string `json:"deleteUrl"`
}

type ImageItem struct {
	Key         int        `json:"key"`
	Name        string     `json:"name"`
	Pathname    string     `json:"pathname"`
	Mimetype    string     `json:"mimetype"`
	Extension   string     `json:"extension"`
	Size        float64    `json:"size"`
	Width       int        `json:"width"`
	Height      int        `json:"height"`
	HumanDate   string     `json:"humanDate"`
	Date        string     `json:"date"`
	Links       ImageLinks `json:"links"`
}

type ImageList struct {
	CurrentPage int        `json:"currentPage"`
	LastPage    int        `json:"lastPage"`
	PerPage     int        `json:"perPage"`
	Total       int        `json:"total"`
	Data        []ImageItem `json:"data"`
}

// ============================ 上传 ============================

type UploadOptions struct {
	AlbumID         string `json:"albumId"`
	StrategyID      string `json:"strategyId"`
	Permission      int    `json:"permission"`
	Compress        bool   `json:"compress"`
	CompressFormat  string `json:"compressFormat"`
	CompressQuality int    `json:"compressQuality"`
	MaxWidth        int    `json:"maxWidth"`
}

type UploadTask struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Size      int64  `json:"size"`
	Status    string `json:"status"` // pending/uploading/success/failed/retrying/canceled
	Progress  int    `json:"progress"`
	Error     string `json:"error"`
	URL       string `json:"url"`
	Markdown  string `json:"markdown"`
	CreatedAt int64  `json:"createdAt"`
	Retries   int    `json:"retries"`
	SiteID    string `json:"siteId"`
}

type HistoryPage struct {
	Total int           `json:"total"`
	Page  int           `json:"page"`
	Data  []HistoryItem `json:"data"`
}

type HistoryItem struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Size      int64  `json:"size"`
	URL       string `json:"url"`
	Markdown  string `json:"markdown"`
	CreatedAt int64  `json:"createdAt"`
	SiteID    string `json:"siteId"`
}
