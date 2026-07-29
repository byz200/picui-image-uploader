// store/index.ts - 全局应用状态（Pinia）
import { defineStore } from 'pinia'
import { reactive, ref, computed } from 'vue'
import * as App from '../../wailsjs/go/main/App'
import * as runtime from '../../wailsjs/runtime'
import type * as T from '../../wailsjs/go/main/App'

export type ToastType = 'success' | 'error' | 'info'
export interface Toast {
  id: number
  type: ToastType
  text: string
}

export type ViewName = 'upload' | 'history' | 'albums' | 'images' | 'settings'

// 默认上传选项（与后端 applyDefaults 配合，空值由后端补齐）
function defaultUploadOpts(): T.main.UploadOptions {
  return {
    albumId: '',
    strategyId: '',
    permission: 0,
    compress: false,
    compressFormat: '',
    compressQuality: 0,
    maxWidth: 0
  }
}

export const useStore = defineStore('app', () => {
  // ---- 启动状态 ----
  const firstRun = ref(false)
  const currentSite = ref('')
  const sites = ref<T.main.SiteInfo[]>([])
  const settings = ref<T.main.Settings | null>(null)
  const ready = ref(false)

  // ---- 视图路由 ----
  const view = ref<ViewName>('upload')

  // ---- 上传任务 ----
  const tasks = ref<T.main.UploadTask[]>([])

  // ---- 截图选区遮罩 ----
  const screenshotPayload = ref<any>(null)
  const screenshotActive = ref(false)

  // ---- Toast ----
  const toasts = ref<Toast[]>([])
  let toastSeq = 0

  function toast(text: string, type: ToastType = 'info', duration = 2500) {
    const id = ++toastSeq
    toasts.value.push({ id, type, text })
    setTimeout(() => removeToast(id), duration)
  }

  function removeToast(id: number) {
    toasts.value = toasts.value.filter((t) => t.id !== id)
  }

  // ---- 当前站点信息 ----
  const currentSiteInfo = computed<T.main.SiteInfo | undefined>(() =>
    sites.value.find((s) => s.id === currentSite.value)
  )

  const hasToken = computed(() => !!settings.value?.token)

  // ---- 主题 ----
  function applyTheme(theme: string) {
    let resolved = theme
    if (theme === 'system' || !theme) {
      resolved = 'light'
    }
    document.documentElement.setAttribute('data-theme', resolved)
  }

  // ---- 初始化 ----
  async function init() {
    try {
      const state = await App.GetAppState()
      firstRun.value = state.firstRun
      currentSite.value = state.currentSite
      sites.value = state.sites || []
      settings.value = state.settings
      ready.value = true
      if (settings.value) applyTheme(settings.value.theme)
      // 拉取现有上传任务快照
      try {
        tasks.value = await App.GetUploadTasks()
      } catch {
        tasks.value = []
      }
    } catch (e) {
      ready.value = true
      toast('初始化失败：' + String(e), 'error')
    }
  }

  // ---- 站点选择/切换 ----
  async function selectSite(siteID: string) {
    await App.SelectSite(siteID)
    currentSite.value = siteID
    firstRun.value = false
    settings.value = await App.GetSettings()
    if (settings.value) applyTheme(settings.value.theme)
  }

  async function switchSite(siteID: string) {
    await App.SwitchSite(siteID)
    currentSite.value = siteID
    settings.value = await App.GetSettings()
    tasks.value = []
    toast(`已切换到 ${siteName(siteID)}`, 'success')
  }

  function siteName(id: string): string {
    return sites.value.find((s) => s.id === id)?.name || id
  }

  // ---- 设置 ----
  async function saveSettings(s: T.main.Settings) {
    await App.SaveSettings(s)
    settings.value = { ...s }
    applyTheme(s.theme)
    toast('设置已保存', 'success')
  }

  async function setTheme(theme: string) {
    await App.SetTheme(theme)
    applyTheme(theme)
    if (settings.value) settings.value.theme = theme
  }

  // ---- 上传操作 ----
  // buildOpts: 合并当前站点设置形成完整的上传选项
  function buildOpts(overrides: Partial<T.main.UploadOptions> = {}): T.main.UploadOptions {
    const s = settings.value
    const base = defaultUploadOpts()
    if (s) {
      base.albumId = s.defaultAlbumId
      base.strategyId = s.defaultStrategyId
      base.permission = s.defaultPermission || 1
      base.compress = s.compress
      base.compressFormat = s.compressFormat
      base.compressQuality = s.compressQuality
      base.maxWidth = s.maxWidth
    }
    return { ...base, ...overrides }
  }

  async function uploadFiles(paths: string[], overrides: Partial<T.main.UploadOptions> = {}) {
    const opts = buildOpts(overrides)
    return await App.UploadFiles(paths, opts)
  }

  async function uploadClipboard(overrides: Partial<T.main.UploadOptions> = {}) {
    const opts = buildOpts(overrides)
    try {
      const id = await App.UploadClipboard(opts)
      if (!id) toast('剪贴板中没有图片', 'info')
      return id
    } catch (e) {
      toast('剪贴板上传失败：' + String(e), 'error')
      return ''
    }
  }

  async function uploadBase64(
    dataURL: string,
    filename: string,
    overrides: Partial<T.main.UploadOptions> = {}
  ) {
    const opts = buildOpts(overrides)
    return await App.UploadBase64(dataURL, filename, opts)
  }

  async function retryUpload(id: string) {
    try {
      await App.RetryUpload(id)
    } catch (e) {
      toast('重试失败：' + String(e), 'error')
    }
  }

  async function removeUploadTask(id: string) {
    try {
      await App.RemoveUploadTask(id)
      tasks.value = tasks.value.filter((t) => t.id !== id)
    } catch (e) {
      toast('删除失败：' + String(e), 'error')
    }
  }

  async function clearUploadTasks() {
    try {
      await App.ClearUploadTasks()
      tasks.value = []
    } catch (e) {
      toast('清空失败：' + String(e), 'error')
    }
  }

  // ---- 截图 ----
  async function startScreenshot() {
    try {
      await App.StartScreenshot()
    } catch (e) {
      toast('截图启动失败：' + String(e), 'error')
    }
  }

  function showScreenshotOverlay(payload: any) {
    screenshotPayload.value = payload
    screenshotActive.value = true
  }

  function hideScreenshotOverlay() {
    screenshotActive.value = false
    screenshotPayload.value = null
  }

  // ---- 复制 ----
  async function copyText(text: string) {
    try {
      await App.CopyText(text)
      toast('已复制到剪贴板', 'success', 1500)
    } catch (e) {
      toast('复制失败：' + String(e), 'error')
    }
  }

  // ---- 系统交互 ----
  async function minimizeToTray() {
    await App.MinimizeToTray()
  }

  async function quitApp() {
    await App.QuitApp()
  }

  async function openURL(url: string) {
    try {
      await App.OpenURL(url)
    } catch {
      // ignore
    }
  }

  // ---- 事件监听 ----
  function registerEvents() {
    // 上传任务更新
    runtime.EventsOn('upload:task', (task: T.main.UploadTask) => {
      const idx = tasks.value.findIndex((t) => t.id === task.id)
      if (idx >= 0) {
        tasks.value[idx] = task
      } else {
        tasks.value.push(task)
      }
      // 任务成功且自动复制 Markdown 由后端完成；此处仅做提示
      if (task.status === 'success' && task.url) {
        // 不重复 toast，避免批量上传刷屏
      }
    })

    runtime.EventsOn('upload:progress', (payload: { id: string; progress: number }) => {
      const t = tasks.value.find((x) => x.id === payload.id)
      if (t) t.progress = payload.progress
    })

    runtime.EventsOn('upload:removed', (id: string) => {
      tasks.value = tasks.value.filter((t) => t.id !== id)
    })

    runtime.EventsOn('upload:cleared', () => {
      tasks.value = []
    })

    // 拖拽文件
    runtime.EventsOn('drop:files', (paths: string[]) => {
      if (paths && paths.length > 0) {
        uploadFiles(paths)
        view.value = 'upload'
        toast(`已添加 ${paths.length} 个文件到上传队列`, 'info')
      }
    })

    // 站点切换（来自托盘或设置页）
    runtime.EventsOn('site:changed', async (siteID: string) => {
      currentSite.value = siteID
      settings.value = await App.GetSettings()
      if (settings.value) applyTheme(settings.value.theme)
    })

    // 设置保存
    runtime.EventsOn('settings:saved', (s: T.main.Settings) => {
      settings.value = s
      applyTheme(s.theme)
    })

    // 主题切换
    runtime.EventsOn('theme:changed', (theme: string) => {
      applyTheme(theme)
      if (settings.value) settings.value.theme = theme
    })

    // 截图数据就绪
    runtime.EventsOn('screenshot:ready', (payload: any) => {
      showScreenshotOverlay(payload)
    })
  }

  return {
    // state
    firstRun,
    currentSite,
    sites,
    settings,
    ready,
    view,
    tasks,
    screenshotPayload,
    screenshotActive,
    toasts,
    // computed
    currentSiteInfo,
    hasToken,
    // actions
    init,
    selectSite,
    switchSite,
    siteName,
    saveSettings,
    setTheme,
    buildOpts,
    uploadFiles,
    uploadClipboard,
    uploadBase64,
    retryUpload,
    removeUploadTask,
    clearUploadTasks,
    startScreenshot,
    showScreenshotOverlay,
    hideScreenshotOverlay,
    copyText,
    minimizeToTray,
    quitApp,
    openURL,
    registerEvents,
    toast,
    removeToast,
    applyTheme
  }
})
