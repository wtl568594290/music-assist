package app

import (
	"context"

	wailsruntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

// App struct
type app struct {
	Ctx context.Context
}

var App = &app{}

// startup is called at application startup
func (a *app) Startup(ctx context.Context) {
	// Perform your setup here
	a.Ctx = ctx
	if Config.PosX >= 0 && Config.PosY >= 0 {
		wailsruntime.WindowSetPosition(a.Ctx, Config.PosX, Config.PosY)
	}
	if Config.Width > 0 && Config.Height > 0 {
		wailsruntime.WindowSetSize(a.Ctx, Config.Width, Config.Height)
	}
	resolveSource(Config.Sources)
}

// domReady is called after front-end resources have been loaded
func (a app) DomReady(ctx context.Context) {
	// Add your action here
	wailsruntime.EventsOff(ctx, "index-change")
	wailsruntime.EventsOn(ctx, "index-change", watchLatestIndex)
}

// beforeClose is called when the application is about to quit,
// either by clicking the window close button or calling runtime.Quit.
// Returning true will cause the application to continue, false will continue shutdown as normal.
func (a *app) BeforeClose(ctx context.Context) (prevent bool) {
	x, y := wailsruntime.WindowGetPosition(a.Ctx)
	w, h := wailsruntime.WindowGetSize(a.Ctx)
	Config.PosX = x
	Config.PosY = y
	Config.Width = w
	Config.Height = h
	s := a.GetSources()
	sources := make([]string, 0)
	for _, v := range s {
		sources = append(sources, v.Path)
	}
	Config.Sources = sources
	Config.Save()
	return false
}

// shutdown is called at application termination
func (a *app) Shutdown(ctx context.Context) {
	// Perform your teardown here
}

func watchLatestIndex(data ...interface{}) {
	if len(data) == 0 {
		return
	}
	index, ok := data[0].(float64)
	if !ok {
		return
	}
	Config.LatestIndex = index
}
