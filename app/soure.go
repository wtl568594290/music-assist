package app

import (
	"encoding/base64"
	"errors"
	"net/http"
	"os"
	"path/filepath"

	wailsruntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

var (
	srcs   = NewMapStringBool()
	imgMap = NewMapStringStruct()
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

func (a *app) addImagePath() error {
	options := wailsruntime.OpenDialogOptions{
		Title: "Select a image folder",
	}
	path, err := wailsruntime.OpenDirectoryDialog(a.Ctx, options)
	if err != nil {
		return err
	}
	if path == "" {
		return errors.New("path is empty")
	}
	if _, ok := srcs[path]; ok {
		return errors.New("path already added")
	}
	srcs[path] = true
	return err
}

func (a *app) addImageFiles() error {
	options := wailsruntime.OpenDialogOptions{
		Title: "Select image files",
		Filters: []wailsruntime.FileFilter{
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
	files, err := wailsruntime.OpenMultipleFilesDialog(a.Ctx, options)
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

func (a *app) AddSource(isDir bool) error {
	if isDir {
		return a.addImagePath()
	}
	return a.addImageFiles()
}

func (a *app) ClearSources() error {
	clear(srcs)
	return nil
}

func (a *app) RemoveSource(path string) error {
	delete(srcs, path)
	return nil
}
func (a *app) GetSources() []Source {
	return srcs.ToSlice()
}

func (a *app) GetImages() ([]string, error) {
	clear(imgMap)
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
					imgMap[filePath] = struct{}{}
				}
			}
		} else {
			isImg, err := isImage(path)
			if err != nil {
				continue
			}
			if isImg {
				imgMap[path] = struct{}{}
			}
		}
	}
	return imgMap.ToSlice(), nil
}

func (a *app) GetBase64(path string) (string, error) {
	fileBytes, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	base64Str := base64.StdEncoding.EncodeToString(fileBytes)
	return "data:image/png;base64," + base64Str, nil
}

func (a *app) GetLatestIndex() float64 {
	return Config.LatestIndex
}
