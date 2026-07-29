package main

import (
	"context"
	"os"
	"path/filepath"
	"sync"
	"time"
)

const maxRetries = 3

// uploadJob 内部任务结构，保留重试所需的源数据。
type uploadJob struct {
	task     *UploadTask
	srcType  string // "file" | "bytes"
	path     string
	data     []byte
	filename string
	mime     string
	opts     UploadOptions
	running  bool
}

// UploadQueue 异步并发上传队列。
type UploadQueue struct {
	api   *APIClient
	cfg   *ConfigManager
	app   *App
	ctx   context.Context

	mu    sync.Mutex
	jobs  map[string]*uploadJob
	sem   chan struct{}
	semMu sync.RWMutex
}

func NewUploadQueue(api *APIClient, cfg *ConfigManager, app *App) *UploadQueue {
	q := &UploadQueue{
		api:  api,
		cfg:  cfg,
		app:  app,
		jobs: map[string]*uploadJob{},
	}
	q.setConcurrency(cfg.currentSite().MaxConcurrency)
	return q
}

func (q *UploadQueue) SetContext(ctx context.Context) { q.ctx = ctx }

// Reload 在设置变更后刷新并发度。
func (q *UploadQueue) Reload() {
	n := q.cfg.currentSite().MaxConcurrency
	q.setConcurrency(n)
}

func (q *UploadQueue) setConcurrency(n int) {
	if n < 1 {
		n = 1
	}
	if n > 10 {
		n = 10
	}
	q.semMu.Lock()
	q.sem = make(chan struct{}, n)
	q.semMu.Unlock()
}

func (q *UploadQueue) currentSem() chan struct{} {
	q.semMu.RLock()
	defer q.semMu.RUnlock()
	return q.sem
}

// ---- 入队 ----

func (q *UploadQueue) EnqueueFile(path string, opts UploadOptions) string {
	id := newID()
	task := &UploadTask{
		ID:        id,
		Name:      filepath.Base(path),
		Status:    "pending",
		CreatedAt: nowMilli(),
		SiteID:    q.cfg.currentSiteIDForUpload(),
	}
	opts = q.applyDefaults(opts)
	j := &uploadJob{task: task, srcType: "file", path: path, opts: opts}
	q.add(j)
	go q.process(j)
	return id
}

func (q *UploadQueue) EnqueueBytes(data []byte, filename string, mime string, opts UploadOptions) string {
	id := newID()
	task := &UploadTask{
		ID:        id,
		Name:      filename,
		Status:    "pending",
		Size:      int64(len(data)),
		CreatedAt: nowMilli(),
		SiteID:    q.cfg.currentSiteIDForUpload(),
	}
	opts = q.applyDefaults(opts)
	j := &uploadJob{task: task, srcType: "bytes", data: data, filename: filename, mime: mime, opts: opts}
	q.add(j)
	go q.process(j)
	return id
}

func (q *UploadQueue) EnqueueClipboard(opts UploadOptions) (string, error) {
	data, _, err := readClipboardImage()
	if err != nil {
		return "", err
	}
	pngBytes, err := clipboardToPNG(data)
	if err != nil {
		return "", err
	}
	return q.EnqueueBytes(pngBytes, "clipboard.png", "image/png", opts), nil
}

func (q *UploadQueue) applyDefaults(opts UploadOptions) UploadOptions {
	s := q.cfg.currentSite()
	if opts.Permission == 0 && s.DefaultPermission != 0 {
		opts.Permission = s.DefaultPermission
	}
	if opts.Permission == 0 {
		opts.Permission = 1
	}
	if opts.AlbumID == "" {
		opts.AlbumID = s.DefaultAlbumID
	}
	if opts.StrategyID == "" {
		opts.StrategyID = s.DefaultStrategyID
	}
	if !opts.Compress {
		opts.Compress = s.Compress
	}
	if opts.CompressFormat == "" {
		opts.CompressFormat = s.CompressFormat
	}
	if opts.CompressQuality == 0 {
		opts.CompressQuality = s.CompressQuality
	}
	if opts.MaxWidth == 0 {
		opts.MaxWidth = s.MaxWidth
	}
	return opts
}

func (q *UploadQueue) add(j *uploadJob) {
	q.mu.Lock()
	q.jobs[j.task.ID] = j
	q.mu.Unlock()
	q.emitTask(j.task)
}

// ---- 调度 ----

func (q *UploadQueue) process(j *uploadJob) {
	sem := q.currentSem()
	sem <- struct{}{}
	defer func() { <-sem }()

	q.mu.Lock()
	if j.running {
		q.mu.Unlock()
		return
	}
	j.running = true
	q.mu.Unlock()

	q.run(j)

	q.mu.Lock()
	j.running = false
	q.mu.Unlock()
}

func (q *UploadQueue) run(j *uploadJob) {
	for attempt := 0; attempt <= maxRetries; attempt++ {
		q.mutate(j.task.ID, func(t *UploadTask) {
			t.Status = "uploading"
			t.Progress = 0
			t.Retries = attempt
			t.Error = ""
		})
		q.emitTask(j.task)

		err := q.attemptUpload(j)
		if err == nil {
			return
		}
		if attempt < maxRetries {
			q.mutate(j.task.ID, func(t *UploadTask) {
				t.Status = "retrying"
				t.Error = err.Error()
			})
			q.emitTask(j.task)
			select {
			case <-time.After(time.Duration(attempt+1) * 2 * time.Second):
			case <-q.ctx.Done():
				return
			}
			continue
		}
		q.mutate(j.task.ID, func(t *UploadTask) {
			t.Status = "failed"
			t.Error = err.Error()
		})
		q.emitTask(j.task)
	}
}

func (q *UploadQueue) attemptUpload(j *uploadJob) error {
	var data []byte
	var filename, mime string
	if j.srcType == "file" {
		b, err := os.ReadFile(j.path)
		if err != nil {
			return err
		}
		data, filename, mime = b, filepath.Base(j.path), detectImageMime(j.path)
	} else {
		data, filename, mime = j.data, j.filename, j.mime
	}

	if j.opts.Compress {
		if res, err := CompressImage(data, filename, CompressOptions{
			Format:   j.opts.CompressFormat,
			Quality:  j.opts.CompressQuality,
			MaxWidth: j.opts.MaxWidth,
		}); err == nil {
			data, filename, mime = res.Bytes, res.Filename, res.Mimetype
		}
	}

	q.mutate(j.task.ID, func(t *UploadTask) { t.Size = int64(len(data)) })

	ctx, cancel := context.WithTimeout(q.ctx, 5*time.Minute)
	defer cancel()
	params := UploadParams{
		Bytes:      data,
		Filename:   filename,
		Mimetype:   mime,
		AlbumID:    j.opts.AlbumID,
		StrategyID: j.opts.StrategyID,
		Permission: j.opts.Permission,
	}
	result, err := q.api.Upload(ctx, params, func(pct int) {
		q.mutate(j.task.ID, func(t *UploadTask) { t.Progress = pct })
		q.app.emit("upload:progress", map[string]interface{}{"id": j.task.ID, "progress": pct})
	})
	if err != nil {
		return err
	}

	q.mutate(j.task.ID, func(t *UploadTask) {
		t.Status = "success"
		t.Progress = 100
		t.URL = result.URL
		t.Markdown = result.Markdown
	})
	q.emitTask(j.task)

	siteID := q.cfg.currentSiteIDForUpload()
	q.cfg.AddHistory(HistoryItem{
		ID:        j.task.ID,
		Name:      filename,
		Size:      int64(len(data)),
		URL:       result.URL,
		Markdown:  result.Markdown,
		CreatedAt: nowMilli(),
		SiteID:    siteID,
	})
	q.cfg.Save()

	if sc := q.cfg.currentSite(); sc.AutoCopyMarkdown && result.Markdown != "" {
		_ = q.app.CopyText(result.Markdown)
	}
	return nil
}

// ---- 管理 ----

func (q *UploadQueue) Retry(id string) error {
	q.mu.Lock()
	j, ok := q.jobs[id]
	if !ok {
		q.mu.Unlock()
		return errNotFound
	}
	if j.running {
		q.mu.Unlock()
		return nil
	}
	j.task.Status = "pending"
	j.task.Error = ""
	q.mu.Unlock()
	q.emitTask(j.task)
	go q.process(j)
	return nil
}

func (q *UploadQueue) Remove(id string) error {
	q.mu.Lock()
	delete(q.jobs, id)
	q.mu.Unlock()
	q.app.emit("upload:removed", id)
	return nil
}

func (q *UploadQueue) Clear() error {
	q.mu.Lock()
	q.jobs = map[string]*uploadJob{}
	q.mu.Unlock()
	q.app.emit("upload:cleared", nil)
	return nil
}

func (q *UploadQueue) Snapshot() []UploadTask {
	q.mu.Lock()
	defer q.mu.Unlock()
	out := make([]UploadTask, 0, len(q.jobs))
	for _, j := range q.jobs {
		out = append(out, *j.task)
	}
	// 按创建时间升序
	sortTasks(out)
	return out
}

func (q *UploadQueue) mutate(id string, fn func(*UploadTask)) {
	q.mu.Lock()
	if j, ok := q.jobs[id]; ok {
		fn(j.task)
	}
	q.mu.Unlock()
}

func (q *UploadQueue) emitTask(t *UploadTask) {
	cp := *t
	q.app.emit("upload:task", cp)
}
