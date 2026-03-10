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
      <el-form-item label="Sort">
        <div class="flex flex-row gap-2 items-center justify-center">
          <el-select v-model="sortSelected" placeholder="Sort" style="width: 240px">
            <el-option
              v-for="(item, idx) in sortTypeOptions"
              :key="idx"
              :label="capitalize(item)"
              :value="item"
            />
          </el-select>
          <el-checkbox v-model="isDesc" label="Desc"></el-checkbox>
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
    </el-form>
    <div class="flex flex-row gap-2">
      <el-button v-show="latestDocPath" type="danger" @click="handleRevoke">Revoke</el-button>
      <el-button :loading="isGenerating" type="primary" @click="handleGenerate"
        >Generate Music Doc</el-button
      >
    </div>
  </div>
</template>
<script setup lang="ts">
import { GenerateDoc, OpenDoc, RemoveFile } from '@/wailsjs/go/app/app'
const pageCfg = reactive({
  top: 5,
  bottom: 3,
  left: 20,
  right: 5,
})
const pageSizeOptions = ['A3', 'A4', 'A5']
const pageSize = ref('A4')
const sortTypeOptions = ['name', 'modTime', 'created']
const sortSelected = ref('name')
const isDesc = ref(false)
const catalogCols = ref(3)
const pageNumberFontSize = ref(16)
const catalogFontSize = ref(12)
const isGenerating = ref(false)
const latestDocPath = ref('')
const handleRevoke = () => {
  if (latestDocPath.value) {
    RemoveFile(latestDocPath.value)
      .then(() => {
        ElMessage.success('Revoke success')
        latestDocPath.value = ''
      })
      .catch((e) => {
        console.log(e)
        ElMessage.error(e)
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
    SortType: sortSelected.value + (isDesc.value ? 'Desc' : ''),
    CatalogCols: catalogCols.value,
    CatalogFontSize: catalogFontSize.value,
    PageNumberFontSize: pageNumberFontSize.value,
  })
    .then((path: string) => {
      if (path) {
        latestDocPath.value = path
        ElMessage({
          type: 'success',
          message: h('span', { style: 'color:var(--el-color-success)' }, [
            'Generate success ',
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
          ]),
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

function capitalize(str: string): string {
  if (!str) return str
  return str[0].toUpperCase() + str.slice(1)
}
</script>
<style scoped></style>
