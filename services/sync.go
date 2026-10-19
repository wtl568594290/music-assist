package services

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"sync/atomic"
	"time"
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

type UploadTask struct {
	Url string
	Img FileInfo
}

type DownloadTask struct {
	Url      string
	Dir      string
	FileName string
}

type SyncMsgType string

const (
	SyncMsg         SyncMsgType = "sync-msg"
	SyncMsgProgress SyncMsgType = "sync-msg-progress"
	SyncMsgError    SyncMsgType = "sync-msg-error"
)

var syncClient = &http.Client{
	Transport: &http.Transport{
		MaxIdleConns:        20,
		MaxIdleConnsPerHost: 8,
		MaxConnsPerHost:     8,
		IdleConnTimeout:     30 * time.Second,
	},
}

func newSyncProgress(cnt int, total int) SyncProgressBody {
	return SyncProgressBody{Total: total, Cnt: cnt}
}

func emitSyncMsg(msgType SyncMsgType, msg any) {
	mainWin.EmitEvent(string(msgType), msg)
}

func SyncRemote(url string, docSrcs []DocSource) {
	if url == "" {
		return
	}
	if k, err := json.Marshal(docSrcs); err == nil {
		Config.RemoteUrl = url
		go func() {
			resp, err := syncClient.Post(url+"/doc", "application/json", bytes.NewReader(k))
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
			var cnt atomic.Int32
			tasks := make([]UploadTask, 0, total)
			emitSyncMsg(SyncMsgProgress, newSyncProgress(0, total))
			for _, s := range syncRes.Data {
				for _, v := range docSrcs {
					for _, img := range v.Imgs {
						if s.Hash == img.Hash && s.Name == img.BaseName {
							tasks = append(tasks, UploadTask{Url: url, Img: img})
						}
					}
				}
			}
			RunConcurrent(context.Background(), tasks, Options{
				MaxConcurrent: 8,
				OnComplete: func(index int, err error) {
					if err != nil {
						log.Println(err)
					} else {
						cnt.Add(1)
						emitSyncMsg(SyncMsgProgress, newSyncProgress(int(cnt.Load()), total))
					}
				},
			}, func(task UploadTask) error {
				return updateImage(task.Url, task.Img)
			})
			emitSyncMsg(SyncMsgProgress, newSyncProgress(0, 0))
			emitSyncMsg(SyncMsg, fmt.Sprintf("Upload complete, %d/%d images success", cnt.Load(), total))
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
	resp, err := syncClient.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download failed, status: %s", resp.Status)
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

func (as *appService) DownloadRemote(urlStr string) error {
	if urlStr == "" {
		return errors.New("Remote url is empty")
	}
	Config.RemoteUrl = urlStr
	go func() {
		resp, err := syncClient.Get(urlStr + "/doc")
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
			emitSyncMsg(SyncMsgError, "No image to download")
			return
		}
		emitSyncMsg(SyncMsg, fmt.Sprintf("Download list success, there are %d images will be downloaded", total))
		tasks := make([]DownloadTask, 0, total)
		emitSyncMsg(SyncMsgProgress, newSyncProgress(0, total))
		for _, f := range fileRes.Data {

			// 创建目录
			if err = os.MkdirAll(filepath.Join(DownloadPath, f.Name), 0755); err != nil {
				emitSyncMsg(SyncMsgError, "Create dir error:"+err.Error())
				return
			}

			for _, img := range f.ImgBaseNames {
				q := url.Values{}
				q.Set("name", img.Name)
				q.Set("hash", img.Hash)
				tasks = append(tasks, DownloadTask{
					Url:      urlStr + "/img?" + q.Encode(),
					Dir:      f.Name,
					FileName: img.Name + img.Ext,
				})
			}
		}
		var cnt atomic.Int32
		RunConcurrent(context.Background(), tasks, Options{
			MaxConcurrent: 8,
			OnComplete: func(index int, err error) {
				if err != nil {
					log.Println(err)
				} else {
					cnt.Add(1)
					emitSyncMsg(SyncMsgProgress, newSyncProgress(int(cnt.Load()), total))
				}
			},
		}, func(t DownloadTask) error {
			return downloadFile(t.Url, t.Dir, t.FileName)
		})
		emitSyncMsg(SyncMsgProgress, newSyncProgress(0, 0))
		emitSyncMsg(SyncMsg, fmt.Sprintf("Download complete, %d/%d images success", int(cnt.Load()), total))
	}()
	return nil
}

func (as *appService) PlaceholderSyncProgressBody() SyncProgressBody {
	return SyncProgressBody{}
}
