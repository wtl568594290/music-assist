import type { SortType } from '@bindings/musicassist/internal/common'

const PREFIX = 'ma_'

export enum StorageKey {
  LatestImageIndex = 'latestImageIndex',
  LastAutoCheckUpdateTime = 'lastAutoCheckUpdateTime',
  SyncRemoteUrlList = 'syncRemoteUrlList',
  DocSortType = 'docSortType',
}

export interface StorageSchema {
  [StorageKey.LatestImageIndex]: number
  [StorageKey.LastAutoCheckUpdateTime]: number
  [StorageKey.SyncRemoteUrlList]: {
    time: number
    url: string
  }[]
  [StorageKey.DocSortType]: SortType
}

function getKey(key: StorageKey): string {
  return `${PREFIX}${key}`
}

export const storage = {
  get<K extends StorageKey>(key: K): StorageSchema[K] | undefined {
    try {
      const value = localStorage.getItem(getKey(key))
      return value === null ? undefined : JSON.parse(value)
    } catch {
      return undefined
    }
  },

  set<K extends StorageKey>(key: K, value: StorageSchema[K]): boolean {
    try {
      localStorage.setItem(getKey(key), JSON.stringify(value))
      return true
    } catch {
      return false
    }
  },

  remove(key: StorageKey): void {
    localStorage.removeItem(getKey(key))
  },

  has(key: StorageKey): boolean {
    return localStorage.getItem(getKey(key)) !== null
  },

  clear(): void {
    Object.values(StorageKey).forEach((key) => {
      localStorage.removeItem(getKey(key))
    })
  },
}
