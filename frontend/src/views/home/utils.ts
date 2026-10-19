import { type FileInfo } from '@bindings/musicassist/services'
import pinyin from 'pinyin'
export interface Image {
  name: string
  pyletter: string
  py: string
  path: string
  base64?: string
}

// 获取中文字符串的首字母
function getFirstLetter(str: string): string {
  const pinyinArray = pinyin(str, {
    style: pinyin.STYLE_FIRST_LETTER, // 只获取首字母
    heteronym: false, // 不处理多音字
  })

  return pinyinArray.flat().join('').toLowerCase()
}
// 获取中文字符串的拼音
function getPinyin(str: string): string {
  const pinyinArray = pinyin(str, {
    style: pinyin.STYLE_NORMAL, // 获取完整拼音
    heteronym: false, // 不处理多音字
  })

  return pinyinArray.flat().join('').toLowerCase()
}

export function procImgFilePaths(list: FileInfo[]): Image[] {
  return list.map((item) => {
    const name = item.BaseName
    return {
      name: name,
      pyletter: getFirstLetter(name),
      py: getPinyin(name),
      path: item.Path,
    }
  })
}
