<script setup lang="ts">
// History.vue - 按站点隔离的上传历史记录
import { ref, onMounted, computed } from 'vue'
import { useStore } from '../store'
import * as App from '../../wailsjs/go/main/App'
import type * as T from '../../wailsjs/go/main/App'

const store = useStore()

const items = ref<T.main.HistoryItem[]>([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(20)
const loading = ref(false)

const lastPage = computed(() => Math.max(1, Math.ceil(total.value / pageSize.value)))

async function load() {
  loading.value = true
  try {
    const res = await App.GetHistory(page.value, pageSize.value)
    items.value = res.data || []
    total.value = res.total
  } catch (e) {
    store.toast('加载历史失败：' + String(e), 'error')
    items.value = []
  } finally {
    loading.value = false
  }
}

async function deleteItem(id: string) {
  try {
    await App.DeleteHistory(id)
    items.value = items.value.filter((i) => i.id !== id)
    total.value = Math.max(0, total.value - 1)
    store.toast('已删除', 'success', 1500)
  } catch (e) {
    store.toast('删除失败：' + String(e), 'error')
  }
}

async function clearAll() {
  if (!confirm('确定清空当前站点的全部历史记录？此操作不可恢复。')) return
  try {
    await App.ClearHistory()
    items.value = []
    total.value = 0
    page.value = 1
    store.toast('已清空历史记录', 'success')
  } catch (e) {
    store.toast('清空失败：' + String(e), 'error')
  }
}

function copyMarkdown(item: T.main.HistoryItem) {
  if (item.markdown) store.copyText(item.markdown)
  else if (item.url) store.copyText(item.url)
}

function openImage(url: string) {
  store.openURL(url)
}

function goPage(p: number) {
  if (p < 1 || p > lastPage.value || loading.value) return
  page.value = p
  load()
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
  const pad = (n: number) => String(n).padStart(2, '0')
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())} ${pad(d.getHours())}:${pad(d.getMinutes())}:${pad(d.getSeconds())}`
}

onMounted(() => {
  load()
})
</script>

<template>
  <div class="history-page">
    <div class="page-header">
      <div>
        <h2 class="page-title">上传历史</h2>
        <p class="page-sub">共 {{ total }} 条记录（仅当前站点）</p>
      </div>
      <button v-if="items.length > 0" class="btn btn-danger btn-sm" @click="clearAll">
        清空全部
      </button>
    </div>

    <div class="card hist-card">
      <div v-if="loading" class="empty">
        <div class="spinner"></div>
        <div>加载中…</div>
      </div>

      <div v-else-if="items.length === 0" class="empty">
        <div class="empty-icon">🗂️</div>
        <div>暂无历史记录</div>
      </div>

      <div v-else class="hist-list">
        <div v-for="item in items" :key="item.id" class="hist-item">
          <div class="hist-thumb" @click="openImage(item.url)">
            <img v-if="item.url" :src="item.url" :alt="item.name" loading="lazy" />
            <span v-else class="hist-thumb-ph">🖼️</span>
          </div>
          <div class="hist-info">
            <div class="hist-name" :title="item.name">{{ item.name }}</div>
            <div class="hist-meta">
              <span>{{ formatSize(item.size) }}</span>
              <span>·</span>
              <span>{{ formatTime(item.createdAt) }}</span>
            </div>
            <div class="hist-url" :title="item.url">{{ item.url }}</div>
          </div>
          <div class="hist-actions">
            <button class="btn btn-outline btn-sm" @click="copyMarkdown(item)">复制</button>
            <button class="btn btn-outline btn-sm" @click="openImage(item.url)">打开</button>
            <button class="btn btn-danger btn-sm" @click="deleteItem(item.id)">删除</button>
          </div>
        </div>
      </div>

      <!-- 分页 -->
      <div v-if="lastPage > 1" class="hist-pager">
        <button class="btn btn-outline btn-sm" :disabled="page <= 1" @click="goPage(page - 1)">
          上一页
        </button>
        <span class="hist-pager-info">{{ page }} / {{ lastPage }}</span>
        <button
          class="btn btn-outline btn-sm"
          :disabled="page >= lastPage"
          @click="goPage(page + 1)"
        >
          下一页
        </button>
      </div>
    </div>
  </div>
</template>

<style scoped>
.history-page {
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

.hist-card {
  flex: 1;
  min-height: 0;
  display: flex;
  flex-direction: column;
  padding: 16px;
}

.hist-list {
  flex: 1;
  overflow-y: auto;
  display: flex;
  flex-direction: column;
  gap: 8px;
  padding-right: 4px;
}

.hist-item {
  display: flex;
  align-items: center;
  gap: 14px;
  padding: 10px 12px;
  background: var(--bg);
  border: 1px solid var(--border);
  border-radius: var(--radius-sm);
}

.hist-thumb {
  width: 56px;
  height: 56px;
  flex-shrink: 0;
  border-radius: var(--radius-sm);
  overflow: hidden;
  background: var(--bg-hover);
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
}

.hist-thumb img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.hist-thumb-ph {
  font-size: 22px;
  opacity: 0.5;
}

.hist-info {
  flex: 1;
  min-width: 0;
}

.hist-name {
  font-size: 13px;
  font-weight: 500;
  color: var(--text);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  margin-bottom: 3px;
}

.hist-meta {
  display: flex;
  gap: 6px;
  font-size: 11px;
  color: var(--text-muted);
  margin-bottom: 3px;
}

.hist-url {
  font-size: 11px;
  color: var(--primary);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  font-family: 'Consolas', 'Monaco', monospace;
}

.hist-actions {
  display: flex;
  gap: 6px;
  flex-shrink: 0;
}

.hist-pager {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 14px;
  padding-top: 14px;
  margin-top: 8px;
  border-top: 1px solid var(--border);
}

.hist-pager-info {
  font-size: 13px;
  color: var(--text-secondary);
}
</style>
