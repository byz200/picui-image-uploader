<script setup lang="ts">
// Albums.vue - 相册管理：列表 / 搜索 / 排序 / 新建 / 编辑 / 删除
import { ref, computed, onMounted } from 'vue'
import { useStore } from '../store'
import * as App from '../../wailsjs/go/main/App'
import type * as T from '../../wailsjs/go/main/App'

const store = useStore()

const items = ref<T.main.Album[]>([])
const total = ref(0)
const page = ref(1)
const lastPage = ref(1)
const loading = ref(false)

const q = ref('')
const order = ref<'newest' | 'earliest' | 'most' | 'least'>('newest')

let searchTimer: any = null

async function load() {
  loading.value = true
  try {
    const res = await App.GetAlbums(page.value, q.value, order.value)
    items.value = res.data || []
    total.value = res.total
    lastPage.value = Math.max(1, res.lastPage)
  } catch (e) {
    store.toast('加载相册失败：' + String(e), 'error')
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

function onOrderChange() {
  page.value = 1
  load()
}

function goPage(p: number) {
  if (p < 1 || p > lastPage.value || loading.value) return
  page.value = p
  load()
}

// ---- 新建/编辑弹窗 ----
const showEditor = ref(false)
const editingId = ref<number | null>(null)
const form = ref({ name: '', intro: '', isPublic: true })
const saving = ref(false)

function openCreate() {
  editingId.value = null
  form.value = { name: '', intro: '', isPublic: true }
  showEditor.value = true
}

function openEdit(album: T.main.Album) {
  editingId.value = album.id
  form.value = { name: album.name, intro: album.intro || '', isPublic: true }
  showEditor.value = true
}

async function saveAlbum() {
  if (!form.value.name.trim()) {
    store.toast('请输入相册名称', 'error')
    return
  }
  saving.value = true
  try {
    if (editingId.value === null) {
      await App.CreateAlbum(form.value.name.trim(), form.value.intro.trim(), form.value.isPublic)
      store.toast('相册已创建', 'success')
    } else {
      await App.UpdateAlbum(
        editingId.value,
        form.value.name.trim(),
        form.value.intro.trim(),
        form.value.isPublic
      )
      store.toast('相册已更新', 'success')
    }
    showEditor.value = false
    load()
  } catch (e) {
    store.toast('保存失败：' + String(e), 'error')
  } finally {
    saving.value = false
  }
}

async function deleteAlbum(album: T.main.Album) {
  if (!confirm(`确定删除相册「${album.name}」？相册内的图片不会被删除。`)) return
  try {
    await App.DeleteAlbum(album.id)
    store.toast('相册已删除', 'success')
    load()
  } catch (e) {
    store.toast('删除失败：' + String(e), 'error')
  }
}

onMounted(() => {
  load()
})
</script>

<template>
  <div class="albums-page">
    <div class="page-header">
      <div>
        <h2 class="page-title">相册管理</h2>
        <p class="page-sub">共 {{ total }} 个相册（仅当前站点）</p>
      </div>
      <button class="btn btn-primary" @click="openCreate">
        <span>＋</span><span>新建相册</span>
      </button>
    </div>

    <div class="card alb-card">
      <div class="alb-toolbar">
        <input
          class="input alb-search"
          v-model="q"
          placeholder="搜索相册名称…"
          @input="onSearch"
        />
        <select class="select alb-order" v-model="order" @change="onOrderChange">
          <option value="newest">最新优先</option>
          <option value="earliest">最早优先</option>
          <option value="most">图片最多</option>
          <option value="least">图片最少</option>
        </select>
      </div>

      <div v-if="loading" class="empty">
        <div class="spinner"></div>
        <div>加载中…</div>
      </div>

      <div v-else-if="items.length === 0" class="empty">
        <div class="empty-icon">📁</div>
        <div>暂无相册</div>
        <button class="btn btn-outline btn-sm" @click="openCreate">新建相册</button>
      </div>

      <div v-else class="alb-list">
        <div v-for="album in items" :key="album.id" class="alb-item">
          <div class="alb-icon">📁</div>
          <div class="alb-info">
            <div class="alb-name" :title="album.name">{{ album.name }}</div>
            <div class="alb-meta">
              <span>{{ album.imageNum }} 张图片</span>
              <span v-if="album.intro"> · {{ album.intro }}</span>
            </div>
          </div>
          <div class="alb-actions">
            <button class="btn btn-outline btn-sm" @click="openEdit(album)">编辑</button>
            <button class="btn btn-danger btn-sm" @click="deleteAlbum(album)">删除</button>
          </div>
        </div>
      </div>

      <div v-if="lastPage > 1" class="alb-pager">
        <button class="btn btn-outline btn-sm" :disabled="page <= 1" @click="goPage(page - 1)">
          上一页
        </button>
        <span class="alb-pager-info">{{ page }} / {{ lastPage }}</span>
        <button
          class="btn btn-outline btn-sm"
          :disabled="page >= lastPage"
          @click="goPage(page + 1)"
        >
          下一页
        </button>
      </div>
    </div>

    <!-- 编辑弹窗 -->
    <div v-if="showEditor" class="modal-mask" @click.self="showEditor = false">
      <div class="modal card">
        <div class="modal-title">
          {{ editingId === null ? '新建相册' : '编辑相册' }}
        </div>
        <div class="modal-body">
          <div class="form-field">
            <label>名称</label>
            <input class="input" v-model="form.name" placeholder="请输入相册名称" maxlength="100" />
          </div>
          <div class="form-field">
            <label>简介</label>
            <textarea
              class="textarea"
              v-model="form.intro"
              placeholder="可选，相册描述"
              maxlength="500"
            ></textarea>
          </div>
          <label class="up-checkbox">
            <input type="checkbox" v-model="form.isPublic" />
            <span>公开相册</span>
          </label>
        </div>
        <div class="modal-footer">
          <button class="btn btn-ghost" @click="showEditor = false">取消</button>
          <button class="btn btn-primary" :disabled="saving" @click="saveAlbum">
            <span v-if="saving" class="spinner"></span>
            <span>保存</span>
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.albums-page {
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

.alb-card {
  flex: 1;
  min-height: 0;
  display: flex;
  flex-direction: column;
  padding: 16px;
}

.alb-toolbar {
  display: flex;
  gap: 10px;
  margin-bottom: 14px;
}

.alb-search {
  flex: 1;
  max-width: 360px;
}

.alb-order {
  width: 150px;
}

.alb-list {
  flex: 1;
  overflow-y: auto;
  display: flex;
  flex-direction: column;
  gap: 8px;
  padding-right: 4px;
}

.alb-item {
  display: flex;
  align-items: center;
  gap: 14px;
  padding: 12px 14px;
  background: var(--bg);
  border: 1px solid var(--border);
  border-radius: var(--radius-sm);
}

.alb-icon {
  font-size: 28px;
  flex-shrink: 0;
}

.alb-info {
  flex: 1;
  min-width: 0;
}

.alb-name {
  font-size: 14px;
  font-weight: 600;
  color: var(--text);
  margin-bottom: 3px;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.alb-meta {
  font-size: 12px;
  color: var(--text-muted);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.alb-actions {
  display: flex;
  gap: 6px;
  flex-shrink: 0;
}

.alb-pager {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 14px;
  padding-top: 14px;
  margin-top: 8px;
  border-top: 1px solid var(--border);
}

.alb-pager-info {
  font-size: 13px;
  color: var(--text-secondary);
}

/* 弹窗 */
.modal-mask {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.4);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 9000;
}

.modal {
  width: 420px;
  max-width: 90vw;
  padding: 20px;
}

.modal-title {
  font-size: 16px;
  font-weight: 600;
  color: var(--text);
  margin-bottom: 16px;
}

.modal-body {
  display: flex;
  flex-direction: column;
  gap: 14px;
  margin-bottom: 18px;
}

.form-field {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.form-field label {
  font-size: 12px;
  color: var(--text-secondary);
  font-weight: 500;
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

.modal-footer {
  display: flex;
  justify-content: flex-end;
  gap: 8px;
}
</style>
