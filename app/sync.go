package app

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

func SyncRemote(url string, docSrcs []DocSource) {
	if url == "" {
		return
	}
	if k, err := json.Marshal(docSrcs); err == nil {
		Config.RemoteUrl = url
		go func() {
			resp, err := http.Post(url+"/doc", "application/json", bytes.NewReader(k))
			if err != nil {
				log.Println(err)
				runtime.EventsEmit(App.Ctx, "sync-msg-error", "Sync list error:"+err.Error())
				return
			}
			runtime.EventsEmit(App.Ctx, "sync-msg", "Sync list success")
			cnt := 0
			uploadImg := resp.StatusCode == http.StatusOK
			resp.Body.Close()
			if uploadImg {
				for _, v := range docSrcs {
					for _, img := range v.Imgs {
						if err := updateImage(url, img.BaseName, img.Path); err != nil {
							log.Println(err)
						} else {
							cnt++
						}
					}
				}
			}
			if cnt > 0 {
				runtime.EventsEmit(App.Ctx, "sync-msg", fmt.Sprintf("Sync %d images success", cnt))
			} else {
				runtime.EventsEmit(App.Ctx, "sync-msg-error", "Sync no image success")
			}
		}()
	}
}

func updateImage(url string, name string, filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)

	if err = writer.WriteField("name", name); err != nil {
		return err
	}

	part, err := writer.CreateFormFile("file", filepath.Base(filePath))
	if err != nil {
		return err
	}

	if _, err = io.Copy(part, file); err != nil {
		return err
	}

	if err = writer.Close(); err != nil {
		return err
	}

	req, err := http.NewRequest(
		http.MethodPost,
		url+"/img",
		&body,
	)
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("upload %s failed: %s", name, resp.Status)
	}

	return nil
}

func (d *DocSource) MarshalJSON() ([]byte, error) {
	type imgBrief struct {
		Name   string `json:"name"`
		Number int    `json:"number"`
	}
	type fileInfoShortT struct {
		Name         string     `json:"name"`
		ImgBaseNames []imgBrief `json:"imgs"`
	}
	fis := fileInfoShortT{Name: d.Name}
	for i, v := range d.Imgs {
		fis.ImgBaseNames = append(fis.ImgBaseNames, imgBrief{
			Name:   v.BaseName,
			Number: i + 1,
		})
	}
	return json.Marshal(fis)
}
