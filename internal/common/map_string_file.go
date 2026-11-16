package common

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"syscall"
	"time"

	"golang.org/x/text/collate"
	"golang.org/x/text/language"
)

type MapStringFileInfo map[string]FileInfo

var changeMap = map[string]FileChangeFlag{}
var changeMapMu sync.Mutex

func NewMapStringStruct() MapStringFileInfo {
	return make(MapStringFileInfo)
}

func (m MapStringFileInfo) ToSlice(sortType SortType) []FileInfo {
	slice := make([]FileInfo, 0, len(m))
	for _, info := range m {
		slice = append(slice, info)
	}
	slice = sortFiles(slice, sortType)
	return slice
}

func (m MapStringFileInfo) Add(path string) error {
	entry, err := os.Stat(path)
	if err != nil {
		return err
	}
	if entry.IsDir() {
		return errors.New("path is dir")
	}
	baseName := strings.TrimSuffix(entry.Name(), filepath.Ext(entry.Name()))
	ext := "." + strings.TrimPrefix(filepath.Ext(entry.Name()), ".")
	dir := filepath.Dir(path)
	modTime := entry.ModTime().UnixMilli()
	created := getCreateTime(entry).UnixMilli()
	size := entry.Size()
	hash := ""
	changeMapMu.Lock()
	change, ok := changeMap[path]
	if ok && change.Size == size && change.ModTime == modTime {
		hash = change.Hash
	}
	changeMapMu.Unlock()

	if hash == "" {
		hash, err = fileSha256(path)
		if err != nil {
			return err
		}
		changeMapMu.Lock()
		changeMap[path] = FileChangeFlag{
			ModTime: modTime,
			Size:    size,
			Hash:    hash,
		}
		changeMapMu.Unlock()
	}
	m[path] = FileInfo{
		Path:     path,
		Dir:      dir,
		BaseName: baseName,
		Ext:      ext,
		Created:  created,
		FileChangeFlag: FileChangeFlag{
			ModTime: modTime,
			Size:    size,
			Hash:    hash,
		},
	}
	return nil
}

func fileSha256(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()
	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", hash.Sum(nil)), nil
}

func getCreateTime(info os.FileInfo) time.Time {
	stat := info.Sys().(*syscall.Win32FileAttributeData)
	return time.Unix(0, stat.CreationTime.Nanoseconds())
}

func sortFiles(files []FileInfo, sortType SortType) []FileInfo {
	// 用 Path 作为兜底次级排序键，确保排序结果完全确定
	// （因为 Go map 遍历顺序不确定，非稳定排序+重复键会导致同一组文件排序结果不一致）
	pathLess := func(i, j int) bool {
		return files[i].Path < files[j].Path
	}

	switch sortType {
	case SortByCreated:
		sort.SliceStable(files, func(i, j int) bool {
			if files[i].Created != files[j].Created {
				return files[i].Created < files[j].Created
			}
			return pathLess(i, j)
		})
	case SortByCreatedDesc:
		sort.SliceStable(files, func(i, j int) bool {
			if files[i].Created != files[j].Created {
				return files[i].Created > files[j].Created
			}
			return pathLess(i, j)
		})
	case SortByModTime:
		sort.SliceStable(files, func(i, j int) bool {
			if files[i].ModTime != files[j].ModTime {
				return files[i].ModTime < files[j].ModTime
			}
			return pathLess(i, j)
		})
	case SortByModTimeDesc:
		sort.SliceStable(files, func(i, j int) bool {
			if files[i].ModTime != files[j].ModTime {
				return files[i].ModTime > files[j].ModTime
			}
			return pathLess(i, j)
		})
	case SortByNameDesc:
		collator := collate.New(language.SimplifiedChinese, collate.Numeric, collate.IgnoreCase)
		sort.SliceStable(files, func(i, j int) bool {
			cmp := collator.CompareString(files[i].BaseName, files[j].BaseName)
			if cmp != 0 {
				return cmp > 0
			}
			return pathLess(i, j)
		})
	default:
		collator := collate.New(language.SimplifiedChinese, collate.Numeric, collate.IgnoreCase)
		sort.SliceStable(files, func(i, j int) bool {
			cmp := collator.CompareString(files[i].BaseName, files[j].BaseName)
			if cmp != 0 {
				return cmp < 0
			}
			return pathLess(i, j)
		})
	}
	return files
}
