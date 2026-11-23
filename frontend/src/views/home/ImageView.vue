<template>
  <div class="h-full w-full relative" @wheel.prevent="handleWheel" @dblclick="emit('operate')">
    <img v-show="curSrc" :src="curSrc" class="object-contain w-full h-full min-w-0 min-h-0" />
    <div
      v-show="titleTips"
      class="absolute p-2 top-2 right-2 text-2xl bg-gray-900 text-white"
      :class="{ 'fade-out': showTitleTips }"
    >
      {{ titleTips }}
    </div>
  </div>
</template>
<script setup lang="ts">
import { onKeyStroke, useThrottleFn } from '@vueuse/core'
import type { ImageInfo } from './useImages'

const THROTTLE_TIME = 50

const emit = defineEmits<{
  (e: 'next'): void
  (e: 'prev'): void
  (e: 'operate'): void
}>()
const props = defineProps<{
  imgInfo?: ImageInfo
}>()
const curSrc = ref('')
const titleTips = ref('')
const showTitleTips = ref(false)

const handleWheel = useThrottleFn((e: WheelEvent) => {
  if (e.deltaY > 0) {
    emit('next')
  } else {
    emit('prev')
  }
}, THROTTLE_TIME)

const stopKeyLeft = onKeyStroke(
  'ArrowLeft',
  useThrottleFn(() => emit('prev'), THROTTLE_TIME),
)
const stopKeyRight = onKeyStroke(
  'ArrowRight',
  useThrottleFn(() => emit('next'), THROTTLE_TIME),
)

const stopEsc = onKeyStroke('Escape', () => emit('operate'))

watch(
  () => props.imgInfo,
  async (newVal, _) => {
    if (newVal) {
      curSrc.value = newVal.base64 || ''
      // 触发消失动画
      titleTips.value = `${newVal.name} - ${newVal.index + 1}/${newVal.total}`
      showTitleTips.value = false
      requestAnimationFrame(() => {
        showTitleTips.value = true
      })
    } else {
      curSrc.value = ''
    }
  },
  { immediate: true },
)

onUnmounted(() => {
  stopKeyLeft()
  stopKeyRight()
  stopEsc()
})
</script>
<style scoped>
/* 定义消失动画 */
@keyframes fadeOut {
  0% {
    opacity: 0.8; /* 开始时透明度 100% */
  }
  75% {
    opacity: 0.8; /* 前 3 秒保持透明度 100% */
  }
  100% {
    opacity: 0; /* 最后 1 秒透明度降到 0 */
  }
}

/* 应用动画 */
.fade-out {
  animation: fadeOut 3s ease-in-out 1 forwards;
  /* 4秒总时长 = 3秒显示 + 1秒消失 */
  /* ease-in-out 使动画过渡更自然 */
  /* 1 表示动画只执行一次 */
  /* forwards 使动画结束后保持最终状态（透明度 0） */
}
</style>
