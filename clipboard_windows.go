//go:build windows

package main

import (
	"encoding/binary"
	"errors"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

const cfDIB uint32 = 8

var (
	kernel32           = syscall.NewLazyDLL("kernel32.dll")
	procGlobalUnlock   = kernel32.NewProc("GlobalUnlock")
)

// globalUnlock 调用 kernel32!GlobalUnlock，兼容不同 x/sys/windows 版本的签名差异。
func globalUnlock(mem windows.Handle) (bool, error) {
	r1, _, e := procGlobalUnlock.Call(uintptr(mem))
	if r1 == 0 {
		if e != syscall.Errno(0) {
			return false, e
		}
	}
	return true, nil
}

// readClipboardImage 从系统剪贴板读取图片（CF_DIB），并封装为可解码的 BMP 字节。
func readClipboardImage() ([]byte, string, error) {
	if err := windows.OpenClipboard(0); err != nil {
		return nil, "", err
	}
	defer windows.CloseClipboard()

	h, err := windows.GetClipboardData(cfDIB)
	if err != nil || h == 0 {
		return nil, "", errors.New("剪贴板中没有图片数据")
	}
	size, err := windows.GlobalSize(windows.Handle(h))
	if err != nil || size == 0 {
		return nil, "", errors.New("无法读取剪贴板图片大小")
	}
	mem := windows.Handle(h)
	ptr, err := windows.GlobalLock(mem)
	if err != nil {
		return nil, "", err
	}
	defer func() {
		// GlobalUnlock returns (BOOL stillLocked, error) on newer x/sys/windows.
		// If x/sys/windows exposes a different signature (error only), we fallback via unsafe lookup below.
		if _, e := globalUnlock(mem); e != nil {
			// best effort, ignore on error
		}
	}()

	dib := make([]byte, size)
	copy(dib, (*[1 << 30]byte)(unsafe.Pointer(ptr))[:size:size])

	bmp, err := dibToBMP(dib)
	if err != nil {
		return nil, "", err
	}
	return bmp, "clipboard.png", nil
}

// dibToBMP 将 CF_DIB 数据（无文件头）封装为完整 BMP 文件字节。
func dibToBMP(dib []byte) ([]byte, error) {
	if len(dib) < 40 {
		return nil, errors.New("invalid DIB data")
	}
	biSize := binary.LittleEndian.Uint32(dib[0:4])
	if biSize < 40 || int(biSize) > len(dib) {
		return nil, errors.New("unsupported DIB header size")
	}
	width := int(int32(binary.LittleEndian.Uint32(dib[4:8])))
	height := int(int32(binary.LittleEndian.Uint32(dib[8:12])))
	bitCount := binary.LittleEndian.Uint16(dib[14:16])
	biSizeImage := binary.LittleEndian.Uint32(dib[20:24])

	var paletteSize uint32
	if bitCount <= 8 {
		paletteSize = (1 << bitCount) * 4
	}
	offBits := 14 + biSize + paletteSize

	if biSizeImage == 0 {
		rowSize := ((uint32(width)*uint32(bitCount) + 31) / 32) * 4
		biSizeImage = rowSize * uint32(absInt(height))
	}

	fileHeader := make([]byte, 14)
	fileHeader[0] = 'B'
	fileHeader[1] = 'M'
	binary.LittleEndian.PutUint32(fileHeader[2:6], uint32(14)+uint32(len(dib)))
	binary.LittleEndian.PutUint16(fileHeader[6:8], 0)
	binary.LittleEndian.PutUint16(fileHeader[8:10], 0)
	binary.LittleEndian.PutUint32(fileHeader[10:14], offBits)

	out := make([]byte, 0, 14+len(dib))
	out = append(out, fileHeader...)
	out = append(out, dib...)
	return out, nil
}

func absInt(v int) int {
	if v < 0 {
		return -v
	}
	return v
}
