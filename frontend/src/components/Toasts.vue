<script setup lang="ts">
// Toasts.vue - 全局 Toast 通知
import { useStore } from '../store'

const store = useStore()
</script>

<template>
  <div class="toast-container">
    <div
      v-for="t in store.toasts"
      :key="t.id"
      class="toast-item"
      :class="'toast-' + t.type"
      @click="store.removeToast(t.id)"
    >
      <span class="toast-icon">{{ t.type === 'success' ? '✓' : t.type === 'error' ? '✕' : 'ℹ' }}</span>
      <span class="toast-text">{{ t.text }}</span>
    </div>
  </div>
</template>

<style scoped>
.toast-container {
  position: fixed;
  top: 20px;
  left: 50%;
  transform: translateX(-50%);
  z-index: 10000;
  display: flex;
  flex-direction: column;
  gap: 8px;
  pointer-events: none;
}

.toast-item {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 10px 16px;
  background: var(--bg-elevated);
  border: 1px solid var(--border);
  border-radius: var(--radius-sm);
  box-shadow: var(--shadow-lg);
  font-size: 13px;
  color: var(--text);
  cursor: pointer;
  pointer-events: auto;
  animation: toast-in 0.2s ease;
  max-width: 460px;
}

.toast-icon {
  font-weight: bold;
  flex-shrink: 0;
}

.toast-success .toast-icon {
  color: var(--success);
}
.toast-error .toast-icon {
  color: var(--danger);
}
.toast-info .toast-icon {
  color: var(--primary);
}

.toast-text {
  flex: 1;
  word-break: break-word;
}

@keyframes toast-in {
  from {
    opacity: 0;
    transform: translateY(-8px);
  }
}
</style>
