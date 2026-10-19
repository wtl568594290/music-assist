<template>
  <div class="flex flex-col items-center justify-center">
    <el-form label-width="auto">
      <el-form-item label="Page Size">
        <el-select v-model="pageSize" placeholder="PageSize" style="width: 240px">
          <el-option
            v-for="(item, idx) in pageSizeOptions"
            :key="idx"
            :label="item"
            :value="item"
          />
        </el-select>
      </el-form-item>
      <el-form-item style="width: 360px" label="Page Margins">
        <div class="w-full flex flex-row items-center justify-center">
          <el-input-number
            style="width: 60px"
            v-model="pageCfg.top"
            :min="0"
            :max="50"
            :step="1"
            :controls="false"
            value-on-clear="min"
          />
        </div>
        <div class="w-full flex flex-row justify-around items-center">
          <el-input-number
            style="width: 60px"
            v-model="pageCfg.left"
            :min="5"
            :max="50"
            :step="1"
            :controls="false"
            value-on-clear="min"
          />
          <div>mm</div>
          <el-input-number
            style="width: 60px"
            v-model="pageCfg.right"
            :min="0"
            :max="50"
            :step="1"
            :controls="false"
            value-on-clear="min"
          />
        </div>
        <div class="w-full flex flex-row items-center justify-center">
          <el-input-number
            style="width: 60px"
            v-model="pageCfg.bottom"
            :min="0"
            :max="50"
            :step="1"
            :controls="false"
            value-on-clear="min"
          />
        </div>
      </el-form-item>
      <el-form-item label="Catalog Cols">
        <el-input-number
          style="width: 120px"
          v-model="catalogCols"
          :min="1"
          :max="10"
          :step="1"
          value-on-clear="min"
        />
      </el-form-item>
      <el-form-item label="Catalog FontSize">
        <el-input-number
          style="width: 120px"
          v-model="catalogFontSize"
          :min="1"
          :max="30"
          :step="1"
          value-on-clear="min"
        />
      </el-form-item>
      <el-form-item label="Page Number FontSize">
        <el-input-number
          style="width: 120px"
          v-model="pageNumberFontSize"
          :min="1"
          :max="30"
          :step="1"
          value-on-clear="min"
        />
      </el-form-item>
      <el-form-item label="Sort By">
        <el-select v-model="selectedSortType" placeholder="Sort Type" class="flex-1">
          <el-option v-for="item in sortTypeList" :key="item" :label="item" :value="item" />
        </el-select>
      </el-form-item>
      <el-form-item label="Split By Dir">
        <el-switch v-model="splitByDir" />
      </el-form-item>
      <el-form-item label="Sync" for="sync-checkbox">
        <sync-view id="sync-checkbox" ref="syncViewRef" />
      </el-form-item>
    </el-form>
    <div class="flex flex-row gap-2">
      <el-button v-show="latestDocsPath.length" type="danger" @click="handleRevoke"
        >Revoke</el-button
      >
      <el-button :loading="isGenerating" type="primary" @click="handleGenerate"
        >Generate Music Doc</el-button
      >
    </div>
  </div>
</template>
<script setup lang="ts">
import { useSortType } from './useSortType'
import SyncView from './SyncView.vue'
import { GenerateDoc, OpenDoc, RemoveFile } from '@bindings/musicassist/services/documentservice'
import type { SortType } from '@bindings/musicassist/tools'

const syncViewRef = ref<InstanceType<typeof SyncView> | null>(null)
const { sortTypeList, selectedSortType, sortTypeInit } = useSortType()
const pageCfg = reactive({
  top: 5,
  bottom: 3,
  left: 20,
  right: 5,
})
const pageSizeOptions = ['A3', 'A4', 'A5']
const pageSize = ref('A4')
const catalogCols = ref(3)
const pageNumberFontSize = ref(16)
const catalogFontSize = ref(12)
const isGenerating = ref(false)
const latestDocsPath: Ref<string[]> = ref([])
const splitByDir = ref(false)

const handleRevoke = () => {
  if (latestDocsPath.value.length) {
    RemoveFile(latestDocsPath.value)
      .then(() => {
        ElMessage.success('Revoke success')
      })
      .catch((e) => {
        ElMessage.error(e)
      })
      .finally(() => {
        latestDocsPath.value = []
      })
  }
}

const handleGenerate = () => {
  isGenerating.value = true
  GenerateDoc({
    PageSize: pageSize.value,
    PageTop: pageCfg.top,
    PageBottom: pageCfg.bottom,
    PageLeft: pageCfg.left,
    PageRight: pageCfg.right,
    CatalogCols: catalogCols.value,
    CatalogFontSize: catalogFontSize.value,
    PageNumberFontSize: pageNumberFontSize.value,
    SortType: selectedSortType.value as SortType,
    SplitByDir: splitByDir.value,
    RemoteUrl: syncViewRef.value?.getRemoteUrl() || '',
  })
    .then((paths: string[] | null) => {
      if (!paths) {
        return
      }
      if (paths.length) {
        latestDocsPath.value = paths
        const body: any[] = ['Generate success']
        paths.forEach((path) => {
          body.push(
            h('br'),
            h(
              'a',
              {
                style: 'color:var(--el-color-primary); cursor:pointer;',
                onClick: () => {
                  OpenDoc(path)
                },
              },
              path,
            ),
          )
        })
        ElMessage({
          type: 'success',
          message: h('span', { style: 'color:var(--el-color-success)' }, body),
          duration: 5000,
        })
      }
    })
    .catch((e) => {
      if (e) {
        ElMessage.error(e)
      }
    })
    .finally(() => {
      isGenerating.value = false
    })
}

onMounted(async () => {
  await sortTypeInit('')
})
</script>
<style scoped></style>
