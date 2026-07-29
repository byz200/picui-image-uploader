// C:\Users\Administrator\Documents\trae_projects\picui\frontend\src\wailsjs\runtime.js - Wails 运行时桥接（手工同步核心 API）
// eslint-disable-next-line no-undef
export function EventsOn(eventName, ...args) {
  if (window['runtime']) {
    return window['runtime'].EventsOn(eventName, ...args);
  }
}

export function EventsOff(eventName, ...additionalEventNames) {
  if (window['runtime']) {
    return window['runtime'].EventsOff(eventName, ...additionalEventNames);
  }
}

export function EventsEmit(eventName, ...args) {
  if (window['runtime']) {
    return window['runtime'].EventsEmit(eventName, ...args);
  }
}

export function WindowMinimise() {
  if (window['runtime']) return window['runtime'].WindowMinimise();
}

export function WindowHide() {
  if (window['runtime']) return window['runtime'].WindowHide();
}

export function WindowShow() {
  if (window['runtime']) return window['runtime'].WindowShow();
}

export function ClipboardSetText(text) {
  if (window['runtime']) return window['runtime'].ClipboardSetText(text);
}

export function BrowserOpenURL(url) {
  if (window['runtime']) return window['runtime'].BrowserOpenURL(url);
}

export function Quit() {
  if (window['runtime']) return window['runtime'].Quit();
}
