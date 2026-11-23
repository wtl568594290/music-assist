<template>
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
          @click="emit('chose', item.path)"
          :type="item.path === curPath ? 'primary' : 'info'"
          class="cursor-pointer"
          size="large"
        >
          {{ item.name }}
        </el-tag>
      </div>
    </el-scrollbar>
  </div>
</template>
<script setup lang="ts">
import type { SortType } from '@bindings/musicassist/internal/common'
import type { Image } from './utils'
import { SetImgSortType } from '@bindings/musicassist/internal/services/appservice'

const selectedSortType = defineModel<SortType>()

const emit = defineEmits<{
  (e: 'change', sortType: SortType): void
  (e: 'chose', path: string): void
}>()

const { imgs, curIndex, sortTypeList } = defineProps<{
  imgs: Image[]
  curIndex: number
  sortTypeList: SortType[]
}>()
const searchText = ref('')

const filteredImgs = computed(() => {
  if (!searchText.value) {
    return imgs
  }
  return imgs.filter(
    (item) =>
      item.name.toLowerCase().includes(searchText.value.toLowerCase()) ||
      item.pyletter.toLowerCase().includes(searchText.value.toLowerCase()) ||
      item.py.toLowerCase().includes(searchText.value.toLowerCase()),
  )
})

const curPath = computed(() => imgs[curIndex]?.path || '')

const handleChangeSortType = async () => {
  if (selectedSortType.value) {
    await SetImgSortType(selectedSortType.value)
    emit('change', selectedSortType.value)
  }
}
</script>
<style scoped></style>
