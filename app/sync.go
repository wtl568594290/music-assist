package app

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"time"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type syncRes[T any] struct {
	Message string `json:"message"`
	Data    T      `json:"data"`
}

type imgBrief struct {
	Name   string `json:"name"`
	Number int    `json:"number"`
	Hash   string `json:"hash"`
	Ext    string `json:"ext"`
}
type fileInfoShortT struct {
	Name         string     `json:"name"`
	ImgBaseNames []imgBrief `json:"imgs"`
}

type SyncProgressBody struct {
	Total int
	Cnt   int
}

type SyncMsgType string

const (
	SyncMsg         SyncMsgType = "sync-msg"
	SyncMsgProgress SyncMsgType = "sync-msg-progress"
	SyncMsgError    SyncMsgType = "sync-msg-error"
)

func newSyncProgress(cnt int, total int) SyncProgressBody {
	return SyncProgressBody{Total: total, Cnt: cnt}
}

func emitSyncMsg(msgType SyncMsgType, msg any) {
	runtime.EventsEmit(App.Ctx, string(msgType), msg)
}
func waitMoment() {
	time.Sleep(50 * time.Millisecond)
}

func SyncRemote(url string, docSrcs []DocSource) {
	if url == "" {
		return
	}
	if k, err := json.Marshal(docSrcs); err == nil {
		Config.RemoteUrl = url
		go func() {
			resp, err := http.Post(url+"/doc", "application/json", bytes.NewReader(k))
			if err != nil || resp.StatusCode != http.StatusOK {
				emitSyncMsg(SyncMsgError, "Sync list error")
				return
			}
			var syncRes syncRes[[]imgBrief]
			err = json.NewDecoder(resp.Body).Decode(&syncRes)
			resp.Body.Close()
			if err != nil {
				emitSyncMsg(SyncMsgError, "Decode sync res error:"+err.Error())
				return
			}
			total := len(syncRes.Data)
			if total == 0 {
				emitSyncMsg(SyncMsg, "No image to upload")
				return
			}
			emitSyncMsg(SyncMsg, fmt.Sprintf("Sync list success, there are %d images will be uploaded", total))
			waitMoment()
			cnt := 0
			for _, s := range syncRes.Data {
				for _, v := range docSrcs {
					for _, img := range v.Imgs {
						if s.Hash == img.Hash && s.Name == img.BaseName {
							waitMoment()
							if err := updateImage(url, img); err == nil {
								cnt++
								emitSyncMsg(SyncMsgProgress, newSyncProgress(cnt, total))
							} else {
								log.Println(err)
							}
						}
					}
				}
			}
			waitMoment()
			emitSyncMsg(SyncMsg, fmt.Sprintf("Upload complete, %d/%d images success", cnt, total))
		}()
	}
}

func updateImage(url string, img FileInfo) error {
	file, err := os.Open(img.Path)
	if err != nil {
		return err
	}
	defer file.Close()

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)

	if err = writer.WriteField("name", img.BaseName); err != nil {
		return err
	}

	if err = writer.WriteField("hash", img.Hash); err != nil {
		return err
	}
	if err = writer.WriteField("ext", img.Ext); err != nil {
		return err
	}

	part, err := writer.CreateFormFile("file", filepath.Base(img.Path))
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
		return fmt.Errorf("upload %s failed: %s", img.BaseName, resp.Status)
	}

	return nil
}

func (d *DocSource) MarshalJSON() ([]byte, error) {

	fis := fileInfoShortT{Name: d.Name}
	for i, v := range d.Imgs {
		fis.ImgBaseNames = append(fis.ImgBaseNames, imgBrief{
			Name:   v.BaseName,
			Number: i + 1,
			Hash:   v.Hash,
			Ext:    v.Ext,
		})
	}
	return json.Marshal(fis)
}

func downloadFile(url, dir, filename string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download failed, status: %s", resp.Status)
	}
	// 创建目录
	if err = os.MkdirAll(filepath.Join(DownloadPath, dir), 0755); err != nil {
		return err
	}

	file, err := os.Create(filepath.Join(DownloadPath, dir, filename))
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = io.Copy(file, resp.Body)
	if err != nil {
		return err
	}

	return nil
}

func (a *app) DownloadRemote(urlStr string) {
	go func() {
		resp, err := http.Get(urlStr + "/doc")
		if err != nil || resp.StatusCode != http.StatusOK {
			emitSyncMsg(SyncMsgError, "Download list error")
			return
		}
		var fileRes syncRes[[]fileInfoShortT]
		err = json.NewDecoder(resp.Body).Decode(&fileRes)
		resp.Body.Close()
		if err != nil {
			emitSyncMsg(SyncMsgError, "Decode file res error:"+err.Error())
			return
		}
		total := 0
		for _, f := range fileRes.Data {
			total += len(f.ImgBaseNames)
		}
		if total == 0 {
			emitSyncMsg(SyncMsg, "No image to download")
			return
		}
		emitSyncMsg(SyncMsg, fmt.Sprintf("Download list success, there are %d images will be downloaded", total))
		waitMoment()
		cnt := 0
		for _, f := range fileRes.Data {
			for _, img := range f.ImgBaseNames {
				q := url.Values{}
				q.Set("name", img.Name)
				q.Set("hash", img.Hash)
				waitMoment()
				if err := downloadFile(urlStr+"/img?"+q.Encode(), f.Name, img.Name+img.Ext); err != nil {
					log.Println(err)
				} else {
					cnt++
					emitSyncMsg(SyncMsgProgress, newSyncProgress(cnt, total))
				}
			}
		}
		waitMoment()
		emitSyncMsg(SyncMsg, fmt.Sprintf("Download complete, %d/%d images success", cnt, total))
	}()
}

func (a *app) PlaceholderSyncProgressBody() SyncProgressBody {
	return SyncProgressBody{}
}
