package services

import (
	"context"

	"github.com/wailsapp/wails/v3/pkg/application"
	"github.com/wailsapp/wails/v3/pkg/events"
)

// App struct
type appService struct {
	app *application.App
}

var mainWin *application.WebviewWindow

func init() {
	application.RegisterEvent[string]("sync-msg")
	application.RegisterEvent[SyncProgressBody]("sync-msg-progress")
	application.RegisterEvent[string]("sync-msg-error")
}

func NewAppService(a *application.App) *appService {
	return &appService{app: a}
}

func (as *appService) ServiceStartup(ctx context.Context, options application.ServiceOptions) error {
	w, h := 1000, 618
	if Config.Width > 0 && Config.Height > 0 {
		w = Config.Width
		h = Config.Height
	}
	x, y := 0, 0
	if Config.PosX >= 0 && Config.PosY >= 0 {
		x = Config.PosX
		y = Config.PosY
	}
	if mainWin == nil {
		mainWin = as.app.Window.NewWithOptions(application.WebviewWindowOptions{
			Title: "music-assist",
			// Window sized to the golden ratio (1000 / 618 ≈ 1.618).
			Width:  w,
			Height: h,
			X:      x,
			Y:      y,
			Mac: application.MacWindow{
				InvisibleTitleBarHeight: 50,
				Backdrop:                application.MacBackdropTranslucent,
				TitleBar:                application.MacTitleBarHiddenInset,
			},
			URL: "/",
		})
		mainWin.RegisterHook(events.Common.WindowClosing, func(event *application.WindowEvent) {
			bounds := mainWin.Bounds()
			x, y := bounds.X, bounds.Y
			w, h := bounds.Width, bounds.Height
			Config.PosX = x
			Config.PosY = y
			Config.Width = w
			Config.Height = h
			s := as.GetSources()
			sources := make([]string, 0)
			for _, v := range s {
				sources = append(sources, v.Path)
			}
			Config.Sources = sources
			Config.Save()
		})
	}
	// as.app.Event.Off("index-change")
	// as.app.Event.On("index-change", func(event *application.CustomEvent) {
	// 	watchLatestIndex(event.Data)
	// })
	resolveSource(Config.Sources)
	mainWin.Show()
	mainWin.Focus()

	return nil
}

func (as *appService) ChangeLatestIndex(index float64) {
	Config.LatestIndex = index
}

// func watchLatestIndex(data any) {
// 	index, ok := data.(float64)
// 	if !ok {
// 		return
// 	}
// 	Config.LatestIndex = index
// }
