<template>
  <div class="w-full h-full relative">
    <div
      class="absolute inset-0 flex justify-center items-center select-none"
      @dblclick="handlePanelDblclick"
    >
      <div class="fit-box overflow-hidden relative">
        <!-- 包裹层按 pan 的固有比例定宽，mask 用 w-full 自动跟随表盘宽度 -->
        <div class="absolute inset-0 flex justify-center">
          <div class="pan-box relative">
            <img src="/pan.webp" class="block h-full w-full object-contain" />
            <img src="/mask.webp" class="absolute bottom-0 left-0 z-10 w-full" />
          </div>
        </div>
        <div class="absolute inset-0 flex justify-center items-center">
          <div
            ref="refNeedle"
            class="w-[12%] h-full relative"
            :style="{ animationDuration: `${animationDuration}s` }"
          >
            <img
              src="/needle.webp"
              class="absolute bottom-[1.5%] left-0 h-auto w-full object-contain drop-shadow-[-15px_6px_12px_rgba(0,0,0,0.35)]"
            />
          </div>
        </div>
        <div
          class="text-stroke-black absolute justify-center items-center"
          :style="{ '--font-zoom': fontZoom, ...positions[fontPosition] }"
        >
          {{ showText }}
        </div>
        <div
          class="info-label text-white absolute right-2 top-2 cursor-pointer"
          @click="handleShowSettings"
        >
          {{ bpm }} / {{ beat }}
        </div>
      </div>
      <!-- 主窗口内居中显示播放/暂停按钮，悬停显示、暂停时常显；独立窗口不显示 -->
      <div
        v-if="!isIndependent"
        class="pointer-events-none absolute inset-0 flex items-center justify-center"
      >
        <div
          class="pointer-events-auto flex h-24 w-28 cursor-pointer items-center justify-center rounded-2xl bg-black/30 text-white transition-opacity duration-200 hover:opacity-100"
          :class="isPlaying ? 'opacity-0' : 'opacity-100'"
          @click.stop="togglePlay"
          @dblclick.stop
        >
          <el-icon :size="60">
            <VideoPause v-if="isPlaying" />
            <VideoPlay v-else />
          </el-icon>
        </div>
      </div>
    </div>
    <el-dialog v-model="isDialogVisible" title="Settings" @close="updateParams">
      <el-scrollbar max-height="60vh" view-class="pr-5">
        <el-form label-width="auto">
          <el-form-item label="Presets">
            <el-button-group>
              <el-button v-for="(p, idx) in presets" :key="idx" @click="handleLoadPresets(p)">
                {{ p.bpm + '/' + p.beat }}
              </el-button>
            </el-button-group>
          </el-form-item>
          <el-form-item label="Recent">
            <el-button-group>
              <el-button v-for="(p, idx) in presetsHis" :key="idx" @click="handleLoadPresets(p)">
                {{ p.bpm + '/' + p.beat }}
              </el-button>
            </el-button-group>
          </el-form-item>
          <el-form-item label="BPM">
            <el-select v-model="bpm" filterable allow-create placeholder="BPM">
              <el-option
                v-for="(item, index) in bpmPresets"
                :key="index"
                :label="item"
                :value="item"
              />
            </el-select>
          </el-form-item>
          <el-form-item label="Beats">
            <el-input-number v-model="beat" :min="1" :max="16" />
          </el-form-item>
          <el-form-item label="Sound">
            <el-select v-model="audioMode" placeholder="Sound">
              <el-option
                v-for="(item, index) in modes"
                :key="index"
                :label="item.label"
                :value="item.value"
              />
            </el-select>
          </el-form-item>
          <el-form-item label="Font size">
            <el-select v-model="fontZoom" placeholder="Font size">
              <el-option
                v-for="(item, index) in fontSizeZooms"
                :key="index"
                :label="item"
                :value="item"
              />
            </el-select>
          </el-form-item>
          <el-form-item label="Position">
            <el-select v-model="fontPosition" placeholder="Position">
              <el-option
                v-for="(item, index) in positionOptions"
                :key="index"
                :label="item.label"
                :value="item.value"
              />
            </el-select>
          </el-form-item>
          <el-form-item label="Volume">
            <el-slider v-model="volume" />
          </el-form-item>
        </el-form>
      </el-scrollbar>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { Window } from '@wailsio/runtime'
import { IsMetronomeIndependentWindow } from '@bindings/musicassist/internal/services/metronomeservice'
import { VideoPause, VideoPlay } from '@element-plus/icons-vue'

const emit = defineEmits<{
  (e: 'split'): void
}>()

/** 是否运行在独立的节拍器窗口中 */
const isIndependent = ref(false)

interface AudioInfo {
  audio: HTMLAudioElement
  duration: number
}
type AudioMode = 'human' | 'di' | 'machine'
type FontPosition = 'leftTop' | 'center' | 'bottomCenter' | 'topCenter'
let audioInfoList: AudioInfo[] = []
const refNeedle = ref<HTMLElement | null>(null)
let isForward = true
const isDialogVisible = ref(false)
const isPlaying = ref(false)
/** 打开设置弹窗前是否处于播放状态，关闭时恢复 */
let wasPlayingBeforeSettings = false
const presetsHis: Ref<{ bpm: string; beat: number }[]> = ref([])
const presets = [
  { bpm: '60', beat: 4 },
  { bpm: '120', beat: 8 },
  { bpm: '150', beat: 10 },
  { bpm: '180', beat: 12 },
  { bpm: '210', beat: 14 },
  { bpm: '240', beat: 16 },
]
const bpm = ref('60')
const bpmPresets = ['30', '40', '60', '80', '100', '120', '140', '160', '180', '210', '240']
const cnt = ref(0)
const fontSizeZooms = [0, 0.5, 1, 1.5, 2]
const fontZoom = ref(1)
const beat = ref(4)
let timer: number | null = null
let beatTimer: number | null = null
const audioMode: Ref<AudioMode> = ref('human')
const modes = [
  {
    value: 'human',
    label: 'Voice',
  },
  {
    value: 'di',
    label: 'Beep',
  },
  {
    value: 'machine',
    label: 'Machine',
  },
]
const volume = ref(100)
const positions = {
  leftTop: {
    left: '1rem',
    top: '1rem',
    transform: 'translate(0, 0)',
  },
  center: {
    left: '50%',
    top: '50%',
    transform: 'translate(-50%, -50%)',
  },
  bottomCenter: {
    left: '50%',
    bottom: '1rem',
    transform: 'translate(-50%, 0)',
  },
  topCenter: {
    left: '50%',
    top: '1rem',
    transform: 'translate(-50%, 0)',
  },
}

const positionOptions = [
  {
    label: 'Top Left',
    value: 'leftTop',
  },
  {
    label: 'Center',
    value: 'center',
  },
  {
    label: 'Bottom Center',
    value: 'bottomCenter',
  },
  {
    label: 'Top Center',
    value: 'topCenter',
  },
]

const fontPosition: Ref<FontPosition> = ref('center')

let isOnTop = false
const handlePanelDblclick = async () => {
  // 主窗口内双击面板分离出独立窗口；独立窗口内双击切换置顶
  if (!isIndependent.value) {
    emit('split')
    return
  }
  isOnTop = !isOnTop
  await Window.SetAlwaysOnTop(isOnTop)
  await Window.SetTitle(isOnTop ? 'Metronome - Always on top' : 'Metronome')
  ElMessage({
    message: isOnTop ? 'Always on top' : 'Not always on top',
    type: 'success',
  })
}
const togglePlay = () => {
  if (isPlaying.value) {
    stop()
  } else {
    start()
  }
}

const handleShowSettings = () => {
  wasPlayingBeforeSettings = isPlaying.value
  isDialogVisible.value = true
  stop()
}

/** 用定时器驱动节拍，窗口最小化或面板隐藏（CSS 动画不渲染、不触发 animationend）时仍能继续播放 */
const nextBeat = () => {
  isForward = !isForward
  if (refNeedle.value) {
    refNeedle.value.classList.remove('animate-rotate-forward', 'animate-rotate-backward')
    // Reapply the class to trigger the animation again
    refNeedle.value.classList.add(isForward ? 'animate-rotate-forward' : 'animate-rotate-backward')
  }
  if (++cnt.value >= beat.value) {
    cnt.value = 0
  }

  playNext((cnt.value + 1) % beat.value)
  beatTimer = setTimeout(nextBeat, animationDuration.value * 1000)
}

const animationDuration = computed(() => {
  return 60 / parseInt(bpm.value)
})

const showText = computed(() => {
  if (audioMode.value === 'human' && cnt.value >= 10) {
    return String.fromCharCode(65 + cnt.value - 10)
  }
  return `${cnt.value + 1}`
})
const updateParams = () => {
  localStorage.setItem('bpm', bpm.value)
  localStorage.setItem('fontZoom', fontZoom.value.toString())
  localStorage.setItem('fontPosition', fontPosition.value)
  localStorage.setItem('beat', beat.value.toString())
  localStorage.setItem('audioMode', audioMode.value)
  localStorage.setItem('volume', volume.value.toString())
  if (presetsHis.value.find((p) => p.bpm === bpm.value && p.beat === beat.value) === undefined) {
    const newPresetsHis = presetsHis.value
    newPresetsHis.push({ bpm: bpm.value, beat: beat.value })
    if (newPresetsHis.length > 5) {
      newPresetsHis.shift()
    }
    presetsHis.value = newPresetsHis
  }
  localStorage.setItem('presetsHis', JSON.stringify(presetsHis.value))

  // 独立窗口没有播放/暂停按钮，只有关闭设置时恢复播放；主窗口内保持原有播放状态
  if (wasPlayingBeforeSettings) {
    start()
  }
}

function loadMeta(src: string): Promise<AudioInfo> {
  return new Promise((resolve, reject) => {
    const audio = new Audio(src)
    audio.addEventListener('loadedmetadata', () => {
      if (isNaN(audio.duration)) {
        reject(new Error('Failed to load audio'))
      }
      resolve({ audio, duration: audio.duration })
    })
    audio.addEventListener('error', () => {
      reject(new Error('Failed to load audio'))
    })
    audio.volume = volume.value / 100
    audio.load()
  })
}

async function loadAudio(mode: AudioMode) {
  audioInfoList = []
  try {
    switch (mode) {
      case 'human':
        for (let i = 1; i <= 16; i++) {
          audioInfoList.push(await loadMeta(`/audio/${i}.mp3`))
        }
        break
      case 'di':
        for (let i = 1; i <= 2; i++) {
          audioInfoList.push(await loadMeta(`/audio/di${i}.mp3`))
        }
        break
      case 'machine':
        audioInfoList.push(await loadMeta(`/audio/machine.mp3`))
        break
    }
  } catch {
    ElMessage.error('Failed to load audio')
    audioInfoList = []
  }
}

function playNext(index: number) {
  let nextIndex = index
  switch (audioMode.value) {
    case 'di':
      nextIndex = index == 0 ? 0 : 1
      break
    case 'machine':
      nextIndex = 0
      break
  }
  if (timer) {
    clearTimeout(timer)
  }
  timer = setTimeout(
    () => {
      audioInfoList[nextIndex].audio.play()
    },
    (animationDuration.value - audioInfoList[nextIndex].duration / 2) * 1000,
  )
}

function start() {
  isPlaying.value = true
  loadAudio(audioMode.value).then(() => {
    if (!isPlaying.value || !refNeedle.value) {
      return
    }
    refNeedle.value.classList.add('animate-rotate-forward')
    audioInfoList[0].audio.play()
    playNext(audioMode.value == 'machine' ? 0 : 1)
    beatTimer = setTimeout(nextBeat, animationDuration.value * 1000)
  })
}

function stop() {
  isPlaying.value = false
  if (timer) {
    clearTimeout(timer)
  }
  if (beatTimer) {
    clearTimeout(beatTimer)
  }
  audioInfoList.forEach(({ audio }) => {
    audio.pause()
    audio.currentTime = 0
  })
  if (refNeedle.value) {
    refNeedle.value.classList.remove('animate-rotate-forward', 'animate-rotate-backward')
  }
  cnt.value = 0
  isForward = true
}

const handleLoadPresets = (p: { bpm: string; beat: number }) => {
  bpm.value = p.bpm
  beat.value = p.beat
}

const keyWatcher = (e: KeyboardEvent) => {
  if (isDialogVisible.value) {
    return
  }
  // 非独立窗口不响应 ESC（交由主窗口处理），空格仍可打开设置
  if (e.key === ' ' || (e.key === 'Escape' && isIndependent.value)) {
    handleShowSettings()
  }
}

onMounted(async () => {
  window.addEventListener('keydown', keyWatcher)
  presetsHis.value = JSON.parse(localStorage.getItem('presetsHis') || '[]')
  bpm.value = localStorage.getItem('bpm') || '60'
  fontZoom.value = parseFloat(localStorage.getItem('fontZoom') || '1')
  beat.value = parseInt(localStorage.getItem('beat') || '4')
  audioMode.value = (localStorage.getItem('audioMode') as AudioMode) || 'human'
  volume.value = parseInt(localStorage.getItem('volume') || '100')
  fontPosition.value = (localStorage.getItem('fontPosition') as FontPosition) || 'center'

  isIndependent.value = await IsMetronomeIndependentWindow()
})

// 主窗口内默认暂停；独立窗口打开后自动开始播放
watch(isIndependent, (independent) => {
  if (independent && !isPlaying.value) {
    start()
  }
})

onUnmounted(() => {
  if (timer) {
    clearTimeout(timer)
  }
  if (beatTimer) {
    clearTimeout(beatTimer)
  }
  audioInfoList.forEach(({ audio }) => audio.pause())
  window.removeEventListener('keydown', keyWatcher)
})
</script>

<style scoped lang="scss">
@use 'sass:math';

$angle: 25deg;
$ratio: 2 * math.tan($angle); // 自动转弧度

.fit-box {
  aspect-ratio: $ratio;
  /* 作为字号计算的查询容器（cqw 相对自身宽度） */
  container-type: inline-size;
}

/* 当容器比较“宽”，让 fit-box 的宽度占满 */
@media (min-aspect-ratio: $ratio) {
  .fit-box {
    height: 100%;
    width: auto;
  }
}

/* 当容器比较“高”，让 fit-box 的高度占满 */
@media (max-aspect-ratio: $ratio) {
  .fit-box {
    width: 100%;
    height: auto;
  }
}

/* pan.webp 的固有宽高比（1400 x 2435），包裹层据此定宽 */
.pan-box {
  height: 100%;
  aspect-ratio: math.div(1400, 2435);
}

@keyframes rotate-forward {
  0% {
    transform: rotate(-$angle);
    transform-origin: bottom center;
  }
  100% {
    transform: rotate($angle);
    transform-origin: bottom center;
  }
}

@keyframes rotate-backward {
  0% {
    transform: rotate($angle);
    transform-origin: bottom center;
  }
  100% {
    transform: rotate(-$angle);
    transform-origin: bottom center;
  }
}

.animate-rotate-forward {
  animation: rotate-forward 1 forwards cubic-bezier(0.42, 0, 0.58, 1);
}

.animate-rotate-backward {
  animation: rotate-backward 1 forwards cubic-bezier(0.42, 0, 0.58, 1);
}

.text-stroke-black {
  /* 字号跟随容器宽度：容器宽度的 1/4，限制在 36~384px，再乘以缩放系数 */
  font-size: calc(clamp(36px, 25cqw, 384px) * var(--font-zoom, 1));
  line-height: 1;
  -webkit-text-stroke-width: calc(1em / 48);
  -webkit-text-stroke-color: black;
  color: white;
}

/* 右上角参数文字，随容器宽度缩放 */
.info-label {
  font-size: clamp(12px, 4cqw, 32px);
}

.needle-shadow {
  position: relative;
}

.needle-shadow::after {
  content: '';
  position: absolute;
  bottom: 1.5%;
  left: 0;
  width: 100%;
  height: 100%;
  background: url('/needle.webp') no-repeat center center;
  background-size: contain;
  color: black;
  filter: blur(8px) opacity(0.55);
  transform: translate(-15px, 10px) scale(1.02); /* 阴影偏移方向和大小 */
  z-index: -1;
  pointer-events: none;
}
</style>
