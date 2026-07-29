//go:build windows

package main

import (
	"encoding/binary"
	"errors"
	"syscall"
	"unsafe"
)

const cfDIB uint32 = 8
const gmemFixed uint32 = 0x0000

var (
	kernel32 = syscall.NewLazyDLL("kernel32.dll")
	user32   = syscall.NewLazyDLL("user32.dll")

	procOpenClipboard     = user32.NewProc("OpenClipboard")
	procCloseClipboard    = user32.NewProc("CloseClipboard")
	procGetClipboardData  = user32.NewProc("GetClipboardData")
	procGlobalSize        = kernel32.NewProc("GlobalSize")
	procGlobalLock        = kernel32.NewProc("GlobalLock")
	procGlobalUnlock      = kernel32.NewProc("GlobalUnlock")
)

// openClipboard 调用 user32!OpenClipboard。
func openClipboard(hwnd syscall.Handle) error {
	r1, _, e := procOpenClipboard.Call(uintptr(hwnd))
	if r1 == 0 {
		return e
	}
	return nil
}

// closeClipboard 调用 user32!CloseClipboard。
func closeClipboard() error {
	r1, _, e := procCloseClipboard.Call()
	if r1 == 0 {
		return e
	}
	return nil
}

// getClipboardData 调用 user32!GetClipboardData，返回 HGLOBAL。
func getClipboardData(format uint32) (syscall.Handle, error) {
	r1, _, e := procGetClipboardData.Call(uintptr(format))
	if r1 == 0 {
		return 0, e
	}
	return syscall.Handle(r1), nil
}

// globalSize 调用 kernel32!GlobalSize。
func globalSize(h syscall.Handle) (uintptr, error) {
	r1, _, e := procGlobalSize.Call(uintptr(h))
	if r1 == 0 {
		return 0, e
	}
	return r1, nil
}

// globalLock 调用 kernel32!GlobalLock，返回内存指针。
func globalLock(h syscall.Handle) (unsafe.Pointer, error) {
	r1, _, e := procGlobalLock.Call(uintptr(h))
	if r1 == 0 {
		return nil, e
	}
	return unsafe.Pointer(r1), nil
}

// globalUnlock 调用 kernel32!GlobalUnlock。
func globalUnlock(h syscall.Handle) error {
	r1, _, e := procGlobalUnlock.Call(uintptr(h))
	// GlobalUnlock 返回 0 表示失败，e 会是 GetLastError
	if r1 == 0 && e != syscall.Errno(0) {
		return e
	}
	return nil
}

// readClipboardImage 从系统剪贴板读取图片（CF_DIB），并封装为可解码的 BMP 字节。
func readClipboardImage() ([]byte, string, error) {
	if err := openClipboard(0); err != nil {
		return nil, "", errors.New("打开剪贴板失败")
	}
	defer closeClipboard()

	h, err := getClipboardData(cfDIB)
	if err != nil || h == 0 {
		return nil, "", errors.New("剪贴板中没有图片数据")
	}
	size, err := globalSize(h)
	if err != nil || size == 0 {
		return nil, "", errors.New("无法读取剪贴板图片大小")
	}
	ptr, err := globalLock(h)
	if err != nil {
		return nil, "", errors.New("无法锁定剪贴板内存")
	}
	defer globalUnlock(h)

	dib := make([]byte, size)
	copy(dib, (*[1 << 30]byte)(ptr)[:size:size])

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
