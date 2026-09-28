package app

import (
	"crypto/md5"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"syscall"
	"time"

	"golang.org/x/text/collate"
	"golang.org/x/text/language"
)

type FileInfo struct {
	Path     string
	Dir      string
	BaseName string
	Ext      string
	ModTime  int64
	Created  int64
	Size     int64
	Hash     string
}

type SortType string

const (
	SortByName        SortType = "Name"
	SortByNameDesc    SortType = "NameDesc"
	SortByModTime     SortType = "ModTime"
	SortByModTimeDesc SortType = "ModTimeDesc"
	SortByCreated     SortType = "Created"
	SortByCreatedDesc SortType = "CreatedDesc"
)

var validSortTypes = map[string]SortType{
	"Name":        SortByName,
	"NameDesc":    SortByNameDesc,
	"ModTime":     SortByModTime,
	"ModTimeDesc": SortByModTimeDesc,
	"Created":     SortByCreated,
	"CreatedDesc": SortByCreatedDesc,
}

type Source struct {
	Path  string
	IsDir bool
}

type MapStringFileInfo map[string]FileInfo
type MapStringBool map[string]bool

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
	size := entry.Size()
	hashByte := md5.Sum((fmt.Appendf(nil, "%s|%d|%d", path, modTime, size)))
	m[path] = FileInfo{
		Path:     path,
		Dir:      dir,
		BaseName: baseName,
		Ext:      ext,
		ModTime:  modTime,
		Created:  getCreateTime(entry).UnixMilli(),
		Size:     size,
		Hash:     fmt.Sprintf("%x", hashByte[:]),
	}
	return nil
}

func NewMapStringStruct() MapStringFileInfo {
	return make(MapStringFileInfo)
}

func (m MapStringBool) ToSlice() []Source {
	slice := make([]Source, 0, len(m))
	for path, isDir := range m {
		slice = append(slice, Source{
			Path:  path,
			IsDir: isDir,
		})
	}
	return slice
}

func (m MapStringBool) Add(path string, isDir bool) bool {
	_, ok := m[path]
	if ok {
		return false
	}
	m[path] = isDir
	return true
}

func NewMapStringBool() MapStringBool {
	return make(MapStringBool)
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

func ParseSortType(s string) (SortType, error) {
	st, ok := validSortTypes[s]
	if !ok {
		return "", fmt.Errorf("invalid SortType: %s", s)
	}
	return st, nil
}

func (a *app) SetSortType(sort string) {
	st, err := ParseSortType(sort)
	if err != nil {
		st = SortByName
	}
	Config.SortType = st
}

func (a *app) GetSortTypeList() []string {
	keys := make([]string, 0, len(validSortTypes))
	for k := range validSortTypes {
		keys = append(keys, k)
	}
	return keys
}
