package main

import (
	"strings"
	"sync"

	"golang.design/x/hotkey"
)

type hkEntry struct {
	hk   *hotkey.Hotkey
	stop chan struct{}
}

// HotkeyManager 全局快捷键管理。
type HotkeyManager struct {
	app    *App
	mu     sync.Mutex
	entries []*hkEntry
	stopped bool
}

func NewHotkeyManager(app *App) *HotkeyManager {
	return &HotkeyManager{app: app}
}

func (h *HotkeyManager) Start()  { h.reload() }
func (h *HotkeyManager) Reload() { h.reload() }

func (h *HotkeyManager) reload() {
	h.mu.Lock()
	if h.stopped {
		h.mu.Unlock()
		return
	}
	// 清理旧的
	for _, e := range h.entries {
		close(e.stop)
		_ = e.hk.Unregister()
	}
	h.entries = nil
	h.mu.Unlock()

	g := h.app.cfg.GetGlobal()
	h.register(g.HotkeyShowWindow, func() { h.app.BringToFront() })
	h.register(g.HotkeyScreenshot, func() {
		h.app.BringToFront()
		_ = h.app.ss.CaptureAll()
	})
	h.register(g.HotkeyClipboard, func() {
		_, _ = h.app.queue.EnqueueClipboard(h.app.defaultUploadOpts())
	})
}

func (h *HotkeyManager) register(combo string, action func()) {
	mods, key, ok := parseHotkey(combo)
	if !ok {
		return
	}
	hk := hotkey.New(mods, key)
	if err := hk.Register(); err != nil {
		return
	}
	e := &hkEntry{hk: hk, stop: make(chan struct{})}
	h.mu.Lock()
	h.entries = append(h.entries, e)
	h.mu.Unlock()
	go func() {
		for {
			select {
			case <-e.stop:
				return
			case <-hk.Keydown():
				action()
			}
		}
	}()
}

func (h *HotkeyManager) Stop() {
	h.mu.Lock()
	if h.stopped {
		h.mu.Unlock()
		return
	}
	h.stopped = true
	for _, e := range h.entries {
		close(e.stop)
		_ = e.hk.Unregister()
	}
	h.entries = nil
	h.mu.Unlock()
}

// parseHotkey 解析 "Ctrl+Shift+P" 形式的快捷键。
func parseHotkey(combo string) ([]hotkey.Modifier, hotkey.Key, bool) {
	combo = strings.TrimSpace(combo)
	if combo == "" {
		return nil, 0, false
	}
	parts := strings.Split(combo, "+")
	if len(parts) < 2 {
		return nil, 0, false
	}
	keyStr := strings.TrimSpace(parts[len(parts)-1])
	var mods []hotkey.Modifier
	for _, p := range parts[:len(parts)-1] {
		switch strings.ToLower(strings.TrimSpace(p)) {
		case "ctrl", "control":
			mods = append(mods, hotkey.ModCtrl)
		case "shift":
			mods = append(mods, hotkey.ModShift)
		case "alt", "option":
			mods = append(mods, hotkey.ModAlt)
		default:
			return nil, 0, false
		}
	}
	key, ok := parseKey(keyStr)
	if !ok {
		return nil, 0, false
	}
	return mods, key, true
}

func parseKey(s string) (hotkey.Key, bool) {
	s = strings.ToLower(strings.TrimSpace(s))
	if len(s) == 1 {
		c := s[0]
		if c >= 'a' && c <= 'z' {
			if k, ok := letterKeys[c]; ok {
				return k, true
			}
		}
		if c >= '0' && c <= '9' {
			if k, ok := digitKeys[c]; ok {
				return k, true
			}
		}
	}
	if k, ok := specialKeys[s]; ok {
		return k, true
	}
	return 0, false
}

var letterKeys = map[byte]hotkey.Key{
	'a': hotkey.KeyA, 'b': hotkey.KeyB, 'c': hotkey.KeyC, 'd': hotkey.KeyD,
	'e': hotkey.KeyE, 'f': hotkey.KeyF, 'g': hotkey.KeyG, 'h': hotkey.KeyH,
	'i': hotkey.KeyI, 'j': hotkey.KeyJ, 'k': hotkey.KeyK, 'l': hotkey.KeyL,
	'm': hotkey.KeyM, 'n': hotkey.KeyN, 'o': hotkey.KeyO, 'p': hotkey.KeyP,
	'q': hotkey.KeyQ, 'r': hotkey.KeyR, 's': hotkey.KeyS, 't': hotkey.KeyT,
	'u': hotkey.KeyU, 'v': hotkey.KeyV, 'w': hotkey.KeyW, 'x': hotkey.KeyX,
	'y': hotkey.KeyY, 'z': hotkey.KeyZ,
}

var digitKeys = map[byte]hotkey.Key{
	'0': hotkey.Key0, '1': hotkey.Key1, '2': hotkey.Key2, '3': hotkey.Key3,
	'4': hotkey.Key4, '5': hotkey.Key5, '6': hotkey.Key6, '7': hotkey.Key7,
	'8': hotkey.Key8, '9': hotkey.Key9,
}

var specialKeys = map[string]hotkey.Key{
	"f1": hotkey.KeyF1, "f2": hotkey.KeyF2, "f3": hotkey.KeyF3, "f4": hotkey.KeyF4,
	"f5": hotkey.KeyF5, "f6": hotkey.KeyF6, "f7": hotkey.KeyF7, "f8": hotkey.KeyF8,
	"f9": hotkey.KeyF9, "f10": hotkey.KeyF10, "f11": hotkey.KeyF11, "f12": hotkey.KeyF12,
	"space":   hotkey.KeySpace,
	"tab":     hotkey.KeyTab,
	"enter":   hotkey.KeyReturn,
	"return":  hotkey.KeyReturn,
	"esc":     hotkey.KeyEscape,
	"escape":  hotkey.KeyEscape,
}
