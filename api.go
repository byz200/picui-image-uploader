package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// APIClient 负责与 Picui API 交互，复用 HTTP 连接池。
type APIClient struct {
	baseURL string
	token   string
	http    *http.Client
}

func NewAPIClient(_ *ConfigManager) *APIClient {
	transport := &http.Transport{
		Proxy:               http.ProxyFromEnvironment,
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 32,
		IdleConnTimeout:     90 * time.Second,
		TLSHandshakeTimeout: 15 * time.Second,
	}
	return &APIClient{
		http: &http.Client{Transport: transport},
	}
}

func (c *APIClient) SetBaseURL(u string) { c.baseURL = strings.TrimRight(u, "/") }
func (c *APIClient) SetToken(t string)    { c.token = t }

// UploadParams 上传参数。
type UploadParams struct {
	Bytes      []byte
	Filename   string
	Mimetype   string
	AlbumID    string
	StrategyID string
	Permission int
	ExpiredAt  string
	Token      string // 临时上传 token（可选）
}

// UploadResult 上传成功后解析出的关键信息。
type UploadResult struct {
	Name     string
	URL      string
	Markdown string
	Links    ImageLinks
}

// ---- 统一响应 ----

type apiResponse struct {
	Status  bool            `json:"status"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data"`
}

func (c *APIClient) authHeader(token string) string {
	t := token
	if t == "" {
		t = c.token
	}
	return "Bearer " + t
}

// request 通用请求，返回响应体字节。
func (c *APIClient) request(ctx context.Context, method, path string, query url.Values, body io.Reader, contentType, tokenOverride string) ([]byte, int, error) {
	u := c.baseURL + path
	if len(query) > 0 {
		u += "?" + query.Encode()
	}
	req, err := http.NewRequestWithContext(ctx, method, u, body)
	if err != nil {
		return nil, 0, err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", c.authHeader(tokenOverride))
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp.StatusCode, err
	}
	return b, resp.StatusCode, nil
}

func (c *APIClient) doJSON(ctx context.Context, method, path string, query url.Values, payload interface{}, tokenOverride string) (json.RawMessage, error) {
	var body io.Reader
	var ct string
	if payload != nil {
		jb, err := json.Marshal(payload)
		if err != nil {
			return nil, err
		}
		body = bytes.NewReader(jb)
		ct = "application/json"
	}
	b, status, err := c.request(ctx, method, path, query, body, ct, tokenOverride)
	if err != nil {
		return nil, err
	}
	if status >= 400 {
		return nil, fmt.Errorf("HTTP %d: %s", status, truncate(string(b), 300))
	}
	var ar apiResponse
	if err := json.Unmarshal(b, &ar); err != nil {
		// 部分端点返回 status 为字符串，尝试宽松解析
		var loose struct {
			Status  interface{}     `json:"status"`
			Message string          `json:"message"`
			Data    json.RawMessage `json:"data"`
		}
		if err2 := json.Unmarshal(b, &loose); err2 == nil {
			if s, ok := loose.Status.(string); ok && s != "success" {
				return nil, fmt.Errorf("%s", orMsg(loose.Message, "请求失败"))
			}
			return loose.Data, nil
		}
		return nil, fmt.Errorf("响应解析失败: %v", err)
	}
	if !ar.Status {
		return nil, fmt.Errorf("%s", orMsg(ar.Message, "请求失败"))
	}
	return ar.Data, nil
}

func orMsg(m, def string) string {
	if m == "" {
		return def
	}
	return m
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}

// ============================ 鉴权 ============================

// TestToken 使用给定 token 调用 /api/v1/profile 校验连通性。
func (c *APIClient) TestToken(token string) (*Profile, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	data, err := c.doJSON(ctx, "GET", "/api/v1/profile", nil, nil, token)
	if err != nil {
		return nil, err
	}
	var rp rawProfile
	if err := json.Unmarshal(data, &rp); err != nil {
		return nil, err
	}
	return &Profile{
		Avatar:       rp.Avatar,
		Name:         rp.Name,
		Username:     rp.Username,
		Email:        rp.Email,
		ImageNum:     rp.ImageNum,
		AlbumNum:     rp.AlbumNum,
		RegisteredIP: rp.RegisteredIP,
		URL:          rp.URL,
		Capacity:     rp.Capacity,
		Size:         rp.Size,
	}, nil
}

// Login 账号密码登录获取 Token。
func (c *APIClient) Login(req LoginRequest) (*LoginResult, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	payload := map[string]interface{}{
		"login_type":   req.LoginType,
		"username":     req.Username,
		"password":     req.Password,
		"remember":     req.Remember,
		"country_code": req.CountryCode,
	}
	if req.LoginType == "" {
		payload["login_type"] = "username"
	}
	if req.CountryCode == "" {
		payload["country_code"] = "cn"
	}
	b, status, err := c.request(ctx, "POST", "/login", nil, jsonReader(payload), "application/json", "")
	if err != nil {
		return nil, err
	}
	if status == 422 {
		return nil, fmt.Errorf("账号或密码错误（%d）", status)
	}
	if status >= 400 {
		return nil, fmt.Errorf("登录失败: HTTP %d %s", status, truncate(string(b), 200))
	}
	var loose struct {
		Status  string `json:"status"`
		Message string `json:"message"`
		Data    struct {
			Name  string `json:"name"`
			Token string `json:"token"`
		} `json:"data"`
	}
	if err := json.Unmarshal(b, &loose); err != nil {
		return nil, err
	}
	if loose.Data.Token == "" {
		return nil, fmt.Errorf("%s", orMsg(loose.Message, "登录失败，未返回 Token"))
	}
	return &LoginResult{Name: loose.Data.Name, Token: loose.Data.Token}, nil
}

func jsonReader(v interface{}) io.Reader {
	b, _ := json.Marshal(v)
	return bytes.NewReader(b)
}

// ============================ 相册 ============================

func (c *APIClient) GetAlbums(page int, q string, order string) (*AlbumList, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	query := url.Values{}
	if page > 0 {
		query.Set("page", strconv.Itoa(page))
	}
	if q != "" {
		query.Set("q", q)
	}
	if order != "" {
		query.Set("order", order)
	}
	data, err := c.doJSON(ctx, "GET", "/api/v1/albums", query, nil, "")
	if err != nil {
		return nil, err
	}
	var raw rawAlbumList
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, err
	}
	out := &AlbumList{
		CurrentPage: raw.CurrentPage,
		LastPage:    raw.LastPage,
		PerPage:     raw.PerPage,
		Total:       raw.Total,
		Data:        make([]Album, 0, len(raw.Data)),
	}
	for _, a := range raw.Data {
		out.Data = append(out.Data, Album{ID: a.ID, Name: a.Name, Intro: a.Intro, ImageNum: a.ImageNum})
	}
	return out, nil
}

func (c *APIClient) CreateAlbum(name string, intro string, isPublic bool) (*CreateAlbumResult, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	payload := map[string]interface{}{
		"name":      name,
		"intro":     intro,
		"is_public": boolStr(isPublic),
	}
	data, err := c.doJSON(ctx, "POST", "/user/albums", nil, payload, "")
	if err != nil {
		return nil, err
	}
	var res CreateAlbumResult
	_ = json.Unmarshal(data, &res)
	return &res, nil
}

func (c *APIClient) UpdateAlbum(id int, name string, intro string, isPublic bool) error {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	payload := map[string]interface{}{
		"name":      name,
		"intro":     intro,
		"is_public": boolStr(isPublic),
	}
	_, err := c.doJSON(ctx, "PUT", fmt.Sprintf("/user/albums/%d", id), nil, payload, "")
	return err
}

func (c *APIClient) DeleteAlbum(id int) error {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	_, err := c.doJSON(ctx, "DELETE", fmt.Sprintf("/api/v1/albums/%d", id), nil, nil, "")
	return err
}

// ============================ 储存策略 ============================

func (c *APIClient) GetStrategies() ([]Strategy, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	data, err := c.doJSON(ctx, "GET", "/api/v1/strategies", nil, nil, "")
	if err != nil {
		return nil, err
	}
	out := make([]Strategy, 0)
	// 文档未给出明确数组结构，兼容 data 为数组或 {data:[]}
	var arr []map[string]interface{}
	if json.Unmarshal(data, &arr) == nil {
		for _, m := range arr {
			out = append(out, strategyFromMap(m))
		}
		return out, nil
	}
	var wrap struct {
		Data []map[string]interface{} `json:"data"`
	}
	if json.Unmarshal(data, &wrap) == nil {
		for _, m := range wrap.Data {
			out = append(out, strategyFromMap(m))
		}
	}
	return out, nil
}

func strategyFromMap(m map[string]interface{}) Strategy {
	s := Strategy{ID: m["id"]}
	if v, ok := m["name"].(string); ok {
		s.Name = v
	}
	if v, ok := m["intro"].(string); ok {
		s.Intro = v
	}
	if v, ok := m["is_default"].(bool); ok {
		s.IsDefault = v
	}
	return s
}

// ============================ 图片 ============================

func (c *APIClient) GetImages(page int, q string, order string, permission string, albumID string) (*ImageList, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	query := url.Values{}
	if page > 0 {
		query.Set("page", strconv.Itoa(page))
	}
	if q != "" {
		query.Set("q", q)
	}
	if order != "" {
		query.Set("order", order)
	}
	if permission != "" {
		query.Set("permission", permission)
	}
	if albumID != "" {
		query.Set("album_id", albumID)
	}
	data, err := c.doJSON(ctx, "GET", "/api/v1/images", query, nil, "")
	if err != nil {
		return nil, err
	}
	var raw rawImageList
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, err
	}
	out := &ImageList{
		CurrentPage: raw.CurrentPage,
		LastPage:    raw.LastPage,
		PerPage:     raw.PerPage,
		Total:       raw.Total,
		Data:        make([]ImageItem, 0, len(raw.Data)),
	}
	for _, im := range raw.Data {
		out.Data = append(out.Data, ImageItem{
			Key:       im.Key,
			Name:      im.Name,
			Pathname:  im.Pathname,
			Mimetype:  im.Mimetype,
			Extension: im.Extension,
			Size:      im.Size,
			Width:     im.Width,
			Height:    im.Height,
			HumanDate: im.HumanDate,
			Date:      im.Date,
			Links:     convertLinks(im.Links),
		})
	}
	return out, nil
}

func (c *APIClient) DeleteImage(key int) error {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	_, err := c.doJSON(ctx, "DELETE", fmt.Sprintf("/api/v1/images/%d", key), nil, nil, "")
	return err
}

// ============================ 上传 ============================

// Upload 执行上传，onProgress 回调 0-100。
func (c *APIClient) Upload(ctx context.Context, p UploadParams, onProgress func(int)) (*UploadResult, error) {
	body := &bytes.Buffer{}
	w := multipart.NewWriter(body)

	h := textproto.MIMEHeader{}
	h.Set("Content-Disposition", fmt.Sprintf(`form-data; name="file"; filename="%s"`, quoteEscape(p.Filename)))
	if p.Mimetype != "" {
		h.Set("Content-Type", p.Mimetype)
	}
	part, err := w.CreatePart(h)
	if err != nil {
		return nil, err
	}
	if _, err := part.Write(p.Bytes); err != nil {
		return nil, err
	}
	_ = w.WriteField("permission", strconv.Itoa(p.Permission))
	if p.Token != "" {
		_ = w.WriteField("token", p.Token)
	}
	if p.StrategyID != "" {
		_ = w.WriteField("strategy_id", p.StrategyID)
	}
	if p.AlbumID != "" {
		_ = w.WriteField("album_id", p.AlbumID)
	}
	if p.ExpiredAt != "" {
		_ = w.WriteField("expired_at", p.ExpiredAt)
	}
	w.Close()

	total := int64(body.Len())
	pr := &progressReader{r: body, total: total, onProgress: onProgress}

	req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/api/v1/upload", pr)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", c.authHeader(""))
	req.ContentLength = total
	req.Header.Set("Content-Type", w.FormDataContentType())

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("上传失败 HTTP %d: %s", resp.StatusCode, truncate(string(b), 300))
	}
	var ar apiResponse
	if err := json.Unmarshal(b, &ar); err != nil {
		return nil, fmt.Errorf("上传响应解析失败: %v", err)
	}
	if !ar.Status {
		return nil, fmt.Errorf("%s", orMsg(ar.Message, "上传失败"))
	}
	var rd rawUploadData
	if err := json.Unmarshal(ar.Data, &rd); err != nil {
		return nil, err
	}
	links := convertLinks(rd.Links)
	return &UploadResult{
		Name:     rd.Name,
		URL:      links.URL,
		Markdown: links.Markdown,
		Links:    links,
	}, nil
}

type progressReader struct {
	r          io.Reader
	total      int64
	read       int64
	onProgress func(int)
}

func (pr *progressReader) Read(p []byte) (int, error) {
	n, err := pr.r.Read(p)
	if n > 0 {
		pr.read += int64(n)
		if pr.onProgress != nil && pr.total > 0 {
			pct := int(float64(pr.read) * 100 / float64(pr.total))
			if pct > 100 {
				pct = 100
			}
			pr.onProgress(pct)
		}
	}
	return n, err
}

// ============================ 原始结构 ============================

type rawProfile struct {
	Avatar       string  `json:"avatar"`
	Name         string  `json:"name"`
	Username     string  `json:"username"`
	Email        string  `json:"email"`
	ImageNum     int     `json:"image_num"`
	AlbumNum     int     `json:"album_num"`
	RegisteredIP string  `json:"registered_ip"`
	URL          string  `json:"url"`
	Capacity     int64   `json:"capacity"`
	Size         float64 `json:"size"`
}

type rawAlbum struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Intro    string `json:"intro"`
	ImageNum int    `json:"image_num"`
}

type rawAlbumList struct {
	CurrentPage int        `json:"current_page"`
	LastPage    int        `json:"last_page"`
	PerPage     int        `json:"per_page"`
	Total       int        `json:"total"`
	Data        []rawAlbum `json:"data"`
}

type rawLinks struct {
	URL              string `json:"url"`
	HTML             string `json:"html"`
	Bbcode           string `json:"bbcode"`
	Markdown         string `json:"markdown"`
	MarkdownWithLink string `json:"markdown_with_link"`
	ThumbnailURL     string `json:"thumbnail_url"`
	DeleteURL        string `json:"delete_url"`
}

type rawUploadData struct {
	Key        int      `json:"key"`
	Name       string   `json:"name"`
	Pathname   string   `json:"pathname"`
	OriginName string   `json:"origin_name"`
	Size       float64  `json:"size"`
	Mimetype   string   `json:"mimetype"`
	Extension  string   `json:"extension"`
	Links      rawLinks `json:"links"`
}

type rawImage struct {
	Key       int      `json:"key"`
	Name      string   `json:"name"`
	Pathname  string   `json:"pathname"`
	Mimetype  string   `json:"mimetype"`
	Extension string   `json:"extension"`
	Size      float64  `json:"size"`
	Width     int      `json:"width"`
	Height    int      `json:"height"`
	HumanDate string   `json:"human_date"`
	Date      string   `json:"date"`
	Links     rawLinks `json:"links"`
}

type rawImageList struct {
	CurrentPage int        `json:"current_page"`
	LastPage    int        `json:"last_page"`
	PerPage     int        `json:"per_page"`
	Total       int        `json:"total"`
	Data        []rawImage `json:"data"`
}

func convertLinks(l rawLinks) ImageLinks {
	return ImageLinks{
		URL:              l.URL,
		HTML:             l.HTML,
		Bbcode:           l.Bbcode,
		Markdown:         l.Markdown,
		MarkdownWithLink: l.MarkdownWithLink,
		ThumbnailURL:     l.ThumbnailURL,
		DeleteURL:        l.DeleteURL,
	}
}

func quoteEscape(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, `"`, `\"`)
	return s
}

func boolStr(b bool) string {
	if b {
		return "1"
	}
	return "0"
}
