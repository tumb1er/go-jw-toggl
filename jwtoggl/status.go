package jwtoggl

import (
	"go-jw-toggl/jwtoggl/builder"
	"go-jw-toggl/jwtoggl/status"
	"log"
)

func (j *jApp) initStatusBar(b *builder.Builder) {
	j.statusBar = status.New(b)
}

func (j *jApp) handleError(err error, msg string) {
	if err != nil {
		j.statusBar.SetOnlineStatus(false)
		j.statusBar.ShowError(msg)
		if err := j.tray.ShowNotification("JWToggle", msg); err != nil {
			log.Fatal(err)
		}
	} else {
		j.statusBar.ClearError()
		j.statusBar.SetOnlineStatus(true)
	}
}
