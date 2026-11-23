<template>
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
</template>
<script setup lang="ts">
import { DocumentAdd, Folder, FolderAdd, Delete, Refresh } from '@element-plus/icons-vue'
import type { IconPropType } from 'element-plus/es/utils/index.mjs'
import type { ButtonType } from 'element-plus'

import {
  AddSource,
  ClearSources,
  RemoveSource,
} from '@bindings/musicassist/internal/services/sourceservice'
import type { Source } from '@bindings/musicassist/internal/common'

interface Actions {
  icon: IconPropType
  fn: () => Promise<void>
  name?: string
  info: string
  type?: ButtonType
}
const emit = defineEmits<{
  (e: 'changeSources'): void
}>()
const { sources } = defineProps<{
  sources: Source[]
}>()

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

const handleChangeSources = async (fn: () => Promise<void>) => {
  try {
    await fn()
  } catch (_) {
    return
  }
  emit('changeSources')
}
</script>
<style scoped></style>
