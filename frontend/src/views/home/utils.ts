import pinyin from 'pinyin'
export interface Image {
  name: string
  pyletter: string
  py: string
  path: string
  base64?: string
}

function getFileNameWithoutExtension(path: string): string {
  // 处理不同操作系统的路径分隔符（Windows: \, Unix: /）
  const normalizedPath = path.replace(/\\/g, '/')

  // 截取最后一个 / 后的部分（获取文件名带扩展名）
  const fileNameWithExt = normalizedPath.split('/').pop() || ''

  // 找到最后一个 . 的位置，截取前面的部分
  const lastDotIndex = fileNameWithExt.lastIndexOf('.')
  if (lastDotIndex === -1) {
    // 没有扩展名的情况
    return fileNameWithExt
  }

  return fileNameWithExt.substring(0, lastDotIndex)
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

export function procImgFilePaths(list: string[]): Image[] {
  return list
    .map((item) => {
      const name = getFileNameWithoutExtension(item)
      return {
        name: name,
        pyletter: getFirstLetter(name),
        py: getPinyin(name),
        path: item,
      }
    })
    .sort((a, b) => a.pyletter.localeCompare(b.pyletter))
}
