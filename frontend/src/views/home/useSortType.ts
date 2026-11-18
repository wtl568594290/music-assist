import { SortType } from '@bindings/musicassist/internal/common'
import { GetSortTypeList } from '@bindings/musicassist/internal/services/appservice'

export function useSortType() {
  const sortTypeList = ref<SortType[]>([])
  const selectedSortType = ref<SortType>(SortType.SortByName)

  async function sortTypeInit(selected: SortType) {
    sortTypeList.value = ((await GetSortTypeList()) || []).sort((a, b) =>
      a.localeCompare(b, 'zh-CN', { numeric: true }),
    )
    if (selected && sortTypeList.value.includes(selected)) {
      selectedSortType.value = selected
    } else {
      selectedSortType.value = SortType.SortByName
    }
  }

  return {
    sortTypeList,
    selectedSortType,
    sortTypeInit,
  }
}
