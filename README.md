# Picui 图床上传工具

基于 **Wails v2 + Go + Vue3** 开发的 Windows 10/11 桌面端图床上传工具，对接 [Picui V2](https://picui.cn) 图床系统，支持 **双站点隔离**、拖拽/剪贴板/截图三种上传入口、异步并发队列、系统托盘常驻与全局快捷键，最终编译为 **单文件绿色 exe**（支持 x86 与 x64）。

> API 文档参考：https://97uklbao3b.apifox.cn/

---

## 功能特性

### 站点与配置
- 首次启动强制选择站点（`picui.cn` / `v2.picui.cn`），二选一后方可进入主界面
- 两套站点的 Token、相册、图片、上传历史 **完全隔离存储**，互不干扰
- 设置页一键切换站点，自动清空当前业务上下文并加载目标站点配置
- Token 配置与连通性校验（调用 `/api/v1/profile`）

### 上传核心
- **三种上传入口**：窗口拖拽上传、剪贴板图片粘贴上传（Ctrl+V）、屏幕选区截图上传
- **异步并发队列**：可配置最大并发数（1–10），超时自动重试（最多 3 次，指数退避）
- 实时进度与状态展示，失败任务支持手动重试
- **客户端压缩**：可选转 JPG/PNG、质量 10–100% 可调、最大宽度等比缩放
- 上传成功后自动复制 Markdown 链接到剪贴板
- 按站点隔离的上传历史记录

### 相册与图库
- 相册列表：分页加载、关键字搜索、排序（最新/最早/最多/最少）
- 相册新建 / 编辑 / 删除
- 图库浏览：分页、搜索、排序（最新/最早/最大/最小）、权限过滤
- 图片详情：复制 Markdown / URL / BBCode / 带链接 Markdown，浏览器打开，删除

### 系统交互
- 系统托盘常驻，关闭窗口默认最小化到托盘
- 托盘菜单：打开主窗口 / 剪贴板快速上传 / 截图上传 / 退出程序
- 全局快捷键（可自定义录制）：唤起主窗口、截图上传、剪贴板上传

### UI
- 简洁现代风格，支持明暗主题切换
- 上传列表清晰展示进度与状态，一键复制链接

---

## 技术栈

| 层 | 技术 |
|----|------|
| 桌面框架 | Wails v2（Go 后端 + WebView 前端，IPC 通信） |
| 后端 | Go 1.22+（HTTP 连接池、图片压缩、托盘、快捷键、截图、剪贴板） |
| 前端 | Vue 3 + Pinia + TypeScript + Vite |
| 系统集成 | `fyne.io/systray`（托盘）、`golang.design/x/hotkey`（快捷键）、`github.com/kbinani/screenshot`（截图）、`golang.org/x/sys`（剪贴板） |

> 本项目 **无 cgo 依赖**，所有平台库均使用纯 Go syscall 实现，可直接交叉编译 x86/x64。

---

## 目录结构

```
picui/
├── main.go                 # Wails 应用入口（窗口/拖拽配置/资源嵌入）
├── app.go                  # IPC 绑定层（暴露给前端的全部方法）
├── types.go                # 前后端共享类型定义
├── config.go               # 配置管理（按站点隔离持久化）
├── api.go                  # Picui API 客户端（连接池 + Bearer 鉴权）
├── upload.go               # 异步并发上传队列（重试/进度）
├── compress.go             # 客户端图片压缩（缩放/转码）
├── clipboard_windows.go    # Windows 剪贴板图片读取
├── screenshot.go           # 多显示器截图捕获
├── tray.go                 # 系统托盘
├── hotkey.go               # 全局快捷键
├── util.go                 # 工具函数
├── go.mod / wails.json     # Go 模块与 Wails 配置
├── build/windows/icon.ico  # 应用图标（6 尺寸）
├── build-x64.bat           # x64 编译脚本
├── build-x86.bat           # x86 编译脚本
├── build-all.bat           # 双架构批量编译
└── frontend/
    ├── package.json / vite.config.ts / tsconfig.json
    ├── index.html
    └── src/
        ├── main.ts                 # Vue 入口
        ├── App.vue                 # 根组件（首启/主界面路由）
        ├── store/index.ts          # Pinia 全局状态
        ├── style.css               # 全局样式与主题变量
        ├── wailsjs/                # Wails 绑定（手工同步至 app.go）
        │   ├── go/main/App.js / App.d.ts / models.ts
        │   └── runtime.js / runtime.d.ts
        └── components/
            ├── SiteSelect.vue       # 首启站点选择
            ├── Sidebar.vue          # 侧边栏导航
            ├── Upload.vue           # 上传页
            ├── History.vue          # 历史记录
            ├── Albums.vue           # 相册管理
            ├── Images.vue           # 图库浏览
            ├── Settings.vue         # 设置
            ├── ScreenshotOverlay.vue # 截图选区遮罩
            └── Toasts.vue           # 全局通知
```

---

## 环境准备

1. **Go** 1.22+：https://go.dev/dl/
2. **Node.js** 18+（含 npm）：https://nodejs.org/
3. **Wails CLI**（可选，推荐）：`go install github.com/wailsapp/wails/v2/cmd/wails@latest`
4. **WebView2 Runtime**：Windows 10/11 通常已预装；未预装时从 https://developer.microsoft.com/microsoft-edge/webview2/ 安装

---

## 编译命令

项目提供三个一键编译脚本，均会自动构建前端并输出单文件 exe：

### x64（windows/amd64）—— 推荐，主流 64 位系统

```bat
build-x64.bat
```

输出：`build\bin\PicuiUploader-x64.exe`（或 `PicuiUploader.exe`）

### x86（windows/386）—— 32 位系统

```bat
build-x86.bat
```

输出：`build\bin\PicuiUploader-x86.exe`（或 `PicuiUploader.exe`）

### 双架构批量编译

```bat
build-all.bat
```

### 编译方案说明

脚本优先使用 **wails build**（方案 A），未安装 Wails CLI 时自动回退到 **go build**（方案 B）：

**方案 A — wails build（推荐）**

```bash
wails build -platform windows/amd64 -clean -ldflags "-s -w -H windowsgui"
wails build -platform windows/386  -clean -ldflags "-s -w -H windowsgui"
```

- 自动构建前端、生成图标资源（`.syso`）、设置 GUI 子系统
- 输出含完整资源图标的单文件 exe

**方案 B — go build（回退）**

```bash
cd frontend && npm install && npm run build && cd ..
go install github.com/akavel/rsrc@latest
%USERPROFILE%\go\bin\rsrc.exe -ico build\windows\icon.ico -arch amd64 -o icon_windows.syso
set CGO_ENABLED=0
set GOOS=windows
set GOARCH=amd64
go build -ldflags "-s -w -H windowsgui" -o build\bin\PicuiUploader-x64.exe .
```

- `-H windowsgui`：隐藏控制台窗口，纯 GUI 程序
- `-s -w`：去除符号表与调试信息，减小体积
- `CGO_ENABLED=0`：纯 Go 编译，确保可交叉编译

> 编译后的 exe 为 **绿色单文件**，无需安装，拷贝即用。配置文件存储于 `%APPDATA%\PicuiUploader\config.json`。

---

## 开发模式

```bash
wails dev
```

启动热重载开发服务器，前端改动实时生效。

若未安装 Wails CLI，可分别启动：

```bash
cd frontend && npm run dev      # 前端开发服务器（另一终端）
go run .                         # Go 后端
```

---

## 使用指南

1. **首次启动**：选择站点（Picui 主站 / V2 站）
2. **配置 Token**：进入「设置」→ 填写 Token → 点击「校验」
3. **上传图片**：
   - 拖拽图片到上传区
   - 复制图片后按 Ctrl+V 或点击「剪贴板上传」
   - 点击「截图上传」框选屏幕区域
4. **复制链接**：上传成功后 Markdown 链接自动复制；也可在队列或历史中手动复制
5. **托盘与快捷键**：关闭窗口最小化到托盘；用全局快捷键快速唤起/截图/上传

---

## API 对接说明

所有接口遵循统一规范：

- **鉴权**：`Authorization: Bearer {Token}` Header
- **响应**：`{ status: boolean, message: string, data: object }`
- **分页**：`data` 含 `current_page`、`data[]`、`total`、`per_page`、`last_page`

已实现接口：

| 功能 | 方法 | 路径 |
|------|------|------|
| 用户资料 / Token 校验 | GET | `/api/v1/profile` |
| 相册列表 | GET | `/api/v1/albums` |
| 新建相册 | POST | `/user/albums` |
| 更新相册 | PUT | `/user/albums/{id}` |
| 删除相册 | DELETE | `/api/v1/albums/{id}` |
| 储存策略 | GET | `/api/v1/strategies` |
| 图片列表 | GET | `/api/v1/images` |
| 删除图片 | DELETE | `/api/v1/images/{key}` |
| 上传图片 | POST | `/api/v1/upload` |

支持的 Query 参数：
- 相册列表：`page`、`q`、`order`（newest/earliest/most/least）
- 图片列表：`page`、`q`、`order`（newest/earliest/utmost/least）、`permission`（public/private）、`album_id`

---

## 架构要点

- **站点隔离**：`ConfigManager` 以 `map[siteID]*SiteConfig` 持久化，所有 API 调用、上传、历史均绑定当前站点
- **上传队列**：worker pool + 信号量控制并发，独立协程不阻塞 UI，复用 `http.Transport` 连接池
- **IPC 双向通信**：Go 通过 `runtime.EventsEmit` 推送任务进度/截图数据；前端通过绑定方法调用 Go
- **跨平台原生能力**：托盘/快捷键/截图/剪贴板均用纯 Go 实现，无 cgo，支持 x86/x64 交叉编译

---

## 常见问题

**Q：编译后 exe 在资源管理器显示默认图标？**
A：使用 `wails build`（方案 A）会自动嵌入图标；`go build`（方案 B）需先用 `rsrc` 生成 `icon_windows.syso`（脚本已自动处理）。

**Q：截图上传框选时只能选窗口范围内？**
A：截图遮罩渲染于应用窗口内，多显示器画面会等比缩放显示后供框选，选区按比例映射回原始分辨率裁剪上传。

**Q：WebP 输出压缩？**
A：当前压缩输出格式为 JPG/PNG（纯 Go 无 cgo）。WebP 输入可正常解码上传；如需 WebP 输出可自行集成 cwebp。

---

## GitHub CI/CD 与发布

项目已内置 GitHub Actions 工作流 [`.github/workflows/build.yml`](.github/workflows/build.yml)，实现自动化构建与发布：

- **触发**：push 到 `main`、PR、打 `v*` tag、手动 `workflow_dispatch`
- **构建**：在 `windows-latest` 上分别编译 `windows/amd64` 与 `windows/386`，产物作为 Artifact 上传
- **发布**：打 tag（如 `v1.0.0`）时自动创建 GitHub Release，并附带 `PicuiUploader-x64.exe` / `PicuiUploader-x86.exe`

### 一键发布到 GitHub

前置：安装 [Git](https://git-scm.com/) 与 [GitHub CLI](https://cli.github.com/)，并完成 `gh auth login`。

```powershell
# 默认 public 仓库
./setup-github.ps1 -RepoName picui-uploader -Visibility public

# 或私有仓库
./setup-github.ps1 -RepoName picui-uploader -Visibility private
```

脚本会自动：初始化 git → 提交代码 → 创建 GitHub 仓库 → 推送 → 配置 `main` 分支保护（要求 CI 通过）。

### 发布 Release

```bash
git tag v1.0.0
git push origin v1.0.0
# Actions 自动构建并发布 Release，附带双架构 exe
```

### 手动流程（不使用脚本）

```bash
git init && git branch -M main
git add -A && git commit -m "feat: 初始化 Picui 图床上传工具"
gh repo create picui-image-uploader --public --source=. --remote=origin --push
# 打 tag 发布
git tag v1.0.0 && git push origin v1.0.0
```

---

## 许可

本项目仅供学习与个人使用，请遵守 Picui 图床的服务条款。
