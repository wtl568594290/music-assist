<template>
  <div class="flex flex-col gap-2">
    <div class="flex flex-row items-center gap-2 w-75">
      <el-checkbox v-model="sync" :id="id" />
      <el-select
        v-show="sync"
        v-model="remoteUrl"
        placeholder="Select Remote URL"
        fit-input-width
        allow-create
        filterable
        popper-class="remote-url-popper"
      >
        <el-option
          v-for="(item, key) in remoteUrlList"
          :key="key"
          :label="item.url"
          :value="item.url"
        >
          <div class="flex flex-row items-center justify-between gap-2">
            <el-text truncated>{{ item.url }}</el-text>
            <el-button
              :icon="Delete"
              size="small"
              text
              class="shrink-0"
              @click.stop="handleDeleteOption(item)"
              @mousedown.stop
            />
          </div>
        </el-option>
      </el-select>
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
import { Delete, Download } from '@element-plus/icons-vue'
import { Events } from '@wailsio/runtime'
import { useThrottleFn } from '@vueuse/core'
import { type SyncProgressBody } from '@bindings/musicassist/internal/services'
import { DownloadRemote } from '@bindings/musicassist/internal/services/syncservice'
import { SyncMsg, SyncMsgError, SyncMsgProgress } from '@bindings/events'
import { storage, StorageKey, type StorageSchema } from '@/utils/storage'

type StorageRemoteUrl = StorageSchema[StorageKey.SyncRemoteUrlList][0]

const props = defineProps<{
  id: string
}>()

defineExpose({
  getRemoteUrl,
})

const REMOTE_URL_LIST_MAX = 10

const sync = ref(true)
const remoteUrlList = ref<StorageRemoteUrl[]>([])
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

function getStoreRemoteUrlList(): StorageRemoteUrl[] {
  return storage.get(StorageKey.SyncRemoteUrlList) || []
}

function addStoreRemoteUrlList(url: string) {
  if (!url) {
    return
  }
  const list = getStoreRemoteUrlList()
  if (!list.find((item) => item.url === url)) {
    list.push({ time: Date.now(), url })
    list.sort((a, b) => b.time - a.time)
    if (list.length > REMOTE_URL_LIST_MAX) {
      list.splice(REMOTE_URL_LIST_MAX)
    }
    storage.set(StorageKey.SyncRemoteUrlList, list)
    remoteUrlList.value = list
  }
}

const handleDeleteOption = (item: StorageRemoteUrl) => {
  if (!item || !item.url) {
    return
  }
  const list = getStoreRemoteUrlList().filter((i) => i.url !== item.url)
  list.sort((a, b) => b.time - a.time)
  storage.set(StorageKey.SyncRemoteUrlList, list)
  remoteUrlList.value = list
  remoteUrl.value = ''
}

const throttledMsgUpload = useThrottleFn(
  (msg: SyncProgressBody) => {
    syncUploadBody.value = msg
  },
  100,
  true,
)

const syncMsgProgress = (msg: SyncProgressBody) => {
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

const handleDownload = () => {
  if (!remoteUrl.value) {
    return
  }
  DownloadRemote(remoteUrl.value)
    .then(() => {
      addStoreRemoteUrlList(remoteUrl.value)
    })
    .catch((err) => {
      ElMessage.error(err)
    })
}

onMounted(async () => {
  Events.On(SyncMsg, (ev) => {
    syncMsg(ev.data)
  })
  Events.On(SyncMsgError, (ev) => {
    syncMsgError(ev.data)
  })
  Events.On(SyncMsgProgress, (ev) => {
    syncMsgProgress(ev.data)
  })
  remoteUrlList.value = getStoreRemoteUrlList()
  if (remoteUrlList.value.length) {
    remoteUrl.value = remoteUrlList.value[0].url
  }
})

onUnmounted(() => {
  Events.Off(SyncMsg)
  Events.Off(SyncMsgError)
  Events.Off(SyncMsgProgress)
})
</script>
<style scoped></style>
<style>
/* popper 由 ElSelect 内部组件渲染并 teleport 到 body，不携带本组件的 scoped 属性 */
.remote-url-popper .el-select-dropdown__item {
  padding-right: 0px;
}
</style>
