// C:\Users\Administrator\Documents\trae_projects\picui\frontend\src\wailsjs\go\models.ts - Wails 自动生成模型导出
export namespace main {
  export class AppState {
    firstRun: boolean = false;
    currentSite: string = '';
    sites: SiteInfo[] = [];
    settings: Settings = new Settings();
  }

  export class SiteInfo {
    id: string = '';
    name: string = '';
    baseUrl: string = '';
    hasToken: boolean = false;
  }

  export class Settings {
    siteId: string = '';
    token: string = '';
    theme: string = '';
    maxConcurrency: number = 0;
    autoCopyMarkdown: boolean = false;
    defaultPermission: number = 0;
    defaultStrategyId: string = '';
    defaultAlbumId: string = '';
    compress: boolean = false;
    compressFormat: string = '';
    compressQuality: number = 0;
    maxWidth: number = 0;
    hotkeyShowWindow: string = '';
    hotkeyScreenshot: string = '';
    hotkeyClipboard: string = '';
    minimizeToTray: boolean = false;
  }
}
