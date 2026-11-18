package services

import (
	"context"
	"log"
	"musicassist/internal/config"
	"musicassist/internal/tools"
	"musicassist/internal/updater"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"time"

	"github.com/pkg/errors"
	"github.com/wailsapp/wails/v3/pkg/application"
	wailsupdater "github.com/wailsapp/wails/v3/pkg/updater"
	"github.com/wailsapp/wails/v3/pkg/updater/providers/github"
)

type updateService struct {
	app         *application.App
	curVersion  string
	checking    atomic.Bool
	downloading atomic.Bool
}

// Shutdown or restart to apply update
var isUpdateRestarting = false

func init() {
	application.RegisterEvent[wailsupdater.ErrorInfo](wailsupdater.EventError)
	application.RegisterEvent[*wailsupdater.Release](wailsupdater.EventUpdateAvailable)
	application.RegisterEvent[wailsupdater.Progress](wailsupdater.EventDownloadProgress)
}

func NewUpdateService(app *application.App) *updateService {
	cloudflareR2Provider := &updater.CloudflareR2Provider{
		Repo:       config.ProjectName,
		HTTPClient: tools.HttpClient,
		BaseUrl:    config.GenConfig.Env.R2PublicBaseUrl,
	}
	providers := []wailsupdater.Provider{cloudflareR2Provider}
	giteeProvider := &updater.GiteeProvider{
		Repo:       config.ProjectName,
		Owner:      "wtl1989",
		HTTPClient: tools.HttpClient,
		AssetsMatcher: func(req wailsupdater.CheckRequest, assets []updater.GiteeAsset) int {
			for i, a := range assets {
				name := strings.ToLower(a.Name)
				if strings.Contains(name, config.GenConfig.ExeName()+".exe") {
					return i
				}
			}
			return -1
		},
		AccessToken: config.GenConfig.Env.GiteeAccessToken,
	}
	providers = append(providers, giteeProvider)
	githubProvier, err := github.New(github.Config{
		Repository: "wtl568594290/" + config.ProjectName,
		Token:      config.GenConfig.Env.GithubAccessToken,
		Prerelease: false,
		BaseURL:    "",
		AssetMatcher: func(req wailsupdater.CheckRequest, assets []github.ReleaseAsset) int {
			for i, a := range assets {
				name := strings.ToLower(a.Name)
				if strings.Contains(name, config.GenConfig.ExeName()+".exe") {
					return i
				}
			}
			return -1
		},
		HTTPClient: tools.HttpClient,
	})

	if err != nil {
		log.Fatal(err)
	} else {
		providers = append(providers, githubProvier)
	}

	if err := app.Updater.Init(wailsupdater.Config{
		CurrentVersion: config.GenConfig.Version(),
		Providers:      providers,
		Window:         wailsupdater.WindowNone,
	}); err != nil {
		log.Fatal(err)
	}
	return &updateService{app: app}
}

func removeUpdateTmpFiles() {
	// search temp dir "wails-update-*"
	exe, err := os.Executable()
	if err != nil {
		log.Println(err)
		return
	}
	dir := filepath.Dir(exe)
	entries, err := os.ReadDir(dir)
	if err != nil {
		log.Println(err)
		return
	}

	for _, entry := range entries {
		if !entry.IsDir() || !strings.HasPrefix(entry.Name(), "wails-update-") {
			continue
		}
		children, err := os.ReadDir(filepath.Join(dir, entry.Name()))
		if err != nil {
			log.Println(err)
			continue
		}
		shouldDel := len(children) == 0
		for _, child := range children {
			childName := child.Name()
			if !child.IsDir() && (strings.Contains(childName, config.GenConfig.ExeName()+".exe") ||
				strings.Contains(childName, ".artifact")) {
				shouldDel = true
				break
			}
		}
		if shouldDel {
			if err := os.RemoveAll(filepath.Join(dir, entry.Name())); err != nil {
				log.Println(err)
			}
		}
	}
}
func (u *updateService) ServiceStartup(ctx context.Context, options application.ServiceOptions) error {
	removeUpdateTmpFiles()
	return nil
}

func (u *updateService) ServiceShutdown() error {
	if !isUpdateRestarting {
		removeUpdateTmpFiles()
	}
	return nil
}

func (u *updateService) CheckForUpdates() (*wailsupdater.Release, error) {
	if !u.checking.CompareAndSwap(false, true) {
		return nil, errors.New("update check already in progress")
	}
	defer u.checking.Store(false)

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	release, err := u.app.Updater.Check(ctx)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	return release, nil
}

func (u *updateService) DownloadUpdate() error {
	if !u.downloading.CompareAndSwap(false, true) {
		return errors.New("update download already in progress")
	}
	defer u.downloading.Store(false)

	if u.app.Updater.State() != wailsupdater.StateAvailable {
		return errors.New("no available update")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()
	err := u.app.Updater.DownloadAndInstall(ctx)
	if err != nil {
		log.Println(err)
	}
	return err
}

func (u *updateService) RestartApp() error {
	if u.app.Updater.State() == wailsupdater.StateReady {
		isUpdateRestarting = true
		return u.app.Updater.Restart(context.Background())
	}
	return nil
}
