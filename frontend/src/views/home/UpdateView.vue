<template>
  <el-dialog
    v-model="dialogVisible"
    :close-on-press-escape="false"
    :close-on-click-modal="false"
    @close="handleClose"
  >
    <template #header>
      <div class="font-bold text-2xl">
        {{ step === UpdateStep.Ready ? 'Update Ready' : 'New version available' }} - v{{
          release?.version
        }}
      </div></template
    >
    <div class="flex justify-center flex-col gap-4" v-show="release">
      <div>Release notes:</div>
      <el-scrollbar max-height="300px">
        <div class="whitespace-pre-wrap">{{ release?.notes }}</div>
      </el-scrollbar>
      <div>
        <div class="w-full flex flex-row" v-if="step !== UpdateStep.Downloading">
          <el-button
            v-show="step === UpdateStep.Available && release?.version"
            type="danger"
            class="flex-[0.5]"
            @click="handleIgnore"
            >{{ 'Ignore v' + release?.version }}</el-button
          >
          <el-button
            :disabled="step === UpdateStep.Checking"
            type="primary"
            class="flex-1"
            @click="handleDo"
            >{{ doButtonText }}</el-button
          >
        </div>
        <el-progress
          v-else
          :percentage="downloadBody.progress"
          :stroke-width="32"
          text-inside
          class="w-full"
        >
          {{ progressText }}
        </el-progress>
      </div>
    </div>
  </el-dialog>
</template>
<script setup lang="ts">
import { storage, StorageKey } from '@/utils/storage'
import { type Release } from '@bindings/github.com/wailsapp/wails/v3/pkg/updater'
import {
  CheckForUpdates,
  DownloadUpdate,
  RestartApp,
} from '@bindings/musicassist/internal/services/updateservice'
import { Events, Updater } from '@wailsio/runtime'

enum UpdateStep {
  Idle,
  Checking,
  Available,
  Downloading,
  Ready,
}
defineExpose({
  check,
})

const AUTO_CHECK_UPDATE_INTERVAL = 2 * 60 * 60 * 1000 // 2小时

const step = ref(UpdateStep.Idle)
const dialogVisible = ref(false)
const release = ref<Release | null>(null)
const downloadBody = reactive({
  progress: 0,
  speed: 0,
})

const progressText = computed(() => {
  return `${downloadBody.progress.toFixed(2)}% - ${fmtSpeed(downloadBody.speed)}`
})
const doButtonText = computed(() => {
  switch (step.value) {
    case UpdateStep.Idle:
      return 'Check for updates'
    case UpdateStep.Checking:
      return 'Checking...'
    case UpdateStep.Available:
      return 'Download it now'
    case UpdateStep.Downloading:
      return 'Downloading...'
    case UpdateStep.Ready:
      return 'Restart and update'
    default:
      return 'Error'
  }
})

function fmtSpeed(speed: number): string {
  if (speed > 1024 * 1024) {
    return `${(speed / 1024 / 1024).toFixed(2)} Mb/s`
  } else if (speed > 1024) {
    return `${(speed / 1024).toFixed(2)} Kb/s`
  }
  return `${speed} b/s`
}

async function check(manual: boolean = true) {
  if (step.value !== UpdateStep.Idle && manual && !dialogVisible.value) {
    dialogVisible.value = true
    return
  }
  step.value = UpdateStep.Checking
  try {
    const rel = await CheckForUpdates()
    release.value = rel
    let ignoreUpdateVersionList = storage.get(StorageKey.IgnoreUpdateVersionList) || []
    // manual check, ignore all ignore update version
    if (manual) {
      ignoreUpdateVersionList = []
    }
    if (rel && !ignoreUpdateVersionList.includes(rel.version)) {
      step.value = UpdateStep.Available
      dialogVisible.value = true
    } else {
      step.value = UpdateStep.Idle
      if (manual) {
        ElMessage.info('This version is the latest')
      }
    }
  } catch (err: any) {
    step.value = UpdateStep.Idle
    if (manual) {
      ElMessage.error(err.toString())
    }
  }
}

async function download() {
  if (step.value !== UpdateStep.Available) {
    return
  }
  downloadBody.progress = 0
  downloadBody.speed = 0
  step.value = UpdateStep.Downloading
  try {
    await DownloadUpdate()
    step.value = UpdateStep.Ready
  } catch (err: any) {
    step.value = UpdateStep.Idle
    ElMessage.error(err.toString())
  }
}

async function restart() {
  if (step.value !== UpdateStep.Ready) {
    return
  }
  try {
    await RestartApp()
  } catch (err: any) {
    ElMessage.error(err.toString())
  } finally {
    step.value = UpdateStep.Idle
  }
}

function autoCheckUpdate() {
  const lastCheckTime = storage.get(StorageKey.LastAutoCheckUpdateTime) || 0
  const now = Date.now()
  if (now - lastCheckTime > AUTO_CHECK_UPDATE_INTERVAL) {
    check(false)
  }
}

const handleClose = () => {
  if (step.value === UpdateStep.Available) {
    storage.set(StorageKey.LastAutoCheckUpdateTime, Date.now())
  }
}

const handleIgnore = async () => {
  if (step.value === UpdateStep.Available && release.value) {
    const ignoreUpdateVersionList = storage.get(StorageKey.IgnoreUpdateVersionList) || []
    if (!ignoreUpdateVersionList.includes(release.value.version)) {
      ignoreUpdateVersionList.push(release.value.version)
    }
    storage.set(StorageKey.IgnoreUpdateVersionList, ignoreUpdateVersionList)
    dialogVisible.value = false
  }
}

const handleDo = async () => {
  switch (step.value) {
    case UpdateStep.Idle:
      await check(true)
      break
    case UpdateStep.Available:
      await download()
      break
    case UpdateStep.Ready:
      await restart()
      break
  }
}

onMounted(() => {
  Events.On(Updater.Events.DownloadProgress, (msg) => {
    if (msg.data.total === 0) {
      downloadBody.progress = 0
    } else {
      downloadBody.progress = (msg.data.written * 100) / msg.data.total
    }
    downloadBody.speed = msg.data.rate
  })
  Events.On(Updater.Events.DownloadComplete, () => {
    downloadBody.progress = 100
  })
  step.value = UpdateStep.Idle
  autoCheckUpdate()
})

onUnmounted(() => {
  Events.Off(Updater.Events.DownloadProgress)
  Events.Off(Updater.Events.DownloadComplete)
})
</script>
<style scoped></style>
