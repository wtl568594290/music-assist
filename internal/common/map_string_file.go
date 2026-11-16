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
	switch sortType {
	case SortByCreated:
		sort.Slice(files, func(i, j int) bool {
			return files[i].Created < files[j].Created
		})
	case SortByCreatedDesc:
		sort.Slice(files, func(i, j int) bool {
			return files[i].Created > files[j].Created
		})
	case SortByModTime:
		sort.Slice(files, func(i, j int) bool {
			return files[i].ModTime < files[j].ModTime
		})
	case SortByModTimeDesc:
		sort.Slice(files, func(i, j int) bool {
			return files[i].ModTime > files[j].ModTime
		})
	case SortByNameDesc:
		collator := collate.New(language.SimplifiedChinese, collate.Numeric, collate.IgnoreCase)
		sort.Slice(files, func(i, j int) bool {
			return collator.CompareString(files[i].BaseName, files[j].BaseName) > 0
		})
	default:
		collator := collate.New(language.SimplifiedChinese, collate.Numeric, collate.IgnoreCase)
		sort.Slice(files, func(i, j int) bool {
			return collator.CompareString(files[i].BaseName, files[j].BaseName) < 0
		})
	}
	return files
}
