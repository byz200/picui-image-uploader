<script setup lang="ts">
// App.vue - 根组件：首启站点选择 / 主界面
import { onMounted, computed } from 'vue'
import { useStore } from './store'
import SiteSelect from './components/SiteSelect.vue'
import Sidebar from './components/Sidebar.vue'
import Upload from './components/Upload.vue'
import History from './components/History.vue'
import Albums from './components/Albums.vue'
import Images from './components/Images.vue'
import Settings from './components/Settings.vue'
import ScreenshotOverlay from './components/ScreenshotOverlay.vue'
import Toasts from './components/Toasts.vue'

const store = useStore()

// 首次启动且尚未选择站点：展示站点选择；否则进入主界面
const showSiteSelect = computed(() => store.firstRun || !store.currentSite)

onMounted(async () => {
  await store.init()
  store.registerEvents()
})
</script>

<template>
  <div class="app-root">
    <!-- 首次启动站点选择 -->
    <SiteSelect v-if="showSiteSelect" />

    <!-- 主界面 -->
    <div v-else class="main-layout">
      <Sidebar />
      <main class="content">
        <Upload v-if="store.view === 'upload'" />
        <History v-else-if="store.view === 'history'" />
        <Albums v-else-if="store.view === 'albums'" />
        <Images v-else-if="store.view === 'images'" />
        <Settings v-else-if="store.view === 'settings'" />
      </main>
    </div>

    <!-- 截图选区遮罩（覆盖全屏） -->
    <ScreenshotOverlay v-if="store.screenshotActive" />

    <!-- 全局 Toast -->
    <Toasts />
  </div>
</template>

<style scoped>
.app-root {
  display: flex;
  flex-direction: column;
  height: 100%;
  width: 100%;
}

.main-layout {
  display: flex;
  flex: 1;
  min-height: 0;
}

.content {
  flex: 1;
  min-width: 0;
  min-height: 0;
  overflow: auto;
  background: var(--bg);
}
</style>
