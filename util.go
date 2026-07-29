package main

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"os"
	"path/filepath"
	"strings"
	"time"
)

var errInvalidSite = errors.New("无效的站点标识")

var errNotFound = errors.New("任务不存在")

// sortTasks 按创建时间升序排序。
func sortTasks(tasks []UploadTask) {
	for i := 1; i < len(tasks); i++ {
		for j := i; j > 0 && tasks[j-1].CreatedAt > tasks[j].CreatedAt; j-- {
			tasks[j-1], tasks[j] = tasks[j], tasks[j-1]
		}
	}
}

// newID 生成短随机 ID。
func newID() string {
	b := make([]byte, 8)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

// formatFileSize 友好显示文件大小。
func formatFileSize(n int64) string {
	const u = 1024
	if n < u {
		return fmt.Sprintf("%d B", n)
	}
	div, exp := int64(u), 0
	for x := n / u; x >= u; x /= u {
		div *= u
		exp++
	}
	units := []string{"KB", "MB", "GB", "TB"}
	return fmt.Sprintf("%.2f %s", float64(n)/float64(div), units[exp])
}

// fileExists 判断文件是否存在。
func fileExists(p string) bool {
	_, err := os.Stat(p)
	return err == nil
}

// detectImageMime 根据扩展名推断 mime。
func detectImageMime(name string) string {
	ext := strings.ToLower(filepath.Ext(name))
	switch ext {
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".png":
		return "image/png"
	case ".gif":
		return "image/gif"
	case ".webp":
		return "image/webp"
	case ".bmp":
		return "image/bmp"
	case ".tiff", ".tif":
		return "image/tiff"
	default:
		return "application/octet-stream"
	}
}

// isImageFile 判断路径是否为支持的图片文件。
func isImageFile(p string) bool {
	return detectImageMime(p) != "application/octet-stream"
}

// nowMilli 当前毫秒时间戳。
func nowMilli() int64 {
	return time.Now().UnixMilli()
}

// decodeImage 通用图片解码（含 webp/bmp/tiff）。
func decodeImage(b []byte) (image.Image, string, error) {
	cfg, fmtName, err := image.DecodeConfig(bytes.NewReader(b))
	if err == nil {
		img, _, err2 := image.Decode(bytes.NewReader(b))
		if err2 != nil {
			return nil, fmtName, err2
		}
		return img, fmtName, nil
	}
	return nil, "", err
}
