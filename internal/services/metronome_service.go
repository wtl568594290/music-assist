package services

import (
	"context"
	"log"
	"musicassist/internal/config"
	"sync"

	"github.com/wailsapp/wails/v3/pkg/application"
	"github.com/wailsapp/wails/v3/pkg/events"
)

type metronomeService struct {
	app *application.App
}

const (
	SplitMsg = "split-msg"
)

var (
	metronomeWin   *application.WebviewWindow
	metronomeWinMu sync.Mutex
)

func NewMetronomeService(app *application.App) *metronomeService {
	return &metronomeService{app: app}
}

func init() {
	application.RegisterEvent[bool](SplitMsg)
}

func (ms *metronomeService) ServiceStartup(ctx context.Context, options application.ServiceOptions) error {
	return nil
}

// takeMetronomeWindow 取出并清空独立节拍器窗口引用，返回取到的窗口（可能为 nil）
func takeMetronomeWindow() *application.WebviewWindow {
	metronomeWinMu.Lock()
	defer metronomeWinMu.Unlock()
	win := metronomeWin
	metronomeWin = nil
	return win
}

func (ms *metronomeService) CreateMetronomeWindow() {
	metronomeWinMu.Lock()
	defer metronomeWinMu.Unlock()

	if metronomeWin == nil {
		w, h := config.GetMetronomeSize()
		x, y := config.GetMetronomePos()
		// 未保存过位置时 PosX/PosY 为 -1，此时交由系统居中
		initialPosition := application.WindowCentered
		if x >= 0 && y >= 0 {
			initialPosition = application.WindowXY
		}
		win := ms.app.Window.NewWithOptions(application.WebviewWindowOptions{
			Title:           "Metronome",
			Width:           w,
			Height:          h,
			InitialPosition: initialPosition,
			X:               x,
			Y:               y,
			URL:             "/#/metronome",
			Windows: application.WindowsWindow{
				Theme: application.Dark,
			},
		})
		// 使用 Hook 保存位置与大小：Hook 会在窗口关闭监听器（含默认销毁逻辑）之前同步执行
		win.RegisterHook(events.Common.WindowClosing, func(event *application.WindowEvent) {
			// 最小化时窗口坐标不可靠，跳过保存，保留上一次的位置与大小
			if win.IsMinimised() {
				return
			}
			bounds := win.Bounds()
			config.SetMetronomePos(bounds.X, bounds.Y)
			config.SetMetronomeSize(bounds.Width, bounds.Height)
			if err := config.Save(); err != nil {
				log.Println(err)
			}
		})
		win.OnWindowEvent(events.Common.WindowClosing, func(event *application.WindowEvent) {
			metronomeWinMu.Lock()
			cleared := metronomeWin == win
			if cleared {
				metronomeWin = nil
			}
			metronomeWinMu.Unlock()
			// 由用户直接关闭窗口时需要通知主窗口恢复标签页
			if cleared {
				ms.app.Event.Emit(SplitMsg, false)
			}
		})
		metronomeWin = win
	}

	metronomeWin.Show()
	metronomeWin.Focus()
	ms.app.Event.Emit(SplitMsg, true)
}

func (ms *metronomeService) CloseMetronomeWindow() {
	win := takeMetronomeWindow()
	if win == nil {
		return
	}
	// 先清空引用再关闭，避免 WindowClosing 回调重复处理
	win.Close()
	ms.app.Event.Emit(SplitMsg, false)
}

func (ms *metronomeService) IsMetronomeIndependentWindow() bool {
	metronomeWinMu.Lock()
	defer metronomeWinMu.Unlock()
	return metronomeWin != nil
}
