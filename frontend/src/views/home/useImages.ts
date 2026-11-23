import type { Source } from '@bindings/musicassist/internal/common'
import { procImgFilePaths, type Image } from './utils'
import {
  GetBase64,
  GetImages,
  GetSources,
} from '@bindings/musicassist/internal/services/sourceservice'
import { storage, StorageKey } from '@/utils/storage'

export interface ImageInfo {
  name: string
  base64: string
  index: number
  total: number
}

export function useImages() {
  let mountedComplete = false
  const curIndex = ref(-1)
  const sources = ref<Source[]>([])
  const imgs = ref<Image[]>([])
  const imgInfo = ref<ImageInfo | undefined>(undefined)

  async function changeSources() {
    sources.value = ((await GetSources()) || []).sort(
      (a, b) =>
        Number(b.IsDir) - Number(a.IsDir) ||
        a.Path.localeCompare(b.Path, 'zh-CN', { numeric: true }),
    )
    await updateImages()
  }

  async function updateImages() {
    const lastPath = imgs.value[curIndex.value]?.path || ''
    const list = (await GetImages()) || []
    imgs.value = procImgFilePaths(list)
    if (imgs.value.length > 0) {
      if (curIndex.value === -1) {
        curIndex.value = 0
      } else {
        // 触发 watch 变化
        curIndex.value = -1
        await nextTick()
        if (lastPath) {
          const lastIndex = imgs.value.findIndex((item) => item.path === lastPath)
          if (lastIndex !== -1) {
            curIndex.value = lastIndex
          } else {
            curIndex.value = 0
          }
        } else {
          curIndex.value = 0
        }
      }
    } else {
      curIndex.value = -1
    }
  }

  function selectImageByPath(path: string) {
    curIndex.value = imgs.value.findIndex((item) => item.path === path)
    if (curIndex.value === -1) {
      curIndex.value = 0
    }
  }

  function prevImg() {
    if (curIndex.value > 0) {
      curIndex.value--
    }
  }

  function nextImg() {
    if (curIndex.value < imgs.value.length - 1) {
      curIndex.value++
    }
  }

  watch(curIndex, async (newVal) => {
    const newImg = imgs.value[newVal]
    if (newVal !== -1 && newImg && newImg.name) {
      if (!newImg.base64) {
        let base64 = ''
        try {
          base64 = (await GetBase64(newImg.path)) || ''
        } catch {
          // 加载失败时保留空串，切回该图时会重试
          base64 = ''
        }
        // 等待期间已切图或列表被刷新，丢弃过期结果
        if (curIndex.value !== newVal || imgs.value[newVal] !== newImg) return
        newImg.base64 = base64
      }
      imgInfo.value = {
        name: newImg.name,
        base64: newImg.base64 || '',
        index: newVal,
        total: imgs.value.length,
      }
    } else {
      imgInfo.value = undefined
    }

    if (mountedComplete) {
      storage.set(StorageKey.LatestImageIndex, newVal)
    }
  })

  async function imagesInit() {
    mountedComplete = false
    const latestIndex = storage.get(StorageKey.LatestImageIndex) ?? -1
    await changeSources()
    if (imgs.value.length > 0) {
      if (latestIndex !== -1 && curIndex.value !== latestIndex) {
        await nextTick()
        curIndex.value = latestIndex
      }
    }
    mountedComplete = true
  }
  // onMounted(async () => {
  // })

  return {
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
  }
}
