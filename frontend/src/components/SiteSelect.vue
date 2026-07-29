<script setup lang="ts">
// SiteSelect.vue - 首次启动站点选择界面
import { ref } from 'vue'
import { useStore } from '../store'

const store = useStore()
const selected = ref<string>('')
const loading = ref(false)

async function choose(id: string) {
  selected.value = id
}

async function confirm() {
  if (!selected.value || loading.value) return
  loading.value = true
  try {
    await store.selectSite(selected.value)
    store.toast('欢迎使用 Picui 图床上传工具', 'success')
  } catch (e) {
    store.toast('站点选择失败：' + String(e), 'error')
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div class="site-select">
    <div class="ss-card">
      <div class="ss-logo">
        <div class="logo-icon">🖼️</div>
        <h1 class="ss-title">Picui 图床上传工具</h1>
        <p class="ss-subtitle">首次使用，请选择要连接的站点环境</p>
      </div>

      <div class="ss-options">
        <div
          v-for="site in store.sites"
          :key="site.id"
          class="ss-option"
          :class="{ active: selected === site.id }"
          @click="choose(site.id)"
        >
          <div class="ss-option-icon">{{ site.id === 'picui' ? '🌐' : '🚀' }}</div>
          <div class="ss-option-info">
            <div class="ss-option-name">{{ site.name }}</div>
            <div class="ss-option-url">{{ site.baseUrl }}</div>
          </div>
          <div v-if="selected === site.id" class="ss-check">✓</div>
        </div>
      </div>

      <div class="ss-tip">
        两套站点账号、相册与图片数据完全隔离，互不干扰。
      </div>

      <button class="btn btn-primary ss-confirm" :disabled="!selected || loading" @click="confirm">
        <span v-if="loading" class="spinner"></span>
        <span>{{ loading ? '正在进入…' : '进入主界面' }}</span>
      </button>
    </div>
  </div>
</template>

<style scoped>
.site-select {
  height: 100%;
  width: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
  background: linear-gradient(135deg, var(--primary-soft), var(--bg));
  padding: 24px;
}

.ss-card {
  width: 100%;
  max-width: 520px;
  background: var(--bg-elevated);
  border: 1px solid var(--border);
  border-radius: var(--radius);
  box-shadow: var(--shadow-lg);
  padding: 40px 36px;
}

.ss-logo {
  text-align: center;
  margin-bottom: 28px;
}

.logo-icon {
  font-size: 48px;
  margin-bottom: 12px;
}

.ss-title {
  font-size: 22px;
  font-weight: 600;
  color: var(--text);
  margin-bottom: 6px;
}

.ss-subtitle {
  font-size: 13px;
  color: var(--text-secondary);
}

.ss-options {
  display: flex;
  flex-direction: column;
  gap: 12px;
  margin-bottom: 20px;
}

.ss-option {
  display: flex;
  align-items: center;
  gap: 14px;
  padding: 16px 18px;
  border: 2px solid var(--border);
  border-radius: var(--radius);
  cursor: pointer;
  transition: all 0.15s ease;
  background: var(--bg-elevated);
}

.ss-option:hover {
  border-color: var(--primary);
  background: var(--bg-hover);
}

.ss-option.active {
  border-color: var(--primary);
  background: var(--primary-soft);
}

.ss-option-icon {
  font-size: 28px;
  flex-shrink: 0;
}

.ss-option-info {
  flex: 1;
  min-width: 0;
}

.ss-option-name {
  font-size: 15px;
  font-weight: 600;
  color: var(--text);
  margin-bottom: 4px;
}

.ss-option-url {
  font-size: 12px;
  color: var(--text-muted);
  font-family: 'Consolas', 'Monaco', monospace;
}

.ss-check {
  color: var(--primary);
  font-size: 20px;
  font-weight: bold;
  flex-shrink: 0;
}

.ss-tip {
  font-size: 12px;
  color: var(--text-muted);
  text-align: center;
  margin-bottom: 24px;
  line-height: 1.6;
}

.ss-confirm {
  width: 100%;
  padding: 12px;
  font-size: 15px;
}
</style>
