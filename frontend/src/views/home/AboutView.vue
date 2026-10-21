<template>
  <el-scrollbar class="h-full">
    <div class="flex flex-col items-center gap-5 py-6 px-4 select-none">
      <!-- Header -->
      <div class="flex flex-col items-center gap-1">
        <div class="text-3xl font-bold">{{ staticConfig?.BuildConfig.Info.ProductName }}</div>
        <div class="text-sm text-gray-500">Numbered Musical Notation Assistant</div>
        <div
          class="cursor-pointer hover:text-green-600 text-sm mt-1"
          title="Check for updates"
          @click="emit('VersionClick')"
        >
          Version : v{{ staticConfig?.BuildConfig.Info.Version }}
        </div>
      </div>

      <div class="text-gray-700 text-center max-w-md">
        A Windows desktop application for browsing numbered musical notation images and generating
        print-ready Docx documents.
      </div>

      <el-divider class="w-80" />

      <!-- Features -->
      <div class="flex flex-col items-center gap-3 w-full max-w-md">
        <div class="text-base font-semibold text-gray-800">Features</div>
        <div class="grid grid-cols-2 gap-x-6 gap-y-2 text-sm text-gray-600 w-full">
          <div v-for="item in features" :key="item.text" class="flex items-center gap-2">
            <el-icon color="#16a34a"><component :is="item.icon" /></el-icon>
            <span>{{ item.text }}</span>
          </div>
        </div>
      </div>

      <el-divider class="w-80" />

      <!-- Tech Stack -->
      <div class="flex flex-col items-center gap-2 w-full max-w-md">
        <div class="text-base font-semibold text-gray-800">Tech Stack</div>
        <div class="flex flex-col gap-1 text-sm text-gray-600 w-full">
          <div v-for="row in techStack" :key="row.label" class="flex">
            <span class="w-20 shrink-0 text-gray-400">{{ row.label }}</span>
            <span>{{ row.value }}</span>
          </div>
        </div>
      </div>

      <el-divider class="w-80" />

      <!-- Repositories -->
      <div class="flex flex-col items-center gap-2 w-full max-w-md">
        <div class="text-base font-semibold text-gray-800">Repositories</div>
        <div class="flex flex-col gap-1 text-sm w-full">
          <div
            v-for="repo in repos"
            :key="repo.url"
            class="flex items-center gap-2 cursor-pointer hover:text-green-600 text-gray-600"
            :title="`Open ${repo.label} in browser`"
            @click="openUrl(repo.url)"
          >
            <el-icon><Link /></el-icon>
            <span class="w-28 shrink-0 text-gray-400">{{ repo.label }}</span>
            <span class="truncate">{{ repo.url }}</span>
          </div>
        </div>
      </div>

      <el-divider class="w-80" />

      <!-- Footer -->
      <div class="flex flex-col items-center gap-1 text-xs text-gray-400">
        <div>
          Licensed under the
          <span class="cursor-pointer hover:text-green-600" @click="openUrl(licenseUrl)">
            MIT License
          </span>
        </div>
        <div>{{ staticConfig?.BuildConfig.Info.Copyright }}</div>
      </div>
    </div>
  </el-scrollbar>
</template>
<script setup lang="ts">
import { GetStaticConfig } from '@bindings/musicassist/services/appservice'
import type { StaticConfig } from '@bindings/musicassist/tools'
import { Browser } from '@wailsio/runtime'
import {
  Link,
  Picture,
  FolderOpened,
  Search,
  Document,
  Download,
  Refresh,
  Monitor,
} from '@element-plus/icons-vue'

const emit = defineEmits<{
  (e: 'VersionClick'): void
}>()

const staticConfig = ref<StaticConfig | null>(null)

const features = [
  { icon: Picture, text: 'Image Viewer' },
  { icon: FolderOpened, text: 'Source Management' },
  { icon: Search, text: 'Search & Sort' },
  { icon: Document, text: 'Docx Generation' },
  { icon: Download, text: 'Remote Sync' },
  { icon: Refresh, text: 'Auto Update' },
  { icon: Monitor, text: 'Server Mode' },
]

const techStack = [
  { label: 'Platform', value: 'Windows 10 / 11 (x64)' },
  { label: 'Backend', value: 'Go 1.26, Wails v3' },
  { label: 'Frontend', value: 'Vue 3, TypeScript, Vite, Element Plus, Tailwind CSS' },
  { label: 'Build', value: 'Task, wails3 CLI, NSIS / MSIX' },
  { label: 'Updater', value: 'Gitee Releases provider' },
]

const repos = [
  { label: 'GitHub', url: 'https://github.com/wtl568594290/music-assist' },
  { label: 'Gitee', url: 'https://gitee.com/wtl1989/music-assist' },
  { label: 'Sync Server', url: 'https://github.com/wtl568594290/render-test' },
]

const licenseUrl = 'https://opensource.org/licenses/MIT'

const openUrl = (url: string) => {
  Browser.OpenURL(url)
}

onMounted(async () => {
  staticConfig.value = await GetStaticConfig()
})
</script>
<style scoped></style>
