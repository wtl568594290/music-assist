<template>
  <div class="w-full h-full select-none">
    <image-view :img-info="imgInfo" @operate="showDialog" @next="nextImg" @prev="prevImg" />
    <el-dialog v-model="dialogVisible" fullscreen :show-close="false" body-class="h-full">
      <el-tabs v-model="activeName" type="card" class="h-full" :before-leave="handleTabClick">
        <el-tab-pane label="Music" name="music" class="h-full">
          <music-view
            v-model="selectedSortType"
            :sort-type-list="sortTypeList"
            :imgs="imgs"
            :cur-index="curIndex"
            @change="handleChangeSortType"
            @chose="handleChose"
          />
        </el-tab-pane>
        <el-tab-pane label="Sources" name="sources" class="h-full">
          <source-view :sources="sources" @change-sources="changeSources" />
        </el-tab-pane>
        <el-tab-pane label="Document" name="document" class="h-full">
          <document-view :disabled="imgs.length === 0" />
        </el-tab-pane>
        <el-tab-pane label="About" name="about" class="h-full">
          <about-view @VersionClick="updateViewRef?.check()" />
        </el-tab-pane>
        <el-tab-pane label="Close" name="close"></el-tab-pane>
      </el-tabs>
    </el-dialog>
  </div>
  <update-view ref="updateViewRef" />
</template>
<script setup lang="ts">
import ImageView from './ImageView.vue'
import MusicView from './MusicView.vue'
import SourceView from './SourceView.vue'
import DocumentView from './DocumentView.vue'
import AboutView from './AboutView.vue'
import UpdateView from './UpdateView.vue'

import { GetImgSortType } from '@bindings/musicassist/internal/services/appservice'

import { useDebounceFn, useEventListener } from '@vueuse/core'
import type { TabPaneName } from 'element-plus'

import { useSortType } from './useSortType'
import { DetectChanges } from '@bindings/musicassist/internal/services/sourceservice'
import { useImages } from './useImages'

const updateViewRef = ref<InstanceType<typeof UpdateView> | null>(null)
const { sortTypeList, selectedSortType, sortTypeInit } = useSortType()
const {
  imgInfo,
  imgs,
  sources,
  curIndex,
  imagesInit,
  changeSources,
  updateImages,
  selectImageByPath,
  prevImg,
  nextImg,
} = useImages()

const dialogVisible = ref(false)
const activeName = ref('sources')

const handleChose = (path: string) => {
  selectImageByPath(path)
  dialogVisible.value = false
}

const handleTabClick = (activeName: TabPaneName, _: TabPaneName) => {
  if (activeName === 'close') {
    dialogVisible.value = false
    return false
  }
  return true
}

async function showDialog() {
  if (!dialogVisible.value) {
    dialogVisible.value = true
  }
}

const handleChangeSortType = async () => {
  await updateImages()
}

/** 点击标题栏时会触发两次 */
const focusUpdateImages = useDebounceFn(async () => {
  if (await DetectChanges()) {
    await updateImages()
    ElMessage.success('Images updated')
  }
}, 500)
useEventListener('focus', async () => {
  await focusUpdateImages()
})

onMounted(async () => {
  await imagesInit()
  const sortType = await GetImgSortType()
  await sortTypeInit(sortType)
  if (imgs.value.length === 0) {
    activeName.value = 'sources'
    showDialog()
  }
})
</script>
<style scoped>
:deep(.el-dialog__header) {
  padding-bottom: 0;
}
</style>
