// C:\Users\Administrator\Documents\trae_projects\picui\frontend\src\wailsjs\go\main\App.d.ts - Wails 自动生成绑定声明（手工同步至 app.go）

export namespace main {
  export interface AppState {
    firstRun: boolean;
    currentSite: string;
    sites: SiteInfo[];
    settings: Settings;
  }

  export interface SiteInfo {
    id: string;
    name: string;
    baseUrl: string;
    hasToken: boolean;
  }

  export interface Settings {
    siteId: string;
    token: string;
    theme: string;
    maxConcurrency: number;
    autoCopyMarkdown: boolean;
    defaultPermission: number;
    defaultStrategyId: string;
    defaultAlbumId: string;
    compress: boolean;
    compressFormat: string;
    compressQuality: number;
    maxWidth: number;
    hotkeyShowWindow: string;
    hotkeyScreenshot: string;
    hotkeyClipboard: string;
    minimizeToTray: boolean;
  }

  export interface LoginRequest {
    loginType: string;
    username: string;
    password: string;
    remember: boolean;
    countryCode: string;
  }

  export interface LoginResult {
    name: string;
    token: string;
  }

  export interface Profile {
    avatar: string;
    name: string;
    username: string;
    email: string;
    imageNum: number;
    albumNum: number;
    registeredIp: string;
    url: string;
    capacity: number;
    size: number;
  }

  export interface Album {
    id: number;
    name: string;
    intro: string;
    imageNum: number;
  }

  export interface AlbumList {
    currentPage: number;
    lastPage: number;
    perPage: number;
    total: number;
    data: Album[];
  }

  export interface CreateAlbumResult {
    id: number;
  }

  export interface Strategy {
    id: any;
    name: string;
    intro: string;
    isDefault: boolean;
  }

  export interface ImageLinks {
    url: string;
    html: string;
    bbcode: string;
    markdown: string;
    markdownWithLink: string;
    thumbnailUrl: string;
    deleteUrl: string;
  }

  export interface ImageItem {
    key: number;
    name: string;
    pathname: string;
    mimetype: string;
    extension: string;
    size: number;
    width: number;
    height: number;
    humanDate: string;
    date: string;
    links: ImageLinks;
  }

  export interface ImageList {
    currentPage: number;
    lastPage: number;
    perPage: number;
    total: number;
    data: ImageItem[];
  }

  export interface UploadOptions {
    albumId: string;
    strategyId: string;
    permission: number;
    compress: boolean;
    compressFormat: string;
    compressQuality: number;
    maxWidth: number;
  }

  export interface UploadTask {
    id: string;
    name: string;
    size: number;
    status: string;
    progress: number;
    error: string;
    url: string;
    markdown: string;
    createdAt: number;
    retries: number;
    siteId: string;
  }

  export interface HistoryPage {
    total: number;
    page: number;
    data: HistoryItem[];
  }

  export interface HistoryItem {
    id: string;
    name: string;
    size: number;
    url: string;
    markdown: string;
    createdAt: number;
    siteId: string;
  }
}

// ============================ App 绑定方法 ============================

export declare function GetAppState(): Promise<main.AppState>;

export declare function SelectSite(siteID: string): Promise<void>;

export declare function SwitchSite(siteID: string): Promise<void>;

export declare function GetSettings(): Promise<main.Settings>;

export declare function SaveSettings(s: main.Settings): Promise<void>;

export declare function TestToken(token: string): Promise<main.Profile>;

export declare function Login(req: main.LoginRequest): Promise<main.LoginResult>;

export declare function GetAlbums(
  page: number,
  q: string,
  order: string
): Promise<main.AlbumList>;

export declare function CreateAlbum(
  name: string,
  intro: string,
  isPublic: boolean
): Promise<main.CreateAlbumResult>;

export declare function UpdateAlbum(
  id: number,
  name: string,
  intro: string,
  isPublic: boolean
): Promise<void>;

export declare function DeleteAlbum(id: number): Promise<void>;

export declare function GetStrategies(): Promise<Array<main.Strategy>>;

export declare function GetImages(
  page: number,
  q: string,
  order: string,
  permission: string,
  albumID: string
): Promise<main.ImageList>;

export declare function DeleteImage(key: number): Promise<void>;

export declare function UploadFiles(
  paths: Array<string>,
  opts: main.UploadOptions
): Promise<Array<string>>;

export declare function UploadClipboard(opts: main.UploadOptions): Promise<string>;

export declare function UploadBase64(
  dataURL: string,
  filename: string,
  opts: main.UploadOptions
): Promise<string>;

export declare function RetryUpload(id: string): Promise<void>;

export declare function RemoveUploadTask(id: string): Promise<void>;

export declare function ClearUploadTasks(): Promise<void>;

export declare function GetUploadTasks(): Promise<Array<main.UploadTask>>;

export declare function StartScreenshot(): Promise<void>;

export declare function GetHistory(page: number, pageSize: number): Promise<main.HistoryPage>;

export declare function DeleteHistory(id: string): Promise<void>;

export declare function ClearHistory(): Promise<void>;

export declare function MinimizeToTray(): Promise<void>;

export declare function QuitApp(): Promise<void>;

export declare function BringToFront(): Promise<void>;

export declare function CopyText(text: string): Promise<void>;

export declare function OpenURL(url: string): Promise<void>;

export declare function SetTheme(theme: string): Promise<void>;
