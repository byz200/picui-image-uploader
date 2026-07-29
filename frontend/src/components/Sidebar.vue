<script setup lang="ts">
// Sidebar.vue - 侧边栏导航
import { computed } from 'vue'
import { useStore, type ViewName } from '../store'

const store = useStore()

interface NavItem {
  key: ViewName
  label: string
  icon: string
}

const navItems: NavItem[] = [
  { key: 'upload', label: '上传', icon: '📤' },
  { key: 'history', label: '历史', icon: '📋' },
  { key: 'albums', label: '相册', icon: '📁' },
  { key: 'images', label: '图库', icon: '🖼️' },
  { key: 'settings', label: '设置', icon: '⚙️' }
]

const pendingCount = computed(() => {
  return store.tasks.filter(
    (t) => t.status === 'pending' || t.status === 'uploading' || t.status === 'retrying'
  ).length
})

function switchView(v: ViewName) {
  store.view = v
}

function toggleTheme() {
  const current = store.settings?.theme || 'light'
  const next = current === 'dark' ? 'light' : 'dark'
  store.setTheme(next)
}

const isDark = computed(() => {
  const t = store.settings?.theme
  if (t === 'dark') return true
  if (t === 'light') return false
  return false
})
</script>

<template>
  <aside class="sidebar">
    <div class="sb-brand">
      <span class="sb-logo">🖼️</span>
      <span class="sb-name">Picui</span>
    </div>

    <div class="sb-site" v-if="store.currentSiteInfo">
      <div class="sb-site-name">{{ store.currentSiteInfo.name }}</div>
      <div class="sb-site-url">{{ store.currentSiteInfo.baseUrl }}</div>
      <div class="sb-site-status">
        <span class="dot" :class="store.hasToken ? 'ok' : 'warn'"></span>
        <span>{{ store.hasToken ? '已配置 Token' : '未配置 Token' }}</span>
      </div>
    </div>

    <nav class="sb-nav">
      <button
        v-for="item in navItems"
        :key="item.key"
        class="sb-item"
        :class="{ active: store.view === item.key }"
        @click="switchView(item.key)"
      >
        <span class="sb-icon">{{ item.icon }}</span>
        <span class="sb-label">{{ item.label }}</span>
        <span v-if="item.key === 'upload' && pendingCount > 0" class="sb-badge">
          {{ pendingCount }}
        </span>
      </button>
    </nav>

    <div class="sb-footer">
      <button class="sb-item sb-theme" @click="toggleTheme" title="切换主题">
        <span class="sb-icon">{{ isDark ? '☀️' : '🌙' }}</span>
        <span class="sb-label">{{ isDark ? '明亮' : '暗黑' }}</span>
      </button>
      <button class="sb-item sb-minimize" @click="store.minimizeToTray" title="最小化到托盘">
        <span class="sb-icon">📥</span>
        <span class="sb-label">最小化</span>
      </button>
    </div>
  </aside>
</template>

<style scoped>
.sidebar {
  width: 200px;
  flex-shrink: 0;
  background: var(--sidebar-bg);
  border-right: 1px solid var(--border);
  display: flex;
  flex-direction: column;
  height: 100%;
  user-select: none;
}

.sb-brand {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 18px 18px 14px;
  font-size: 18px;
  font-weight: 700;
  color: var(--text);
}

.sb-logo {
  font-size: 22px;
}

.sb-site {
  margin: 0 12px 12px;
  padding: 12px;
  background: var(--bg-hover);
  border-radius: var(--radius-sm);
}

.sb-site-name {
  font-size: 13px;
  font-weight: 600;
  color: var(--text);
  margin-bottom: 2px;
}

.sb-site-url {
  font-size: 11px;
  color: var(--text-muted);
  font-family: 'Consolas', 'Monaco', monospace;
  margin-bottom: 6px;
  word-break: break-all;
}

.sb-site-status {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 11px;
  color: var(--text-secondary);
}

.dot {
  width: 7px;
  height: 7px;
  border-radius: 50%;
  flex-shrink: 0;
}

.dot.ok {
  background: var(--success);
}

.dot.warn {
  background: var(--warning);
}

.sb-nav {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 2px;
  padding: 0 10px;
  overflow-y: auto;
}

.sb-item {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 10px 12px;
  border-radius: var(--radius-sm);
  color: var(--text-secondary);
  font-size: 14px;
  transition: all 0.12s ease;
  width: 100%;
  text-align: left;
  position: relative;
}

.sb-item:hover {
  background: var(--bg-hover);
  color: var(--text);
}

.sb-item.active {
  background: var(--primary-soft);
  color: var(--primary);
  font-weight: 600;
}

.sb-icon {
  font-size: 16px;
  width: 20px;
  text-align: center;
  flex-shrink: 0;
}

.sb-label {
  flex: 1;
}

.sb-badge {
  background: var(--primary);
  color: #fff;
  font-size: 11px;
  font-weight: 600;
  min-width: 18px;
  height: 18px;
  padding: 0 5px;
  border-radius: 9px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
}

.sb-footer {
  padding: 8px 10px 12px;
  display: flex;
  flex-direction: column;
  gap: 2px;
  border-top: 1px solid var(--border);
}
</style>
