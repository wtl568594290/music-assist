package services

import (
	"context"
	"log"
	"musicassist/tools"
	"musicassist/updater"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/wailsapp/wails/v3/pkg/application"
	wailsupdater "github.com/wailsapp/wails/v3/pkg/updater"
	"github.com/wailsapp/wails/v3/pkg/updater/providers/github"
)

type updateService struct {
	app        *application.App
	curVersion string
}

// Shutdown or restart to apply update
var isUpdateRestarting = false

func init() {
	application.RegisterEvent[wailsupdater.ErrorInfo](wailsupdater.EventError)
	application.RegisterEvent[*wailsupdater.Release](wailsupdater.EventUpdateAvailable)
	application.RegisterEvent[wailsupdater.Progress](wailsupdater.EventDownloadProgress)
}

func NewUpdateService(app *application.App) *updateService {
	giteeProvider := &updater.GiteeProvider{
		Repo:       tools.ProjectName,
		Owner:      "wtl1989",
		HTTPClient: tools.HttpClient,
		AssetsMatcher: func(req wailsupdater.CheckRequest, assets []updater.GiteeAsset) int {
			for i, a := range assets {
				name := strings.ToLower(a.Name)
				if strings.Contains(name, globalApp.staticConfig.ExeName()+".exe") {
					return i
				}
			}
			return -1
		},
		AccessToken: tools.Env.GiteeAccessToken,
	}
	providers := []wailsupdater.Provider{giteeProvider}
	githubProvier, err := github.New(github.Config{
		Repository: "wtl568594290/" + tools.ProjectName,
		Token:      tools.Env.GithubAccessToken,
		Prerelease: false,
		BaseURL:    "",
		AssetMatcher: func(req wailsupdater.CheckRequest, assets []github.ReleaseAsset) int {
			for i, a := range assets {
				name := strings.ToLower(a.Name)
				if strings.Contains(name, globalApp.staticConfig.ExeName()+".exe") {
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
		CurrentVersion: globalApp.staticConfig.Version(),
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
			if !child.IsDir() && (strings.Contains(childName, globalApp.staticConfig.ExeName()+".exe") ||
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

func (u *updateService) CheckForUpdates() {
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cancel()
		if _, err := u.app.Updater.Check(ctx); err != nil {
			log.Println(err)
		}
	}()
}

func (u *updateService) DownloadUpdate() error {
	if u.app.Updater.State() != wailsupdater.StateAvailable {
		return errors.New("no available update")
	}
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
		defer cancel()
		if err := u.app.Updater.DownloadAndInstall(ctx); err != nil {
			log.Println(err)
		}
	}()
	return nil
}

func (u *updateService) RestartApp() error {
	if u.app.Updater.State() == wailsupdater.StateReady {
		isUpdateRestarting = true
		return u.app.Updater.Restart(context.Background())
	}
	return nil
}
