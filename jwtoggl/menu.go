package jwtoggl

import (
	"go-jw-toggl/jwtoggl/builder"
	"go-jw-toggl/jwtoggl/tray"
	"log"
)

func (j *jApp) initTray(b *builder.Builder) {
	var err error
	if j.tray, err = tray.New(b); err != nil {
		log.Fatal("tray init error:", err)
	}
	if err = j.tray.ConnectSignals(j.onShow, j.onExit); err != nil {
		log.Fatal("tray signals connect error:", err)
	}
}

func (j *jApp) onShow() {
	log.Println("onShow")
	j.Call(func() {
		visible := j.window.GetVisible()
		j.window.SetVisible(!visible)
	})
}

func (j *jApp) onExit() {
	log.Println("onExit")
	j.cancel()
	j.Call(j.app.Quit)
}
