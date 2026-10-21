<template>
  <el-dialog v-model="dialogVisible" :close-on-press-escape="false" :close-on-click-modal="false">
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
        <el-button
          v-if="
            step === UpdateStep.Available || step === UpdateStep.Ready || step === UpdateStep.Check
          "
          type="primary"
          class="w-full"
          @click="handleDo"
          >{{ doButtonText }}</el-button
        >
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
import { Stage, type Release } from '@bindings/github.com/wailsapp/wails/v3/pkg/updater'
import {
  CheckForUpdates,
  DownloadUpdate,
  RestartApp,
} from '@bindings/musicassist/services/updateservice'
import { Events, Updater } from '@wailsio/runtime'
enum UpdateStep {
  Check,
  Available,
  Download,
  Verify,
  Install,
  Ready,
}
defineExpose({
  manualCheckUpdate,
})

const AUTO_CHECK_UPDATE_INTERVAL = 2 * 60 * 60 * 1000 // 2小时

const step = ref(UpdateStep.Check)
const dialogVisible = ref(false)
const release = ref<Release | null>(null)
const downloadBody = reactive({
  progress: 0,
  speed: 0,
})

const progressText = computed(() => {
  let res = ''
  switch (step.value) {
    case UpdateStep.Download:
      res = `${downloadBody.progress.toFixed(2)}% - ${fmtSpeed(downloadBody.speed)}`
      break
    case UpdateStep.Verify:
      res = 'Verifying...'
      break
    case UpdateStep.Install:
      res = 'Installing...'
      break
  }
  return res
})
const doButtonText = computed(() => {
  switch (step.value) {
    case UpdateStep.Check:
      return 'Check for updates'
    case UpdateStep.Available:
      return 'Download it now'
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

/** Manual check update */
async function manualCheckUpdate() {
  if (step.value === UpdateStep.Check) {
    await CheckForUpdates()
  } else {
    if (!dialogVisible.value) {
      dialogVisible.value = true
    }
  }
}

const handleDo = async () => {
  if (step.value === UpdateStep.Check) {
    await CheckForUpdates()
  }
  if (step.value === UpdateStep.Available && release.value) {
    await DownloadUpdate()
  }
  if (step.value === UpdateStep.Ready && release.value) {
    await RestartApp()
  }
}

function autoCheckUpdate() {
  const raw = localStorage.getItem('lastCheckTime')
  const lastCheckTime = raw === null ? 0 : Number(raw)
  const now = Date.now()
  if (now - lastCheckTime > AUTO_CHECK_UPDATE_INTERVAL) {
    localStorage.setItem('lastCheckTime', now.toString())
    CheckForUpdates()
  }
}

onMounted(() => {
  Events.On(Updater.Events.Error, (msg) => {
    if (msg.data.stage !== Stage.StageCheck) {
      ElMessage.error(msg.data.message)
      return
    }
    step.value = UpdateStep.Check
    release.value = null
    downloadBody.progress = 0
    downloadBody.speed = 0
  })

  Events.On(Updater.Events.UpdateAvailable, (msg) => {
    step.value = UpdateStep.Available
    release.value = msg.data
    if (release.value) {
      dialogVisible.value = true
    }
  })
  Events.On(Updater.Events.DownloadStarted, () => {
    step.value = UpdateStep.Download
    downloadBody.progress = 0
    downloadBody.speed = 0
  })
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
  Events.On(Updater.Events.Verifying, () => {
    step.value = UpdateStep.Verify
  })
  Events.On(Updater.Events.Installing, () => {
    step.value = UpdateStep.Install
  })
  Events.On(Updater.Events.UpdateReady, () => {
    step.value = UpdateStep.Ready
    if (!dialogVisible.value) {
      dialogVisible.value = true
    }
  })

  step.value = UpdateStep.Check
  autoCheckUpdate()
})

onUnmounted(() => {
  Events.Off(Updater.Events.Error)
  Events.Off(Updater.Events.UpdateAvailable)
  Events.Off(Updater.Events.DownloadStarted)
  Events.Off(Updater.Events.DownloadProgress)
  Events.Off(Updater.Events.DownloadComplete)
  Events.Off(Updater.Events.Verifying)
  Events.Off(Updater.Events.Installing)
  Events.Off(Updater.Events.UpdateReady)
})
</script>
<style scoped></style>
