import { type FileInfo } from '@bindings/musicassist/tools'
import pinyin from 'tiny-pinyin'
export interface Image {
  name: string
  pyletter: string
  py: string
  path: string
  base64?: string
}

// 获取中文字符串的拼音
function getPinyin(str: string): { firstLetter: string; all: string } {
  if (pinyin.isSupported()) {
    let firstLetter = ''
    let all = ''
    pinyin.parse(str).forEach((token) => {
      const word = token.target
      if (word.length > 0) {
        firstLetter += word[0].toLowerCase()
      }
      all += word.toLowerCase()
    })
    return { firstLetter, all }
  }
  return { firstLetter: str, all: str }
}

export function procImgFilePaths(list: FileInfo[]): Image[] {
  return list.map((item) => {
    const name = item.BaseName
    const { firstLetter, all } = getPinyin(name)
    return {
      name: name,
      pyletter: firstLetter,
      py: all,
      path: item.Path,
    }
  })
}
