import { GetSortTypeList } from '@bindings/musicassist/services/appservice'

export function useSortType() {
  const sortTypeList = ref<string[]>([])
  const selectedSortType = ref('')

  async function sortTypeInit(selected: string) {
    sortTypeList.value = ((await GetSortTypeList()) || []).sort((a, b) =>
      a.localeCompare(b, 'zh-CN', { numeric: true }),
    )
    if (selected && sortTypeList.value.includes(selected)) {
      selectedSortType.value = selected
    } else {
      selectedSortType.value = 'Name'
    }
  }

  return {
    sortTypeList,
    selectedSortType,
    sortTypeInit,
  }
}
