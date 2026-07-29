<script setup lang="ts">
// Images.vue - 图库浏览：分页 / 搜索 / 排序 / 权限过滤 / 复制链接 / 删除
import { ref, computed, onMounted } from 'vue'
import { useStore } from '../store'
import * as App from '../../wailsjs/go/main/App'
import type * as T from '../../wailsjs/go/main/App'

const store = useStore()

const items = ref<T.main.ImageItem[]>([])
const total = ref(0)
const page = ref(1)
const lastPage = ref(1)
const loading = ref(false)

const q = ref('')
const order = ref<'newest' | 'earliest' | 'utmost' | 'least'>('newest')
const permission = ref<'' | 'public' | 'private'>('')

let searchTimer: any = null

async function load() {
  loading.value = true
  try {
    const res = await App.GetImages(
      page.value,
      q.value,
      order.value,
      permission.value,
      ''
    )
    items.value = res.data || []
    total.value = res.total
    lastPage.value = Math.max(1, res.lastPage)
  } catch (e) {
    store.toast('加载图库失败：' + String(e), 'error')
    items.value = []
  } finally {
    loading.value = false
  }
}

function onSearch() {
  if (searchTimer) clearTimeout(searchTimer)
  searchTimer = setTimeout(() => {
    page.value = 1
    load()
  }, 350)
}

function onFilterChange() {
  page.value = 1
  load()
}

function goPage(p: number) {
  if (p < 1 || p > lastPage.value || loading.value) return
  page.value = p
  load()
}

// ---- 详情弹窗 ----
const detail = ref<T.main.ImageItem | null>(null)

function openDetail(img: T.main.ImageItem) {
  detail.value = img
}

function closeDetail() {
  detail.value = null
}

function copyLink(field: 'markdown' | 'url' | 'bbcode' | 'markdownWithLink' | 'html') {
  if (!detail.value) return
  const links = detail.value.links
  let text = ''
  switch (field) {
    case 'markdown':
      text = links.markdown
      break
    case 'url':
      text = links.url
      break
    case 'bbcode':
      text = links.bbcode
      break
    case 'markdownWithLink':
      text = links.markdownWithLink
      break
    case 'html':
      text = links.html
      break
  }
  if (text) store.copyText(text)
}

async function deleteImage(img: T.main.ImageItem) {
  if (!confirm(`确定删除图片「${img.name}」？此操作不可恢复。`)) return
  try {
    await App.DeleteImage(img.key)
    store.toast('图片已删除', 'success')
    items.value = items.value.filter((i) => i.key !== img.key)
    total.value = Math.max(0, total.value - 1)
    if (detail.value?.key === img.key) closeDetail()
  } catch (e) {
    store.toast('删除失败：' + String(e), 'error')
  }
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

const orderOptions = [
  { value: 'newest', label: '最新' },
  { value: 'earliest', label: '最早' },
  { value: 'utmost', label: '最大' },
  { value: 'least', label: '最小' }
]

onMounted(() => {
  load()
})
</script>

<template>
  <div class="images-page">
    <div class="page-header">
      <div>
        <h2 class="page-title">图库</h2>
        <p class="page-sub">共 {{ total }} 张图片（仅当前站点）</p>
      </div>
    </div>

    <div class="card img-card">
      <div class="img-toolbar">
        <input
          class="input img-search"
          v-model="q"
          placeholder="搜索图片名称…"
          @input="onSearch"
        />
        <select class="select img-filter" v-model="order" @change="onFilterChange">
          <option v-for="o in orderOptions" :key="o.value" :value="o.value">{{ o.label }}</option>
        </select>
        <select class="select img-filter" v-model="permission" @change="onFilterChange">
          <option value="">全部权限</option>
          <option value="public">公开</option>
          <option value="private">私有</option>
        </select>
      </div>

      <div v-if="loading" class="empty">
        <div class="spinner"></div>
        <div>加载中…</div>
      </div>

      <div v-else-if="items.length === 0" class="empty">
        <div class="empty-icon">🖼️</div>
        <div>暂无图片</div>
      </div>

      <div v-else class="img-grid">
        <div
          v-for="img in items"
          :key="img.key"
          class="img-cell"
          @click="openDetail(img)"
        >
          <div class="img-thumb">
            <img :src="img.links.thumbnailUrl || img.links.url" :alt="img.name" loading="lazy" />
          </div>
          <div class="img-cap" :title="img.name">{{ img.name }}</div>
          <div class="img-cap-sub">{{ img.humanDate }}</div>
        </div>
      </div>

      <div v-if="lastPage > 1" class="img-pager">
        <button class="btn btn-outline btn-sm" :disabled="page <= 1" @click="goPage(page - 1)">
          上一页
        </button>
        <span class="img-pager-info">{{ page }} / {{ lastPage }}</span>
        <button
          class="btn btn-outline btn-sm"
          :disabled="page >= lastPage"
          @click="goPage(page + 1)"
        >
          下一页
        </button>
      </div>
    </div>

    <!-- 详情弹窗 -->
    <div v-if="detail" class="modal-mask" @click.self="closeDetail">
      <div class="modal card img-detail">
        <div class="img-detail-preview">
          <img :src="detail.links.url" :alt="detail.name" />
        </div>
        <div class="img-detail-info">
          <div class="img-detail-name">{{ detail.name }}</div>
          <div class="img-detail-meta">
            <span>{{ formatSize(detail.size) }}</span>
            <span v-if="detail.width"> · {{ detail.width }}×{{ detail.height }}</span>
            <span> · {{ detail.date }}</span>
          </div>

          <div class="img-detail-links">
            <button class="btn btn-outline btn-sm" @click="copyLink('markdown')">复制 Markdown</button>
            <button class="btn btn-outline btn-sm" @click="copyLink('url')">复制 URL</button>
            <button class="btn btn-outline btn-sm" @click="copyLink('bbcode')">复制 BBCode</button>
            <button class="btn btn-outline btn-sm" @click="copyLink('markdownWithLink')">
              复制带链接 MD
            </button>
            <button class="btn btn-outline btn-sm" @click="store.openURL(detail.links.url)">
              在浏览器打开
            </button>
            <button class="btn btn-danger btn-sm" @click="deleteImage(detail)">删除图片</button>
          </div>

          <div class="img-detail-url" :title="detail.links.url">{{ detail.links.url }}</div>
        </div>
        <button class="modal-close" @click="closeDetail">✕</button>
      </div>
    </div>
  </div>
</template>

<style scoped>
.images-page {
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

.img-card {
  flex: 1;
  min-height: 0;
  display: flex;
  flex-direction: column;
  padding: 16px;
}

.img-toolbar {
  display: flex;
  gap: 10px;
  margin-bottom: 14px;
  flex-wrap: wrap;
}

.img-search {
  flex: 1;
  min-width: 200px;
  max-width: 360px;
}

.img-filter {
  width: 130px;
}

.img-grid {
  flex: 1;
  overflow-y: auto;
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(150px, 1fr));
  gap: 14px;
  padding-right: 4px;
  align-content: start;
}

.img-cell {
  cursor: pointer;
  border: 1px solid var(--border);
  border-radius: var(--radius-sm);
  overflow: hidden;
  background: var(--bg);
  transition: all 0.15s ease;
}

.img-cell:hover {
  border-color: var(--primary);
  transform: translateY(-2px);
  box-shadow: var(--shadow-lg);
}

.img-thumb {
  width: 100%;
  aspect-ratio: 1;
  background: var(--bg-hover);
  display: flex;
  align-items: center;
  justify-content: center;
  overflow: hidden;
}

.img-thumb img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.img-cap {
  padding: 6px 8px 2px;
  font-size: 12px;
  color: var(--text);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.img-cap-sub {
  padding: 0 8px 6px;
  font-size: 11px;
  color: var(--text-muted);
}

.img-pager {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 14px;
  padding-top: 14px;
  margin-top: 8px;
  border-top: 1px solid var(--border);
}

.img-pager-info {
  font-size: 13px;
  color: var(--text-secondary);
}

/* 详情弹窗 */
.modal-mask {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.55);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 9000;
  padding: 24px;
}

.img-detail {
  display: flex;
  gap: 20px;
  max-width: 900px;
  width: 100%;
  max-height: 86vh;
  padding: 20px;
  position: relative;
}

.img-detail-preview {
  flex: 1;
  min-width: 0;
  max-height: 80vh;
  display: flex;
  align-items: center;
  justify-content: center;
  background: var(--bg);
  border-radius: var(--radius-sm);
  overflow: hidden;
}

.img-detail-preview img {
  max-width: 100%;
  max-height: 80vh;
  object-fit: contain;
}

.img-detail-info {
  width: 280px;
  flex-shrink: 0;
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.img-detail-name {
  font-size: 15px;
  font-weight: 600;
  color: var(--text);
  word-break: break-all;
}

.img-detail-meta {
  font-size: 12px;
  color: var(--text-muted);
}

.img-detail-links {
  display: flex;
  flex-direction: column;
  gap: 6px;
  margin-top: 6px;
}

.img-detail-url {
  font-size: 11px;
  color: var(--primary);
  word-break: break-all;
  font-family: 'Consolas', 'Monaco', monospace;
  margin-top: 6px;
  padding: 8px;
  background: var(--bg-hover);
  border-radius: var(--radius-sm);
}

.modal-close {
  position: absolute;
  top: 10px;
  right: 12px;
  width: 28px;
  height: 28px;
  border-radius: 50%;
  color: var(--text-secondary);
  font-size: 14px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.modal-close:hover {
  background: var(--bg-hover);
  color: var(--text);
}
</style>
