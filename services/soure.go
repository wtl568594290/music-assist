package services

import (
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"

	"github.com/wailsapp/wails/v3/pkg/application"
)

var (
	srcs       = NewMapStringBool()
	imgMap     = NewMapStringStruct()
	latestHash = ""
)

func isImage(filePath string) (bool, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return false, err
	}
	defer file.Close()

	// 读取前 512 字节用于检测
	buffer := make([]byte, 512)
	_, err = file.Read(buffer)
	if err != nil {
		return false, err
	}

	// 检测 MIME 类型
	contentType := http.DetectContentType(buffer)
	return contentType[:6] == "image/", nil
}

func (as *appService) addImagePath(isTraversal bool) error {
	title := "Select a image folder"
	if isTraversal {
		title += " (traverse)"
	}
	options := &application.OpenFileDialogOptions{
		Title:                title,
		CanChooseDirectories: true,
	}
	path, err := as.app.Dialog.OpenFileWithOptions(options).PromptForSingleSelection()
	if err != nil {
		return err
	}
	if path == "" {
		return errors.New("path is empty")
	}

	if isTraversal {
		// traverse the folder
		err = filepath.WalkDir(path, func(p string, d fs.DirEntry, e error) error {
			if e != nil {
				return e
			}
			if d.IsDir() {
				if _, ok := srcs[p]; !ok {
					srcs[p] = true
				}
			}
			return nil
		})
		if err != nil {
			return errors.New("traverse folder failed")
		}
	} else {
		if _, ok := srcs[path]; ok {
			return errors.New("path already added")
		}
		srcs[path] = true
	}
	return err
}

func (as *appService) addImageFiles() error {
	options := &application.OpenFileDialogOptions{
		Title: "Select image files",
		Filters: []application.FileFilter{
			{
				DisplayName: "jpg, jpeg",
				Pattern:     "*.jpg;*.jpeg;",
			},
			{
				DisplayName: "png",
				Pattern:     "*.png;",
			},
			{
				DisplayName: "gif",
				Pattern:     "*.gif;",
			},
			{
				DisplayName: "bmp",
				Pattern:     "*.bmp;",
			},
			{
				DisplayName: "webp",
				Pattern:     "*.webp;",
			},
			{
				DisplayName: "All files",
				Pattern:     "*.*",
			},
		},
	}
	files, err := as.app.Dialog.OpenFileWithOptions(options).PromptForMultipleSelection()
	if err != nil {
		return err
	}
	if len(files) == 0 {
		return errors.New("files is empty")
	}
	for _, path := range files {
		isImg, err := isImage(path)
		if err != nil {
			continue
		}
		if isImg {
			if _, ok := srcs[path]; ok {
				continue
			}
			srcs[path] = false
		}
	}
	return err
}

func resolveSource(paths []string) {
	clear(srcs)
	for _, path := range paths {
		info, err := os.Stat(path)
		if err != nil {
			continue
		}
		if info.IsDir() {
			srcs[path] = true
		} else {
			srcs[path] = false
		}
	}
}

func generateImagesMap() (MapStringFileInfo, error) {
	var imgs = NewMapStringStruct()
	for path, isDir := range srcs {
		if isDir {
			files, err := os.ReadDir(path)
			if err != nil {
				continue
			}
			for _, file := range files {
				if file.IsDir() {
					continue
				}
				filePath := filepath.Join(path, file.Name())
				isImg, err := isImage(filePath)
				if err != nil {
					continue
				}
				if isImg {
					imgs.Add(filePath)
				}
			}
		} else {
			isImg, err := isImage(path)
			if err != nil {
				continue
			}
			if isImg {
				imgs.Add(path)
			}
		}
	}
	return imgs, nil
}

func (as *appService) AddSource(isDir bool, isTraversal bool) error {
	if isDir {
		return as.addImagePath(isTraversal)
	}
	return as.addImageFiles()
}

func (as *appService) ClearSources() error {
	clear(srcs)
	return nil
}

func (as *appService) RemoveSource(path string) error {
	delete(srcs, path)
	return nil
}
func (as *appService) GetSources() []Source {
	return srcs.ToSlice()
}

func (as *appService) GetImages() ([]FileInfo, error) {
	clear(imgMap)
	mp, err := generateImagesMap()
	if err != nil {
		return mp.ToSlice(Config.SortType), err
	}
	// 缓存图片
	imgMap = mp
	imgs := imgMap.ToSlice(Config.SortType)
	latestHash = calcHash(imgs)
	return imgs, nil
}

func (as *appService) GetBase64(path string) (string, error) {
	fileBytes, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	base64Str := base64.StdEncoding.EncodeToString(fileBytes)
	return "data:image/png;base64," + base64Str, nil
}

func calcHash(files []FileInfo) string {
	h := sha256.New()
	for _, img := range files {
		h.Write([]byte(img.BaseName + img.Hash))
	}
	hash := h.Sum(nil)
	return fmt.Sprintf("%x", hash[:])
}

func (as *appService) DetectChanges() bool {
	m, _ := generateImagesMap()
	imgs := m.ToSlice(Config.SortType)
	change := latestHash != "" && calcHash(imgs) != latestHash
	return change
}
