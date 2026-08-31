<template>
  <div class="w-full h-full select-none relative">
    <div class="h-full w-full" @wheel.prevent="handleWheel" @dblclick="showDialog">
      <img v-show="curSrc" :src="curSrc" class="object-contain w-full h-full min-w-0 min-h-0" />
    </div>
    <el-dialog v-model="dialogVisible" fullscreen :show-close="false" body-class="h-full">
      <el-tabs v-model="activeName" type="card" class="h-full" :before-leave="handleTabClick">
        <el-tab-pane label="Music" name="music" class="h-full">
          <div class="w-full h-full flex flex-col gap-2">
            <div class="flex flex-row gap-2">
              <el-input v-model="searchText" placeholder="Search" clearable class="flex-2" />
              <el-select
                v-model="selectedSortType"
                placeholder="Sort Type"
                class="flex-1"
                @change="handleChangeSortType"
              >
                <el-option v-for="item in sortTypeList" :key="item" :label="item" :value="item" />
              </el-select>
            </div>
            <el-scrollbar class="flex-1 min-h-0">
              <div class="flex flex-wrap gap-1">
                <el-tag
                  v-for="(item, idx) in filteredImgs"
                  :key="idx"
                  @click="handleChose(item.path)"
                  :type="item.path === curPath ? 'primary' : 'info'"
                  class="cursor-pointer"
                  size="large"
                >
                  {{ item.name }}
                </el-tag>
              </div>
            </el-scrollbar>
          </div>
        </el-tab-pane>
        <el-tab-pane label="Sources" name="sources" class="h-full">
          <div class="flex flex-col gap-2 h-full">
            <el-button-group>
              <el-button
                v-for="item in actions"
                :key="item.name"
                :icon="item.icon"
                :title="item.info"
                :type="item.type || undefined"
                @click="handleChangeSources(item.fn)"
              />
            </el-button-group>
            <el-scrollbar class="flex-1 min-h-0">
              <div
                v-for="(item, index) in sources"
                :key="index"
                class="pl-1 truncate hover:bg-green-600 hover:text-white cursor-pointer"
                @click="handleChangeSources(() => RemoveSource(item.Path))"
              >
                <span>{{ item.Path }}</span>
              </div>
            </el-scrollbar>
          </div>
        </el-tab-pane>
        <el-tab-pane label="Document" name="document" :disabled="imgs.length === 0" class="h-full">
          <el-scrollbar class="flex-1 min-h-0">
            <MusicDocument />
          </el-scrollbar>
        </el-tab-pane>
        <el-tab-pane label="Close" name="close"></el-tab-pane>
      </el-tabs>
    </el-dialog>
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
import MusicDocument from './MusicDocument.vue'
import {
  AddSource,
  ClearSources,
  GetBase64,
  GetConfigs,
  GetImages,
  GetSources,
  RemoveSource,
  SetSortType,
} from '@/wailsjs/go/app/app'
import { DocumentAdd, Folder, FolderAdd, Delete, Refresh } from '@element-plus/icons-vue'
import { app } from '@/wailsjs/go/models'
import { procImgFilePaths, type Image } from './utils'
import { onKeyStroke, useThrottleFn } from '@vueuse/core'
import type { ButtonType, TabPaneName } from 'element-plus'
import { EventsEmit } from '@/wailsjs/runtime/runtime'
import type { IconPropType } from 'element-plus/es/utils/index.mjs'
import { useSortType } from './sortType.ts'

interface Actions {
  icon: IconPropType
  fn: () => Promise<void>
  name?: string
  info: string
  type?: ButtonType
}

const THROTTLE_TIME = 50

const { sortTypeList, selectedSortType, sortTypeInit } = useSortType()
const sources = ref<app.Source[]>([])
const imgs = ref<Image[]>([])
const searchText = ref('')
const curIndex = ref(-1)
const curSrc = ref('')
const dialogVisible = ref(false)
const activeName = ref('sources')
const titleTips = ref('')
const showTitleTips = ref(false)

const actions: Actions[] = [
  {
    icon: Folder,
    fn: () => AddSource(true, true),
    name: 'Add Source (Traversal)',
    info: 'Add a image folder (traversal)',
  },
  {
    icon: FolderAdd,
    fn: () => AddSource(true, false),
    name: 'Add Source',
    info: 'Add a image folder',
  },
  {
    icon: DocumentAdd,
    fn: () => AddSource(false, false),
    name: 'Add Source (Single)',
    info: 'Add a image file',
  },
  {
    icon: Delete,
    fn: () => ClearSources(),
    name: 'Clear Sources',
    info: 'Clear all sources',
  },
  {
    icon: Refresh,
    fn: async () => {},
    name: 'Refresh',
    info: 'Refresh sources',
  },
]

const filteredImgs = computed(() => {
  if (!searchText.value) {
    return imgs.value
  }
  return imgs.value.filter(
    (item) =>
      item.name.toLowerCase().includes(searchText.value.toLowerCase()) ||
      item.pyletter.toLowerCase().includes(searchText.value.toLowerCase()) ||
      item.py.toLowerCase().includes(searchText.value.toLowerCase()),
  )
})

const curPath = computed(() => imgs.value[curIndex.value]?.path || '')

async function updateImages() {
  const lastPath = imgs.value[curIndex.value]?.path || ''
  const list = await GetImages()
  imgs.value = procImgFilePaths(list)
  if (imgs.value.length > 0) {
    if (curIndex.value === -1) {
      curIndex.value = 0
    } else {
      // 触发 watch 变化
      curIndex.value = -1
      await nextTick()
      if (lastPath) {
        const lastIndex = imgs.value.findIndex((item) => item.path === lastPath)
        if (lastIndex !== -1) {
          curIndex.value = lastIndex
        } else {
          curIndex.value = 0
        }
      } else {
        curIndex.value = 0
      }
    }
  } else {
    curIndex.value = -1
  }
}

async function changeSources() {
  sources.value = (await GetSources()).sort(
    (a, b) =>
      Number(b.IsDir) - Number(a.IsDir) || a.Path.localeCompare(b.Path, 'zh-CN', { numeric: true }),
  )
  await updateImages()
}

const handleChangeSources = async (fn: () => Promise<void>) => {
  await fn()
  await changeSources()
}

const handleChose = (path: string) => {
  curIndex.value = imgs.value.findIndex((item) => item.path === path)
  dialogVisible.value = false
}

const handleTabClick = (activeName: TabPaneName, _: TabPaneName) => {
  if (activeName === 'close') {
    dialogVisible.value = false
    return false
  }
  return true
}

function prevImg() {
  if (curIndex.value > 0) {
    curIndex.value--
  }
}

function nextImg() {
  if (curIndex.value < imgs.value.length - 1) {
    curIndex.value++
  }
}

async function showDialog() {
  if (!dialogVisible.value) {
    dialogVisible.value = true
  }
}

const handleChangeSortType = async () => {
  if (selectedSortType.value) {
    await SetSortType(selectedSortType.value)
    await updateImages()
  }
}

watch(curIndex, async (newVal, _) => {
  const newImg = imgs.value[newVal]
  if (newVal !== -1 && newImg) {
    if (!newImg.base64) {
      const base64 = await GetBase64(newImg.path)
      newImg.base64 = base64
      curSrc.value = base64
    } else {
      curSrc.value = newImg.base64
    }

    if (newImg && newImg.name) {
      // 触发消失动画
      titleTips.value = `${newImg.name} - ${newVal + 1}/${imgs.value.length}`
      showTitleTips.value = false
      requestAnimationFrame(() => {
        showTitleTips.value = true
      })
    }
  } else {
    curSrc.value = ''
  }

  EventsEmit('index-change', newVal)
})

const stopKeyLeft = onKeyStroke('ArrowLeft', useThrottleFn(prevImg, THROTTLE_TIME))
const stopKeyRight = onKeyStroke('ArrowRight', useThrottleFn(nextImg, THROTTLE_TIME))
const stopEsc = onKeyStroke('Escape', showDialog)
const handleWheel = useThrottleFn((e: WheelEvent) => {
  if (e.deltaY > 0) {
    nextImg()
  } else {
    prevImg()
  }
}, THROTTLE_TIME)

onMounted(async () => {
  await handleChangeSources(async () => {})
  const cfg = await GetConfigs()
  await sortTypeInit(cfg.SortType)
  if (imgs.value.length > 0) {
    const latestIndex = cfg.LatestIndex
    if (latestIndex !== -1 && curIndex.value != latestIndex) {
      await nextTick()
      curIndex.value = latestIndex
    }
  } else {
    activeName.value = 'sources'
    showDialog()
  }
})

onUnmounted(() => {
  stopKeyLeft()
  stopKeyRight()
  stopEsc()
})
</script>
<style scoped>
:deep(.el-dialog__header) {
  padding-bottom: 0;
}

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
