<script setup lang="ts">
// ScreenshotOverlay.vue - 屏幕选区截图遮罩
// 后端捕获所有显示器后通过 screenshot:ready 事件下发，前端展示并让用户框选区域，
// 框选完成后将选区从源截图裁剪为 PNG data URL，调用 UploadBase64 上传。
import { ref, computed, onMounted, onBeforeUnmount, nextTick } from 'vue'
import { useStore } from '../store'

interface Monitor {
  index: number
  x: number
  y: number
  width: number
  height: number
  dataUrl: string
  img: HTMLImageElement | null
}

const store = useStore()

const payload = computed(() => store.screenshotPayload)
const monitors = ref<Monitor[]>([])
const minX = ref(0)
const minY = ref(0)
const totalWidth = ref(0)
const totalHeight = ref(0)

// 显示区域（webview 窗口内容区）
const viewW = ref(window.innerWidth)
const viewH = ref(window.innerHeight)

// 缩放：让虚拟屏幕等比适应窗口
const scale = computed(() => {
  if (!totalWidth.value || !totalHeight.value) return 1
  const sx = viewW.value / totalWidth.value
  const sy = viewH.value / totalHeight.value
  return Math.min(sx, sy)
})

const offsetX = computed(() => (viewW.value - totalWidth.value * scale.value) / 2)
const offsetY = computed(() => (viewH.value - totalHeight.value * scale.value) / 2)

// 选区（显示坐标）
const selecting = ref(false)
const startX = ref(0)
const startY = ref(0)
const curX = ref(0)
const curY = ref(0)
const hasSelection = ref(false)

const selRect = computed(() => {
  if (!hasSelection.value && !selecting.value) return null
  const x = Math.min(startX.value, curX.value)
  const y = Math.min(startY.value, curY.value)
  const w = Math.abs(curX.value - startX.value)
  const h = Math.abs(curY.value - startY.value)
  return { x, y, w, h }
})

// 加载截图数据
async function loadPayload() {
  const p = payload.value
  if (!p) return
  minX.value = p.minX || 0
  minY.value = p.minY || 0
  totalWidth.value = p.totalWidth || 0
  totalHeight.value = p.totalHeight || 0
  viewW.value = window.innerWidth
  viewH.value = window.innerHeight

  const list: Monitor[] = (p.monitors || []).map((m: any) => ({
    index: m.index,
    x: m.x,
    y: m.y,
    width: m.width,
    height: m.height,
    dataUrl: m.dataUrl,
    img: null
  }))

  // 预加载图片
  await Promise.all(
    list.map(
      (m) =>
        new Promise<void>((resolve) => {
          const img = new Image()
          img.onload = () => {
            m.img = img
            resolve()
          }
          img.onerror = () => resolve()
          img.src = m.dataUrl
        })
    )
  )
  monitors.value = list
}

// 把虚拟屏幕坐标转换为显示坐标
function virtToView(vx: number, vy: number): { x: number; y: number } {
  return {
    x: offsetX.value + (vx - minX.value) * scale.value,
    y: offsetY.value + (vy - minY.value) * scale.value
  }
}

// 鼠标事件
function onMouseDown(e: MouseEvent) {
  if (e.button !== 0) return
  selecting.value = true
  hasSelection.value = false
  startX.value = e.clientX
  startY.value = e.clientY
  curX.value = e.clientX
  curY.value = e.clientY
}

function onMouseMove(e: MouseEvent) {
  if (!selecting.value) return
  curX.value = e.clientX
  curY.value = e.clientY
}

function onMouseUp(e: MouseEvent) {
  if (!selecting.value) return
  selecting.value = false
  const w = Math.abs(curX.value - startX.value)
  const h = Math.abs(curY.value - startY.value)
  if (w < 5 || h < 5) {
    hasSelection.value = false
    return
  }
  hasSelection.value = true
}

function onKey(e: KeyboardEvent) {
  if (e.key === 'Escape') {
    cancelShot()
  } else if (e.key === 'Enter' && hasSelection.value) {
    confirmSelection()
  }
}

// 取选区在虚拟屏幕中的像素坐标
function selectionVirtualRect(): { x: number; y: number; w: number; h: number } | null {
  const r = selRect.value
  if (!r || r.w < 5 || r.h < 5) return null
  // 显示坐标 -> 虚拟坐标
  const vx1 = (r.x - offsetX.value) / scale.value + minX.value
  const vy1 = (r.y - offsetY.value) / scale.value + minY.value
  const vx2 = (r.x + r.w - offsetX.value) / scale.value + minX.value
  const vy2 = (r.y + r.h - offsetY.value) / scale.value + minY.value
  return {
    x: Math.round(vx1),
    y: Math.round(vy1),
    w: Math.round(vx2 - vx1),
    h: Math.round(vy2 - vy1)
  }
}

// 裁剪：遍历每个显示器，取与选区交集部分绘制到输出 canvas
async function cropSelection(): Promise<{ dataUrl: string; width: number; height: number } | null> {
  const rect = selectionVirtualRect()
  if (!rect) return null
  const out = document.createElement('canvas')
  out.width = rect.w
  out.height = rect.h
  const ctx = out.getContext('2d')
  if (!ctx) return null

  for (const m of monitors.value) {
    if (!m.img) continue
    // 显示器在虚拟屏幕中的矩形
    const mx = m.x
    const my = m.y
    const mw = m.width
    const mh = m.height
    // 与选区的交集
    const ix1 = Math.max(rect.x, mx)
    const iy1 = Math.max(rect.y, my)
    const ix2 = Math.min(rect.x + rect.w, mx + mw)
    const iy2 = Math.min(rect.y + rect.h, my + mh)
    if (ix2 <= ix1 || iy2 <= iy1) continue
    // 源图片中的坐标
    const sx = ix1 - mx
    const sy = iy1 - my
    const sw = ix2 - ix1
    const sh = iy2 - iy1
    // 目标 canvas 中的坐标
    const dx = ix1 - rect.x
    const dy = iy1 - rect.y
    // 源图片像素与虚拟坐标 1:1（截图原始分辨率）
    // 但 m.img 的实际像素可能等于 mw/mh（原始捕获），直接按像素绘制
    const imgW = m.img.naturalWidth || mw
    const imgH = m.img.naturalHeight || mh
    const scaleX = imgW / mw
    const scaleY = imgH / mh
    try {
      ctx.drawImage(
        m.img,
        sx * scaleX,
        sy * scaleY,
        sw * scaleX,
        sh * scaleY,
        dx,
        dy,
        sw,
        sh
      )
    } catch {
      // ignore draw errors
    }
  }

  const dataUrl = out.toDataURL('image/png')
  return { dataUrl, width: rect.w, height: rect.h }
}

async function confirmSelection() {
  const result = await cropSelection()
  if (!result) {
    store.toast('选区无效', 'error')
    return
  }
  store.hideScreenshotOverlay()
  const filename = `screenshot_${Date.now()}.png`
  try {
    const id = await store.uploadBase64(result.dataUrl, filename)
    if (id) {
      store.view = 'upload'
      store.toast('截图已加入上传队列', 'success', 1500)
    }
  } catch (e) {
    store.toast('截图上传失败：' + String(e), 'error')
  }
}

function cancelShot() {
  store.hideScreenshotOverlay()
}

// 监听窗口尺寸变化
function onResize() {
  viewW.value = window.innerWidth
  viewH.value = window.innerHeight
}

onMounted(async () => {
  await loadPayload()
  window.addEventListener('resize', onResize)
  window.addEventListener('keydown', onKey)
  await nextTick()
})

onBeforeUnmount(() => {
  window.removeEventListener('resize', onResize)
  window.removeEventListener('keydown', onKey)
})
</script>

<template>
  <div
    class="shot-overlay"
    @mousedown="onMouseDown"
    @mousemove="onMouseMove"
    @mouseup="onMouseUp"
  >
    <!-- 渲染所有显示器截图 -->
    <div class="shot-stage">
      <div
        v-for="m in monitors"
        :key="m.index"
        class="shot-monitor"
        :style="{
          left: offsetX + (m.x - minX) * scale + 'px',
          top: offsetY + (m.y - minY) * scale + 'px',
          width: m.width * scale + 'px',
          height: m.height * scale + 'px'
        }"
      >
        <img :src="m.dataUrl" draggable="false" />
      </div>
      <!-- 遮罩暗层 -->
      <div class="shot-dim"></div>
      <!-- 选区高亮 -->
      <div
        v-if="selRect"
        class="shot-sel"
        :style="{
          left: selRect.x + 'px',
          top: selRect.y + 'px',
          width: selRect.w + 'px',
          height: selRect.h + 'px'
        }"
      >
        <div class="shot-sel-info" v-if="selectionVirtualRect()">
          {{ selectionVirtualRect()?.w }} × {{ selectionVirtualRect()?.h }}
        </div>
      </div>
    </div>

    <!-- 顶部提示 -->
    <div class="shot-tip">
      <span>拖动鼠标框选截图区域 · Enter 确认 · Esc 取消</span>
    </div>

    <!-- 操作按钮 -->
    <div class="shot-actions" v-if="hasSelection">
      <button class="btn btn-primary" @click.stop="confirmSelection">上传选区</button>
      <button class="btn btn-outline" @click.stop="cancelShot">取消</button>
    </div>
  </div>
</template>

<style scoped>
.shot-overlay {
  position: fixed;
  inset: 0;
  z-index: 9500;
  background: #000;
  cursor: crosshair;
  overflow: hidden;
}

.shot-stage {
  position: absolute;
  inset: 0;
}

.shot-monitor {
  position: absolute;
}

.shot-monitor img {
  width: 100%;
  height: 100%;
  display: block;
  user-select: none;
  -webkit-user-drag: none;
}

.shot-dim {
  position: absolute;
  inset: 0;
  background: rgba(0, 0, 0, 0.35);
  pointer-events: none;
}

.shot-sel {
  position: absolute;
  border: 2px solid #4f8cff;
  background: rgba(79, 140, 255, 0.12);
  box-shadow: 0 0 0 9999px rgba(0, 0, 0, 0);
  pointer-events: none;
}

.shot-sel-info {
  position: absolute;
  top: -28px;
  left: 0;
  background: #4f8cff;
  color: #fff;
  font-size: 12px;
  padding: 3px 8px;
  border-radius: 4px;
  white-space: nowrap;
  font-family: 'Consolas', 'Monaco', monospace;
}

.shot-tip {
  position: absolute;
  top: 20px;
  left: 50%;
  transform: translateX(-50%);
  background: rgba(0, 0, 0, 0.7);
  color: #fff;
  font-size: 13px;
  padding: 8px 18px;
  border-radius: 6px;
  pointer-events: none;
}

.shot-actions {
  position: absolute;
  bottom: 30px;
  left: 50%;
  transform: translateX(-50%);
  display: flex;
  gap: 10px;
}

.shot-actions .btn {
  min-width: 100px;
}
</style>
