<script setup lang="ts">
// Settings.vue - 设置页：站点切换 / Token / 主题 / 上传 / 压缩 / 快捷键 / 托盘
import { ref, reactive, computed, onMounted, onBeforeUnmount } from 'vue'
import { useStore } from '../store'
import * as App from '../../wailsjs/go/main/App'
import type * as T from '../../wailsjs/go/main/App'

const store = useStore()

// 表单副本
const form = reactive<T.main.Settings>({
  siteId: '',
  token: '',
  theme: 'light',
  maxConcurrency: 3,
  autoCopyMarkdown: true,
  defaultPermission: 1,
  defaultStrategyId: '',
  defaultAlbumId: '',
  compress: false,
  compressFormat: 'original',
  compressQuality: 85,
  maxWidth: 0,
  hotkeyShowWindow: 'Ctrl+Shift+P',
  hotkeyScreenshot: 'Ctrl+Shift+A',
  hotkeyClipboard: 'Ctrl+Shift+C',
  minimizeToTray: true
})

const saving = ref(false)

// ---- 相册 / 策略选项 ----
const albums = ref<T.main.Album[]>([])
const strategies = ref<T.main.Strategy[]>([])

async function loadOptions() {
  try {
    const [al, st] = await Promise.all([
      App.GetAlbums(1, '', 'newest').catch(() => ({
        currentPage: 1,
        lastPage: 1,
        perPage: 20,
        total: 0,
        data: [] as T.main.Album[]
      } satisfies T.main.AlbumList)),
      App.GetStrategies().catch(() => [] as T.main.Strategy[])
    ])
    albums.value = al.data || []
    strategies.value = st || []
  } catch {
    // ignore
  }
}

// ---- 站点切换 ----
async function switchSite(siteID: string) {
  if (siteID === form.siteId) return
  if (
    !confirm(
      `切换到「${store.siteName(siteID)}」将清空当前上传队列与业务上下文，并加载目标站点配置。是否继续？`
    )
  )
    return
  try {
    await store.switchSite(siteID)
    syncFromStore()
    await loadOptions()
  } catch (e) {
    store.toast('切换站点失败：' + String(e), 'error')
  }
}

// ---- Token 校验 ----
const testing = ref(false)
const profile = ref<T.main.Profile | null>(null)

async function testToken() {
  if (!form.token.trim()) {
    store.toast('请先填写 Token', 'error')
    return
  }
  testing.value = true
  profile.value = null
  try {
    profile.value = await App.TestToken(form.token.trim())
    store.toast('Token 校验通过', 'success')
  } catch (e) {
    store.toast('Token 校验失败：' + String(e), 'error')
  } finally {
    testing.value = false
  }
}

// ---- 保存 ----
async function save() {
  saving.value = true
  try {
    await store.saveSettings({ ...form })
  } catch (e) {
    store.toast('保存失败：' + String(e), 'error')
  } finally {
    saving.value = false
  }
}

// ---- 主题切换 ----
function setTheme(t: string) {
  form.theme = t
  store.applyTheme(t)
}

// ---- 快捷键捕获 ----
const capturingField = ref<'' | 'hotkeyShowWindow' | 'hotkeyScreenshot' | 'hotkeyClipboard'>('')

function startCapture(field: typeof capturingField.value) {
  capturingField.value = field
}

function stopCapture() {
  capturingField.value = ''
}

function comboFromEvent(e: KeyboardEvent): string {
  const mods: string[] = []
  if (e.ctrlKey) mods.push('Ctrl')
  if (e.shiftKey) mods.push('Shift')
  if (e.altKey) mods.push('Alt')
  let key = e.key
  // 规范化
  const map: Record<string, string> = {
    Control: '',
    Shift: '',
    Alt: '',
    Meta: '',
    ' ': 'Space',
    Enter: 'Enter',
    Escape: 'Esc',
    Tab: 'Tab',
    ArrowUp: 'Up',
    ArrowDown: 'Down',
    ArrowLeft: 'Left',
    ArrowRight: 'Right'
  }
  if (map[key] !== undefined) key = map[key]
  if (!key) return ''
  if (key.length === 1) key = key.toUpperCase()
  // 仅支持字母/数字/功能键
  if (!/^[A-Z0-9]$/.test(key) && !/^F\d{1,2}$/.test(key) && !['Space', 'Enter', 'Esc', 'Tab'].includes(key)) {
    return ''
  }
  if (mods.length === 0) return ''
  return [...mods, key].join('+')
}

function onKeyCapture(e: KeyboardEvent) {
  if (!capturingField.value) return
  e.preventDefault()
  e.stopPropagation()
  if (e.key === 'Escape') {
    stopCapture()
    return
  }
  const combo = comboFromEvent(e)
  if (combo) {
    ;(form as any)[capturingField.value] = combo
    stopCapture()
  }
}

// ---- 同步 ----
function syncFromStore() {
  if (!store.settings) return
  const s = store.settings
  Object.assign(form, s)
}

onMounted(() => {
  syncFromStore()
  loadOptions()
  window.addEventListener('keydown', onKeyCapture, true)
})

onBeforeUnmount(() => {
  window.removeEventListener('keydown', onKeyCapture, true)
})

const themes = [
  { value: 'light', label: '明亮', icon: '☀️' },
  { value: 'dark', label: '暗黑', icon: '🌙' }
]
</script>

<template>
  <div class="settings-page">
    <div class="page-header">
      <div>
        <h2 class="page-title">设置</h2>
        <p class="page-sub">当前站点：{{ store.currentSiteInfo?.name }}（{{ store.currentSiteInfo?.baseUrl }}）</p>
      </div>
      <button class="btn btn-primary" :disabled="saving" @click="save">
        <span v-if="saving" class="spinner"></span>
        <span>保存设置</span>
      </button>
    </div>

    <div class="set-scroll">
      <!-- 站点切换 -->
      <section class="card set-section">
        <h3 class="set-section-title">站点切换</h3>
        <p class="set-section-desc">切换后自动清空当前业务上下文，加载目标站点配置与历史。</p>
        <div class="set-site-list">
          <div
            v-for="site in store.sites"
            :key="site.id"
            class="set-site"
            :class="{ active: form.siteId === site.id }"
            @click="switchSite(site.id)"
          >
            <div class="set-site-icon">{{ site.id === 'picui' ? '🌐' : '🚀' }}</div>
            <div class="set-site-info">
              <div class="set-site-name">{{ site.name }}</div>
              <div class="set-site-url">{{ site.baseUrl }}</div>
            </div>
            <span v-if="form.siteId === site.id" class="badge badge-success">当前</span>
          </div>
        </div>
      </section>

      <!-- Token 配置 -->
      <section class="card set-section">
        <h3 class="set-section-title">Token 配置</h3>
        <p class="set-section-desc">在 Header 中以 <code>Authorization: Bearer {Token}</code> 形式携带。</p>
        <div class="set-row">
          <label class="set-label">访问 Token</label>
          <div class="set-token-input">
            <input
              class="input"
              v-model="form.token"
              type="password"
              placeholder="粘贴你的 Picui API Token"
            />
            <button class="btn btn-outline" :disabled="testing" @click="testToken">
              <span v-if="testing" class="spinner"></span>
              <span>校验</span>
            </button>
          </div>
        </div>
        <div v-if="profile" class="set-profile">
          <div class="set-profile-name">{{ profile.name || profile.username }}</div>
          <div class="set-profile-meta">
            <span v-if="profile.email">{{ profile.email }}</span>
            <span> · 图片 {{ profile.imageNum }}</span>
            <span> · 相册 {{ profile.albumNum }}</span>
          </div>
        </div>
      </section>

      <!-- 外观 -->
      <section class="card set-section">
        <h3 class="set-section-title">外观</h3>
        <div class="set-row">
          <label class="set-label">主题</label>
          <div class="set-theme-list">
            <button
              v-for="t in themes"
              :key="t.value"
              class="set-theme"
              :class="{ active: form.theme === t.value }"
              @click="setTheme(t.value)"
            >
              <span>{{ t.icon }}</span>
              <span>{{ t.label }}</span>
            </button>
          </div>
        </div>
        <div class="set-row">
          <label class="set-label">关闭窗口时</label>
          <label class="up-checkbox">
            <input type="checkbox" v-model="form.minimizeToTray" />
            <span>最小化到系统托盘（取消则直接退出）</span>
          </label>
        </div>
      </section>

      <!-- 上传默认值 -->
      <section class="card set-section">
        <h3 class="set-section-title">上传默认值</h3>
        <div class="set-grid">
          <div class="set-row">
            <label class="set-label">最大并发数</label>
            <input
              class="input set-num"
              type="number"
              min="1"
              max="10"
              v-model.number="form.maxConcurrency"
            />
          </div>
          <div class="set-row">
            <label class="set-label">默认权限</label>
            <select class="select" v-model.number="form.defaultPermission">
              <option :value="1">公开</option>
              <option :value="0">私有</option>
            </select>
          </div>
          <div class="set-row">
            <label class="set-label">默认相册</label>
            <select class="select" v-model="form.defaultAlbumId">
              <option value="">无</option>
              <option v-for="a in albums" :key="a.id" :value="String(a.id)">{{ a.name }}</option>
            </select>
          </div>
          <div class="set-row">
            <label class="set-label">默认储存策略</label>
            <select class="select" v-model="form.defaultStrategyId">
              <option value="">默认</option>
              <option v-for="s in strategies" :key="String(s.id)" :value="String(s.id)">{{ s.name }}</option>
            </select>
          </div>
        </div>
        <div class="set-row">
          <label class="set-label">上传成功后</label>
          <label class="up-checkbox">
            <input type="checkbox" v-model="form.autoCopyMarkdown" />
            <span>自动复制 Markdown 链接到剪贴板</span>
          </label>
        </div>
      </section>

      <!-- 客户端压缩 -->
      <section class="card set-section">
        <h3 class="set-section-title">客户端压缩</h3>
        <div class="set-row">
          <label class="set-label">启用压缩</label>
          <label class="up-checkbox">
            <input type="checkbox" v-model="form.compress" />
            <span>上传前在本地压缩图片</span>
          </label>
        </div>
        <div class="set-grid">
          <div class="set-row">
            <label class="set-label">压缩格式</label>
            <select class="select" v-model="form.compressFormat" :disabled="!form.compress">
              <option value="original">保持原格式</option>
              <option value="jpg">转 JPG</option>
              <option value="png">转 PNG</option>
            </select>
          </div>
          <div class="set-row">
            <label class="set-label">质量 ({{ form.compressQuality }}%)</label>
            <input
              class="set-range"
              type="range"
              min="10"
              max="100"
              v-model.number="form.compressQuality"
              :disabled="!form.compress || form.compressFormat === 'png'"
            />
          </div>
          <div class="set-row">
            <label class="set-label">最大宽度 (px)</label>
            <input
              class="input set-num"
              type="number"
              min="0"
              max="20000"
              v-model.number="form.maxWidth"
              placeholder="0 表示不缩放"
              :disabled="!form.compress"
            />
          </div>
        </div>
        <p class="set-section-desc">注：WebP「输入」可解码上传；WebP「输出」编码需 cwebp，当前输出格式仅 JPG/PNG。</p>
      </section>

      <!-- 全局快捷键 -->
      <section class="card set-section">
        <h3 class="set-section-title">全局快捷键</h3>
        <p class="set-section-desc">点击右侧按钮后按下组合键进行录制（需包含 Ctrl/Shift/Alt 至少一个修饰键）。</p>
        <div class="set-row">
          <label class="set-label">唤起主窗口</label>
          <div class="set-hotkey">
            <span class="set-hotkey-display">{{ form.hotkeyShowWindow || '未设置' }}</span>
            <button
              class="btn btn-outline btn-sm"
              @click="startCapture('hotkeyShowWindow')"
              @blur="stopCapture"
            >
              {{ capturingField === 'hotkeyShowWindow' ? '按下组合键…' : '录制' }}
            </button>
          </div>
        </div>
        <div class="set-row">
          <label class="set-label">截图上传</label>
          <div class="set-hotkey">
            <span class="set-hotkey-display">{{ form.hotkeyScreenshot || '未设置' }}</span>
            <button
              class="btn btn-outline btn-sm"
              @click="startCapture('hotkeyScreenshot')"
              @blur="stopCapture"
            >
              {{ capturingField === 'hotkeyScreenshot' ? '按下组合键…' : '录制' }}
            </button>
          </div>
        </div>
        <div class="set-row">
          <label class="set-label">剪贴板上传</label>
          <div class="set-hotkey">
            <span class="set-hotkey-display">{{ form.hotkeyClipboard || '未设置' }}</span>
            <button
              class="btn btn-outline btn-sm"
              @click="startCapture('hotkeyClipboard')"
              @blur="stopCapture"
            >
              {{ capturingField === 'hotkeyClipboard' ? '按下组合键…' : '录制' }}
            </button>
          </div>
        </div>
      </section>
    </div>
  </div>
</template>

<style scoped>
.settings-page {
  padding: 20px;
  display: flex;
  flex-direction: column;
  gap: 16px;
  height: 100%;
}

.page-header {
  display: flex;
  align-items: flex-end;
  justify-content: space-between;
}

.page-title {
  font-size: 18px;
  font-weight: 600;
  color: var(--text);
  margin-bottom: 4px;
}

.page-sub {
  font-size: 12px;
  color: var(--text-muted);
}

.set-scroll {
  flex: 1;
  overflow-y: auto;
  display: flex;
  flex-direction: column;
  gap: 16px;
  padding-right: 4px;
}

.set-section {
  padding: 18px 20px;
}

.set-section-title {
  font-size: 15px;
  font-weight: 600;
  color: var(--text);
  margin-bottom: 6px;
}

.set-section-desc {
  font-size: 12px;
  color: var(--text-muted);
  margin-bottom: 14px;
  line-height: 1.5;
}

.set-section-desc code {
  background: var(--bg-hover);
  padding: 1px 5px;
  border-radius: 3px;
  font-size: 11px;
  font-family: 'Consolas', 'Monaco', monospace;
}

.set-site-list {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.set-site {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px 14px;
  border: 1px solid var(--border);
  border-radius: var(--radius-sm);
  cursor: pointer;
  transition: all 0.12s ease;
}

.set-site:hover {
  border-color: var(--primary);
  background: var(--bg-hover);
}

.set-site.active {
  border-color: var(--primary);
  background: var(--primary-soft);
}

.set-site-icon {
  font-size: 24px;
}

.set-site-info {
  flex: 1;
}

.set-site-name {
  font-size: 14px;
  font-weight: 600;
  color: var(--text);
}

.set-site-url {
  font-size: 12px;
  color: var(--text-muted);
  font-family: 'Consolas', 'Monaco', monospace;
}

.set-row {
  display: flex;
  align-items: center;
  gap: 14px;
  margin-bottom: 12px;
}

.set-row:last-child {
  margin-bottom: 0;
}

.set-label {
  width: 130px;
  flex-shrink: 0;
  font-size: 13px;
  color: var(--text-secondary);
}

.set-token-input {
  display: flex;
  gap: 8px;
  flex: 1;
  max-width: 460px;
}

.set-profile {
  margin-top: 10px;
  padding: 10px 12px;
  background: var(--success-soft);
  border-radius: var(--radius-sm);
}

.set-profile-name {
  font-size: 14px;
  font-weight: 600;
  color: var(--success);
}

.set-profile-meta {
  font-size: 12px;
  color: var(--text-secondary);
  margin-top: 2px;
}

.set-theme-list {
  display: flex;
  gap: 8px;
}

.set-theme {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 8px 16px;
  border: 1px solid var(--border);
  border-radius: var(--radius-sm);
  background: var(--bg-elevated);
  color: var(--text-secondary);
  font-size: 13px;
}

.set-theme.active {
  border-color: var(--primary);
  color: var(--primary);
  background: var(--primary-soft);
}

.set-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(240px, 1fr));
  gap: 12px 20px;
  margin-bottom: 12px;
}

.set-grid .set-row {
  margin-bottom: 0;
}

.set-grid .set-label {
  width: auto;
}

.set-num {
  max-width: 140px;
}

.set-range {
  flex: 1;
  max-width: 240px;
  cursor: pointer;
}

.set-hotkey {
  display: flex;
  align-items: center;
  gap: 10px;
}

.set-hotkey-display {
  display: inline-block;
  min-width: 140px;
  padding: 6px 12px;
  background: var(--bg);
  border: 1px solid var(--border);
  border-radius: var(--radius-sm);
  font-size: 13px;
  font-family: 'Consolas', 'Monaco', monospace;
  color: var(--text);
}

.up-checkbox {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 13px;
  color: var(--text);
  cursor: pointer;
  user-select: none;
}
</style>
