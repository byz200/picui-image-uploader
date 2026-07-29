<script setup lang="ts">
// Upload.vue - 上传核心页：拖拽 / 剪贴板 / 截图 / 文件选择 + 队列
import { ref, computed, onMounted, onBeforeUnmount } from 'vue'
import { useStore } from '../store'
import * as App from '../../wailsjs/go/main/App'
import type * as T from '../../wailsjs/go/main/App'

const store = useStore()

// ---- 相册选择 ----
const albums = ref<T.main.Album[]>([])
const selectedAlbumId = ref<string>('')

async function loadAlbums() {
  try {
    const res = await App.GetAlbums(1, '', 'newest')
    albums.value = res.data || []
  } catch (e) {
    // 静默失败，可能在未配置 Token 时
    albums.value = []
  }
}

// ---- 快速压缩开关 ----
const compressEnabled = ref(false)
const compressFormat = ref<string>('jpg')
const compressQuality = ref<number>(85)

// 同步到当前 store 设置
function syncCompressFromSettings() {
  if (store.settings) {
    compressEnabled.value = store.settings.compress
    compressFormat.value = store.settings.compressFormat || 'jpg'
    compressQuality.value = store.settings.compressQuality || 85
  }
}

// ---- 上传选项构造 ----
function buildOpts(): T.main.UploadOptions {
  return store.buildOpts({
    albumId: selectedAlbumId.value,
    compress: compressEnabled.value,
    compressFormat: compressFormat.value,
    compressQuality: compressQuality.value
  })
}

// ---- 文件选择 ----
const fileInput = ref<HTMLInputElement | null>(null)

function pickFiles() {
  fileInput.value?.click()
}

async function onFilePicked(e: Event) {
  const input = e.target as HTMLInputElement
  if (!input.files || input.files.length === 0) return
  // webview 的 File API 无法直接拿到绝对路径；通过读取为 base64 上传
  const files = Array.from(input.files)
  for (const f of files) {
    try {
      const dataURL = await fileToDataURL(f)
      const filename = f.name
      await store.uploadBase64(dataURL, filename, {
        albumId: selectedAlbumId.value,
        compress: compressEnabled.value,
        compressFormat: compressFormat.value,
        compressQuality: compressQuality.value
      })
    } catch (err) {
      store.toast(`文件 ${f.name} 读取失败：` + String(err), 'error')
    }
  }
  input.value = ''
}

function fileToDataURL(f: File): Promise<string> {
  return new Promise((resolve, reject) => {
    const reader = new FileReader()
    reader.onload = () => resolve(reader.result as string)
    reader.onerror = () => reject(reader.error)
    reader.readAsDataURL(f)
  })
}

// ---- 拖拽（前端 webview 内拖入）----
const dragOver = ref(false)

function onDragOver(e: DragEvent) {
  e.preventDefault()
  dragOver.value = true
}

function onDragLeave(e: DragEvent) {
  e.preventDefault()
  dragOver.value = false
}

async function onDrop(e: DragEvent) {
  e.preventDefault()
  dragOver.value = false
  if (!e.dataTransfer?.files || e.dataTransfer.files.length === 0) return
  const files = Array.from(e.dataTransfer.files)
  for (const f of files) {
    try {
      const dataURL = await fileToDataURL(f)
      await store.uploadBase64(dataURL, f.name, {
        albumId: selectedAlbumId.value,
        compress: compressEnabled.value,
        compressFormat: compressFormat.value,
        compressQuality: compressQuality.value
      })
    } catch (err) {
      store.toast(`文件 ${f.name} 读取失败：` + String(err), 'error')
    }
  }
}

// ---- 剪贴板上传 ----
async function uploadClipboard() {
  await store.uploadClipboard({
    albumId: selectedAlbumId.value,
    compress: compressEnabled.value,
    compressFormat: compressFormat.value,
    compressQuality: compressQuality.value
  })
}

// ---- 截图上传 ----
function startScreenshot() {
  store.startScreenshot()
}

// ---- 队列展示（按创建时间倒序，最新在上）----
const sortedTasks = computed(() => {
  return [...store.tasks].sort((a, b) => b.createdAt - a.createdAt)
})

const stats = computed(() => {
  const all = store.tasks
  return {
    total: all.length,
    success: all.filter((t) => t.status === 'success').length,
    failed: all.filter((t) => t.status === 'failed').length,
    pending: all.filter(
      (t) => t.status === 'pending' || t.status === 'uploading' || t.status === 'retrying'
    ).length
  }
})

// ---- 任务操作 ----
function retry(id: string) {
  store.retryUpload(id)
}

function remove(id: string) {
  store.removeUploadTask(id)
}

function copyLink(task: T.main.UploadTask) {
  if (task.markdown) {
    store.copyText(task.markdown)
  } else if (task.url) {
    store.copyText(task.url)
  }
}

function clearAll() {
  store.clearUploadTasks()
}

// ---- 状态文案/样式 ----
function statusText(s: string): string {
  switch (s) {
    case 'pending':
      return '等待中'
    case 'uploading':
      return '上传中'
    case 'success':
      return '已完成'
    case 'failed':
      return '失败'
    case 'retrying':
      return '重试中'
    default:
      return s
  }
}

function statusClass(s: string): string {
  switch (s) {
    case 'success':
      return 'badge-success'
    case 'failed':
      return 'badge-danger'
    case 'uploading':
    case 'retrying':
      return 'badge-warning'
    default:
      return 'badge-muted'
  }
}

function progressBarClass(s: string): string {
  if (s === 'success') return 'success'
  if (s === 'failed') return 'danger'
  return ''
}

function formatSize(n: number): string {
  if (!n) return '-'
  const u = 1024
  if (n < u) return n + ' B'
  const units = ['KB', 'MB', 'GB']
  let v = n / u
  let i = 0
  while (v >= u && i < units.length - 1) {
    v /= u
    i++
  }
  return v.toFixed(2) + ' ' + units[i]
}

function formatTime(ts: number): string {
  if (!ts) return ''
  const d = new Date(ts)
  return (
    String(d.getHours()).padStart(2, '0') +
    ':' +
    String(d.getMinutes()).padStart(2, '0') +
    ':' +
    String(d.getSeconds()).padStart(2, '0')
  )
}

// ---- 全局粘贴监听（Ctrl+V 触发剪贴板上传）----
function onPaste(e: ClipboardEvent) {
  // 仅当剪贴板含图片时由后端处理；这里检测有图片项时触发上传
  const items = e.clipboardData?.items
  if (!items) return
  let hasImage = false
  for (const it of items) {
    if (it.type.startsWith('image/')) {
      hasImage = true
      break
    }
  }
  if (hasImage) {
    // 让后端从系统剪贴板读取（CF_DIB），保证质量
    e.preventDefault()
    uploadClipboard()
  }
}

onMounted(() => {
  loadAlbums()
  syncCompressFromSettings()
  window.addEventListener('paste', onPaste)
})

onBeforeUnmount(() => {
  window.removeEventListener('paste', onPaste)
})
</script>

<template>
  <div class="upload-page">
    <!-- 顶部操作区 -->
    <div class="up-actions card">
      <div
        class="dropzone"
        :class="{ active: dragOver }"
        @dragover="onDragOver"
        @dragleave="onDragLeave"
        @drop="onDrop"
      >
        <div class="dz-icon">📥</div>
        <div class="dz-text">
          <div class="dz-title">将图片拖拽到此处上传</div>
          <div class="dz-sub">支持 JPG / PNG / GIF / WebP / BMP / TIFF</div>
        </div>
      </div>

      <div class="up-buttons">
        <button class="btn btn-primary" @click="pickFiles">
          <span>📁</span><span>选择文件</span>
        </button>
        <button class="btn btn-outline" @click="uploadClipboard" title="上传剪贴板中的图片">
          <span>📋</span><span>剪贴板上传</span>
        </button>
        <button class="btn btn-outline" @click="startScreenshot" title="屏幕选区截图上传">
          <span>✂️</span><span>截图上传</span>
        </button>
      </div>

      <input
        ref="fileInput"
        type="file"
        multiple
        accept="image/*"
        style="display: none"
        @change="onFilePicked"
      />

      <div class="up-options">
        <div class="up-field">
          <label class="up-label">目标相册</label>
          <select class="select up-album" v-model="selectedAlbumId">
            <option value="">默认（不上传到相册）</option>
            <option v-for="a in albums" :key="a.id" :value="String(a.id)">
              {{ a.name }}{{ a.imageNum ? ` (${a.imageNum})` : '' }}
            </option>
          </select>
        </div>

        <div class="up-field up-compress">
          <label class="up-checkbox">
            <input type="checkbox" v-model="compressEnabled" />
            <span>压缩</span>
          </label>
          <select class="select up-format" v-model="compressFormat" :disabled="!compressEnabled">
            <option value="original">保持原格式</option>
            <option value="jpg">转 JPG</option>
            <option value="png">转 PNG</option>
          </select>
          <div class="up-quality" :class="{ disabled: !compressEnabled || compressFormat === 'png' }">
            <label>质量 {{ compressQuality }}%</label>
            <input
              type="range"
              min="10"
              max="100"
              v-model.number="compressQuality"
              :disabled="!compressEnabled || compressFormat === 'png'"
            />
          </div>
        </div>
      </div>
    </div>

    <!-- 队列列表 -->
    <div class="up-queue card">
      <div class="up-queue-header">
        <div class="up-queue-title">
          <span>上传队列</span>
          <span class="up-stats">
            共 {{ stats.total }} · 成功 {{ stats.success }} · 失败 {{ stats.failed }} ·
            进行中 {{ stats.pending }}
          </span>
        </div>
        <button v-if="store.tasks.length > 0" class="btn btn-ghost btn-sm" @click="clearAll">
          清空
        </button>
      </div>

      <div v-if="sortedTasks.length === 0" class="empty">
        <div class="empty-icon">📭</div>
        <div>暂无上传任务</div>
        <div style="font-size: 12px">拖拽图片、粘贴或截图即可开始上传</div>
      </div>

      <div v-else class="up-list">
        <div v-for="task in sortedTasks" :key="task.id" class="up-task">
          <div class="up-task-main">
            <div class="up-task-name" :title="task.name">{{ task.name }}</div>
            <div class="up-task-meta">
              <span>{{ formatSize(task.size) }}</span>
              <span>·</span>
              <span>{{ formatTime(task.createdAt) }}</span>
              <span v-if="task.retries > 0">· 重试 {{ task.retries }} 次</span>
            </div>
            <div v-if="task.error" class="up-task-error" :title="task.error">
              {{ task.error }}
            </div>
          </div>

          <div class="up-task-progress">
            <div class="up-task-top">
              <span class="badge" :class="statusClass(task.status)">
                {{ statusText(task.status) }}
              </span>
              <span v-if="task.status === 'uploading' || task.status === 'retrying'" class="up-pct">
                {{ task.progress }}%
              </span>
            </div>
            <div class="progress">
              <div
                class="progress-bar"
                :class="progressBarClass(task.status)"
                :style="{ width: (task.status === 'success' ? 100 : task.progress) + '%' }"
              ></div>
            </div>
          </div>

          <div class="up-task-actions">
            <button
              v-if="task.status === 'success'"
              class="btn btn-outline btn-sm"
              @click="copyLink(task)"
              title="复制 Markdown 链接"
            >
              复制
            </button>
            <button
              v-if="task.status === 'failed'"
              class="btn btn-outline btn-sm"
              @click="retry(task.id)"
            >
              重试
            </button>
            <button class="btn btn-ghost btn-sm" @click="remove(task.id)" title="移除">✕</button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.upload-page {
  padding: 20px;
  display: flex;
  flex-direction: column;
  gap: 16px;
  height: 100%;
}

.up-actions {
  padding: 18px;
  display: flex;
  flex-direction: column;
  gap: 14px;
}

.dropzone {
  border: 2px dashed var(--border);
  border-radius: var(--radius);
  padding: 24px;
  display: flex;
  align-items: center;
  gap: 16px;
  transition: all 0.15s ease;
  background: var(--bg);
  cursor: pointer;
}

.dropzone:hover,
.dropzone.active {
  border-color: var(--primary);
  background: var(--primary-soft);
}

.dz-icon {
  font-size: 36px;
}

.dz-title {
  font-size: 15px;
  font-weight: 600;
  color: var(--text);
  margin-bottom: 4px;
}

.dz-sub {
  font-size: 12px;
  color: var(--text-muted);
}

.up-buttons {
  display: flex;
  gap: 10px;
  flex-wrap: wrap;
}

.up-options {
  display: flex;
  gap: 20px;
  flex-wrap: wrap;
  align-items: flex-end;
  padding-top: 4px;
  border-top: 1px solid var(--border);
  padding-top: 14px;
}

.up-field {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.up-label {
  font-size: 12px;
  color: var(--text-secondary);
  font-weight: 500;
}

.up-album {
  min-width: 220px;
}

.up-compress {
  flex-direction: row;
  align-items: center;
  gap: 14px;
  flex-wrap: wrap;
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

.up-checkbox input {
  cursor: pointer;
}

.up-format {
  min-width: 130px;
}

.up-quality {
  display: flex;
  flex-direction: column;
  gap: 4px;
  font-size: 11px;
  color: var(--text-secondary);
}

.up-quality.disabled {
  opacity: 0.45;
}

.up-quality input[type='range'] {
  width: 120px;
  cursor: pointer;
}

.up-queue {
  flex: 1;
  min-height: 0;
  display: flex;
  flex-direction: column;
  padding: 16px;
}

.up-queue-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 12px;
}

.up-queue-title {
  display: flex;
  align-items: center;
  gap: 10px;
  font-size: 15px;
  font-weight: 600;
  color: var(--text);
}

.up-stats {
  font-size: 12px;
  font-weight: 400;
  color: var(--text-muted);
}

.up-list {
  flex: 1;
  overflow-y: auto;
  display: flex;
  flex-direction: column;
  gap: 8px;
  padding-right: 4px;
}

.up-task {
  display: flex;
  align-items: center;
  gap: 14px;
  padding: 12px 14px;
  background: var(--bg);
  border: 1px solid var(--border);
  border-radius: var(--radius-sm);
}

.up-task-main {
  flex: 1;
  min-width: 0;
}

.up-task-name {
  font-size: 13px;
  font-weight: 500;
  color: var(--text);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  margin-bottom: 3px;
}

.up-task-meta {
  display: flex;
  gap: 6px;
  font-size: 11px;
  color: var(--text-muted);
}

.up-task-error {
  font-size: 11px;
  color: var(--danger);
  margin-top: 3px;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.up-task-progress {
  width: 220px;
  flex-shrink: 0;
  display: flex;
  flex-direction: column;
  gap: 5px;
}

.up-task-top {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.up-pct {
  font-size: 11px;
  color: var(--text-secondary);
}

.up-task-actions {
  display: flex;
  gap: 6px;
  flex-shrink: 0;
}
</style>
