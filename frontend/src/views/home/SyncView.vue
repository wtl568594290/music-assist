<template>
  <div class="flex flex-col gap-2">
    <div class="flex flex-row items-center gap-2 w-75">
      <el-checkbox v-model="sync" :id="id" />
      <el-input v-show="sync" v-model="remoteUrl" placeholder="Remote URL" />
      <el-button v-show="sync" @click="handleDownload" :icon="Download" />
    </div>
    <el-progress
      v-show="sync && syncUploadBody.Total"
      :percentage="syncProgressValue"
      :width="300"
      :stroke-width="2"
    />
  </div>
</template>
<script setup lang="ts">
import { Download } from '@element-plus/icons-vue'
import { DownloadRemote, GetConfigs } from '@/wailsjs/go/app/app'
import type { app } from '@/wailsjs/go/models'
import { EventsOff, EventsOn } from '@/wailsjs/runtime/runtime'
import { useThrottleFn } from '@vueuse/core'

const props = defineProps<{
  id: string
}>()
defineExpose({
  getRemoteUrl,
})

const sync = ref(true)
const remoteUrl = ref('')
const syncUploadBody = ref({
  Total: 0,
  Cnt: 0,
})

const syncProgressValue = computed(() => {
  if (syncUploadBody.value.Total === 0) {
    return 0
  }
  return Math.floor((syncUploadBody.value.Cnt / syncUploadBody.value.Total) * 100)
})

function getRemoteUrl() {
  return sync.value ? remoteUrl.value : ''
}

const throttledMsgUpload = useThrottleFn(
  (msg: app.SyncProgressBody) => {
    syncUploadBody.value = msg
  },
  100,
  true,
)

const syncMsgProgress = (msg: app.SyncProgressBody) => {
  throttledMsgUpload(msg)
}

const syncMsg = (msg: string) => {
  syncUploadBody.value = {
    Total: 0,
    Cnt: 0,
  }
  ElMessage.success(msg)
}
const syncMsgError = (msg: string) => {
  syncUploadBody.value = {
    Total: 0,
    Cnt: 0,
  }
  ElMessage.error(msg)
}

const handleDownload = async () => {
  await DownloadRemote(remoteUrl.value)
}

onMounted(async () => {
  EventsOn('sync-msg', syncMsg)
  EventsOn('sync-msg-error', syncMsgError)
  EventsOn('sync-msg-progress', syncMsgProgress)
  const cfg = await GetConfigs()
  remoteUrl.value = cfg.RemoteUrl
})

onUnmounted(() => {
  EventsOff('sync-msg')
  EventsOff('sync-msg-error')
  EventsOff('sync-msg-progress')
})
</script>
<style scoped></style>
