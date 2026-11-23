import { SplitMsg } from '@bindings/events'
import {
  CloseMetronomeWindow,
  CreateMetronomeWindow,
} from '@bindings/musicassist/internal/services/metronomeservice'
import { Events } from '@wailsio/runtime'

export const useMetronomeManage = () => {
  let isSpliting = false

  const aloneMetronome = ref(false)
  const splitMetronome = async () => {
    if (isSpliting) {
      return
    }
    isSpliting = true
    if (!aloneMetronome.value) {
      aloneMetronome.value = true
      await CreateMetronomeWindow()
    } else {
      await CloseMetronomeWindow()
      aloneMetronome.value = false
    }
    isSpliting = false
  }

  onMounted(() => {
    Events.On(SplitMsg, (data) => {
      aloneMetronome.value = data.data
    })
  })

  onUnmounted(() => {
    Events.Off(SplitMsg)
  })

  return {
    aloneMetronome,
    splitMetronome,
  }
}
