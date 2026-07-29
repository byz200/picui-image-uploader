package main

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"path/filepath"
	"strings"

	"golang.org/x/image/draw"
	_ "golang.org/x/image/bmp"
	_ "golang.org/x/image/tiff"
	_ "golang.org/x/image/webp"
)

// CompressOptions 客户端压缩选项。
type CompressOptions struct {
	Format   string // original | jpg | png
	Quality  int    // 1-100（仅 jpg 生效）
	MaxWidth int    // 0 表示不缩放
}

// CompressResult 压缩结果。
type CompressResult struct {
	Bytes    []byte
	Mimetype string
	Filename string
	Width    int
	Height   int
}

// CompressImage 解码、可选缩放、可选转码。
// 说明：本构建为单文件无 cgo，WebP「输出」编码未集成（需 cwebp）；
// WebP「输入」解码已支持（可直接上传或压缩为 JPG/PNG）。
func CompressImage(src []byte, srcName string, opts CompressOptions) (*CompressResult, error) {
	mime := detectImageMime(srcName)

	// 不压缩且不缩放：原样返回
	if (opts.Format == "" || opts.Format == "original") && opts.MaxWidth <= 0 {
		return &CompressResult{Bytes: src, Mimetype: mime, Filename: srcName}, nil
	}

	img, fmtName, err := decodeImage(src)
	if err != nil {
		return nil, fmt.Errorf("图片解码失败: %w", err)
	}

	b := img.Bounds()
	// 等比缩放
	if opts.MaxWidth > 0 && b.Dx() > opts.MaxWidth {
		newW := opts.MaxWidth
		newH := int(float64(b.Dy()) * float64(newW) / float64(b.Dx()))
		if newH < 1 {
			newH = 1
		}
		dst := image.NewRGBA(image.Rect(0, 0, newW, newH))
		draw.CatmullRom.Scale(dst, dst.Bounds(), img, b, draw.Over, nil)
		img = dst
		b = dst.Bounds()
	}

	outFmt := opts.Format
	if outFmt == "" || outFmt == "original" {
		switch fmtName {
		case "jpeg":
			outFmt = "jpg"
		case "png":
			outFmt = "png"
		default: // gif/webp/bmp/tiff → png
			outFmt = "png"
		}
	}

	buf := &bytes.Buffer{}
	var outMime, outExt string
	switch outFmt {
	case "jpg":
		quality := opts.Quality
		if quality <= 0 || quality > 100 {
			quality = 85
		}
		if err := jpeg.Encode(buf, flattenOnWhite(img), &jpeg.Options{Quality: quality}); err != nil {
			return nil, err
		}
		outMime = "image/jpeg"
		outExt = ".jpg"
	default: // png
		enc := &png.Encoder{CompressionLevel: png.BestCompression}
		if err := enc.Encode(buf, img); err != nil {
			return nil, err
		}
		outMime = "image/png"
		outExt = ".png"
	}

	return &CompressResult{
		Bytes:    buf.Bytes(),
		Mimetype: outMime,
		Filename: renameExt(srcName, outExt),
		Width:    b.Dx(),
		Height:   b.Dy(),
	}, nil
}

// flattenOnWhite 将透明图层平铺到白底（JPEG 不支持透明）。
func flattenOnWhite(img image.Image) image.Image {
	b := img.Bounds()
	dst := image.NewRGBA(b)
	draw.Draw(dst, b, &image.Uniform{C: color.White}, image.Point{}, draw.Src)
	draw.Draw(dst, b, img, b.Min, draw.Over)
	return dst
}

func renameExt(name, newExt string) string {
	base := strings.TrimSuffix(name, filepath.Ext(name))
	return base + newExt
}

// clipboardToPNG 将任意可解码图片字节转换为 PNG（用于剪贴板 BMP 统一化）。
func clipboardToPNG(src []byte) ([]byte, error) {
	img, _, err := decodeImage(src)
	if err != nil {
		return nil, err
	}
	var buf bytes.Buffer
	enc := &png.Encoder{CompressionLevel: png.BestCompression}
	if err := enc.Encode(&buf, img); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
