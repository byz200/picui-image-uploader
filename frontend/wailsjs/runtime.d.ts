// C:\Users\Administrator\Documents\trae_projects\picui\frontend\src\wailsjs\runtime.d.ts - Wails 运行时桥接类型声明
export declare function EventsOn(eventName: string, cb: (...data: any) => void): void;
export declare function EventsOff(eventName: string, ...additionalEventNames: string[]): void;
export declare function EventsEmit(eventName: string, ...data: any[]): void;
export declare function WindowMinimise(): void;
export declare function WindowHide(): void;
export declare function WindowShow(): void;
export declare function ClipboardSetText(text: string): Promise<void>;
export declare function BrowserOpenURL(url: string): void;
export declare function Quit(): void;
