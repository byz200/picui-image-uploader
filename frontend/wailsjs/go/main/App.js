// C:\Users\Administrator\Documents\trae_projects\picui\frontend\src\wailsjs\go\main\App.js - Wails 自动生成绑定（手工同步至 app.go）
// @ts-check
// eslint-disable-next-line no-undef
export function GetAppState() {
  return window['go']['main']['App']['GetAppState']();
}

export function SelectSite(siteID) {
  return window['go']['main']['App']['SelectSite'](siteID);
}

export function SwitchSite(siteID) {
  return window['go']['main']['App']['SwitchSite'](siteID);
}

export function GetSettings() {
  return window['go']['main']['App']['GetSettings']();
}

export function SaveSettings(s) {
  return window['go']['main']['App']['SaveSettings'](s);
}

export function TestToken(token) {
  return window['go']['main']['App']['TestToken'](token);
}

export function Login(req) {
  return window['go']['main']['App']['Login'](req);
}

export function GetAlbums(page, q, order) {
  return window['go']['main']['App']['GetAlbums'](page, q, order);
}

export function CreateAlbum(name, intro, isPublic) {
  return window['go']['main']['App']['CreateAlbum'](name, intro, isPublic);
}

export function UpdateAlbum(id, name, intro, isPublic) {
  return window['go']['main']['App']['UpdateAlbum'](id, name, intro, isPublic);
}

export function DeleteAlbum(id) {
  return window['go']['main']['App']['DeleteAlbum'](id);
}

export function GetStrategies() {
  return window['go']['main']['App']['GetStrategies']();
}

export function GetImages(page, q, order, permission, albumID) {
  return window['go']['main']['App']['GetImages'](page, q, order, permission, albumID);
}

export function DeleteImage(key) {
  return window['go']['main']['App']['DeleteImage'](key);
}

export function UploadFiles(paths, opts) {
  return window['go']['main']['App']['UploadFiles'](paths, opts);
}

export function UploadClipboard(opts) {
  return window['go']['main']['App']['UploadClipboard'](opts);
}

export function UploadBase64(dataURL, filename, opts) {
  return window['go']['main']['App']['UploadBase64'](dataURL, filename, opts);
}

export function RetryUpload(id) {
  return window['go']['main']['App']['RetryUpload'](id);
}

export function RemoveUploadTask(id) {
  return window['go']['main']['App']['RemoveUploadTask'](id);
}

export function ClearUploadTasks() {
  return window['go']['main']['App']['ClearUploadTasks']();
}

export function GetUploadTasks() {
  return window['go']['main']['App']['GetUploadTasks']();
}

export function StartScreenshot() {
  return window['go']['main']['App']['StartScreenshot']();
}

export function GetHistory(page, pageSize) {
  return window['go']['main']['App']['GetHistory'](page, pageSize);
}

export function DeleteHistory(id) {
  return window['go']['main']['App']['DeleteHistory'](id);
}

export function ClearHistory() {
  return window['go']['main']['App']['ClearHistory']();
}

export function MinimizeToTray() {
  return window['go']['main']['App']['MinimizeToTray']();
}

export function QuitApp() {
  return window['go']['main']['App']['QuitApp']();
}

export function BringToFront() {
  return window['go']['main']['App']['BringToFront']();
}

export function CopyText(text) {
  return window['go']['main']['App']['CopyText'](text);
}

export function OpenURL(url) {
  return window['go']['main']['App']['OpenURL'](url);
}

export function SetTheme(theme) {
  return window['go']['main']['App']['SetTheme'](theme);
}
