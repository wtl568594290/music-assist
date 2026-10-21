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
	"musicassist/config"
	"musicassist/tools"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"sync/atomic"

	"github.com/wailsapp/wails/v3/pkg/application"
)

type syncService struct {
}

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

type uploadTask struct {
	Url string
	Img tools.FileInfo
}

type downloadTask struct {
	Url      string
	Dir      string
	FileName string
}

type SyncMsgType string

const (
	SyncMsg         SyncMsgType = "sync-msg"
	SyncMsgProgress SyncMsgType = "sync-msg-progress"
	SyncMsgError    SyncMsgType = "sync-msg-error"

	notificationSyncCompleteId            = "sync-complete"
	notificationSyncDownloadCompleteTitle = "Sync download complete"
	notificationSyncUploadCompleteTitle   = "Sync upload complete"
)

func init() {
	application.RegisterEvent[string](string(SyncMsg))
	application.RegisterEvent[SyncProgressBody](string(SyncMsgProgress))
	application.RegisterEvent[string](string(SyncMsgError))
}

func NewSyncService() *syncService {
	return &syncService{}
}

func (s *syncService) SyncMsgTypes() []SyncMsgType {
	return []SyncMsgType{SyncMsg, SyncMsgProgress, SyncMsgError}
}

func newSyncProgress(cnt int, total int) SyncProgressBody {
	return SyncProgressBody{Total: total, Cnt: cnt}
}

func emitSyncMsg(msgType SyncMsgType, msg any) {
	mainWin.EmitEvent(string(msgType), msg)
}

func syncRemote(url string, docSrcs []docSource) {
	if url == "" {
		return
	}
	if k, err := json.Marshal(docSrcs); err == nil {
		config.SetRemoteUrl(url)
		go func() {
			resp, err := tools.HttpClient.Post(url+"/doc", "application/json", bytes.NewReader(k))
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
			tasks := make([]uploadTask, 0, total)
			emitSyncMsg(SyncMsgProgress, newSyncProgress(0, total))
			for _, s := range syncRes.Data {
				for _, v := range docSrcs {
					for _, img := range v.Imgs {
						if s.Hash == img.Hash && s.Name == img.BaseName {
							tasks = append(tasks, uploadTask{Url: url, Img: img})
						}
					}
				}
			}
			tools.RunConcurrent(context.Background(), tasks, tools.ConcurrentOptions{
				MaxConcurrent: 8,
				OnComplete: func(index int, err error) {
					if err != nil {
						log.Println(err)
					} else {
						cnt.Add(1)
						emitSyncMsg(SyncMsgProgress, newSyncProgress(int(cnt.Load()), total))
					}
				},
			}, func(task uploadTask) error {
				return updateImage(task.Url, task.Img)
			})
			emitSyncMsg(SyncMsgProgress, newSyncProgress(0, 0))
			completeStr := fmt.Sprintf("Upload complete, %d/%d images success", cnt.Load(), total)
			if mainWin != nil && !mainWin.IsFocused() {
				sendNotification(notificationSyncCompleteId, notificationSyncUploadCompleteTitle, completeStr)
			}
			emitSyncMsg(SyncMsg, completeStr)
		}()
	}
}

func updateImage(url string, img tools.FileInfo) error {
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

func (d *docSource) MarshalJSON() ([]byte, error) {

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
	resp, err := tools.HttpClient.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download failed, status: %s", resp.Status)
	}
	file, err := os.Create(filepath.Join(downloadPath, dir, filename))
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

func (syncs *syncService) DownloadRemote(urlStr string) error {
	if urlStr == "" {
		return errors.New("Remote url is empty")
	}
	config.SetRemoteUrl(urlStr)
	go func() {
		resp, err := tools.HttpClient.Get(urlStr + "/doc")
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
		tasks := make([]downloadTask, 0, total)
		emitSyncMsg(SyncMsgProgress, newSyncProgress(0, total))
		for _, f := range fileRes.Data {

			// 创建目录
			if err = os.MkdirAll(filepath.Join(downloadPath, f.Name), 0755); err != nil {
				emitSyncMsg(SyncMsgError, "Create dir error:"+err.Error())
				return
			}

			for _, img := range f.ImgBaseNames {
				q := url.Values{}
				q.Set("name", img.Name)
				q.Set("hash", img.Hash)
				tasks = append(tasks, downloadTask{
					Url:      urlStr + "/img?" + q.Encode(),
					Dir:      f.Name,
					FileName: img.Name + img.Ext,
				})
			}
		}
		var cnt atomic.Int32
		tools.RunConcurrent(context.Background(), tasks, tools.ConcurrentOptions{
			MaxConcurrent: 8,
			OnComplete: func(index int, err error) {
				if err != nil {
					log.Println(err)
				} else {
					cnt.Add(1)
					emitSyncMsg(SyncMsgProgress, newSyncProgress(int(cnt.Load()), total))
				}
			},
		}, func(t downloadTask) error {
			return downloadFile(t.Url, t.Dir, t.FileName)
		})
		emitSyncMsg(SyncMsgProgress, newSyncProgress(0, 0))

		completeStr := fmt.Sprintf("Download complete, %d/%d images success", int(cnt.Load()), total)
		if mainWin != nil && !mainWin.IsFocused() {
			sendNotification(notificationSyncCompleteId, notificationSyncDownloadCompleteTitle, completeStr)
		}
		emitSyncMsg(SyncMsg, completeStr)
	}()
	return nil
}
