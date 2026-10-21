package services

import (
	"context"
	"musicassist/config"
	"musicassist/tools"

	"github.com/wailsapp/wails/v3/pkg/application"
	"github.com/wailsapp/wails/v3/pkg/events"
	"github.com/wailsapp/wails/v3/pkg/services/notifications"
)

// App struct
type appService struct {
	app          *application.App
	notifier     *notifications.NotificationService
	staticConfig *tools.StaticConfig
}

const (
	maxRequestNotificationCnt = 3
)

var requestNotificationCnt int = 0
var globalApp *appService
var mainWin *application.WebviewWindow

func NewAppService(a *application.App, notifier *notifications.NotificationService, staticConfig *tools.StaticConfig) *appService {
	if globalApp != nil {
		return globalApp
	}
	globalApp = &appService{app: a, notifier: notifier, staticConfig: staticConfig}
	return globalApp
}

func sendNotification(id, title, body string) {
	if globalApp == nil || globalApp.notifier == nil || requestNotificationCnt >= maxRequestNotificationCnt {
		return
	}
	if authorized, err := globalApp.notifier.CheckNotificationAuthorization(); err == nil {
		if !authorized {
			// Request authorization
			authorized, err = globalApp.notifier.RequestNotificationAuthorization()
			requestNotificationCnt++
		}
		if authorized {
			// Send notifications
			globalApp.notifier.SendNotification(notifications.NotificationOptions{
				ID:    id,
				Title: title,
				Body:  body,
			})
		}
	}
}

func (as *appService) ServiceStartup(ctx context.Context, options application.ServiceOptions) error {
	w, h := 1000, 618
	cw, ch := config.GetSize()
	if cw > 0 && ch > 0 {
		w = cw
		h = ch
	}
	x, y := config.GetPos()
	if cw > 0 && ch > 0 {
		x = cw
		y = ch
	}
	if mainWin == nil {
		mainWin = as.app.Window.NewWithOptions(application.WebviewWindowOptions{
			Title: as.app.Config().Name,
			// Window sized to the golden ratio (1000 / 618 ≈ 1.618).
			Width:  w,
			Height: h,
			X:      x,
			Y:      y,
			Windows: application.WindowsWindow{
				Theme: application.Dark,
			},
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
			config.SetPos(x, y)
			config.SetSize(w, h)
			s := getCurSources()
			sources := make([]string, 0)
			for _, v := range s {
				sources = append(sources, v.Path)
			}
			config.SetSources(sources)
			config.Save()
		})
	}
	// as.app.Event.Off("index-change")
	// as.app.Event.On("index-change", func(event *application.CustomEvent) {
	// 	watchLatestIndex(event.Data)
	// })
	resolveSource(config.GetSources())
	mainWin.Show()
	mainWin.Focus()

	return nil
}

func (as *appService) ChangeLatestIndex(index float64) {
	config.SetLatestIndex(index)
}

func (as *appService) GetConfigs() config.Conf {
	return config.ViewableConfig()
}

func (as *appService) SetSortType(sort string) {
	st, err := tools.ParseSortType(sort)
	if err != nil {
		st = tools.SortByName
	}
	config.SetSortType(st)
}

func (as *appService) GetSortTypeList() []string {
	return tools.GetSortTypeList()
}

func (as *appService) GetStaticConfig() *tools.StaticConfig {
	return as.staticConfig
}
