package jwtoggl

import (
	"fmt"
	"go-jw-toggl/jwtoggl/builder"
	"go-jw-toggl/jwtoggl/status"
	"log"
)

func (j *jApp) initStatusBar(b *builder.Builder) {
	j.statusBar = status.New(b)
	j.db.OnOnlineStatusChanged(j.OnOnlineStatusChanged)
}

func (j *jApp) OnOnlineStatusChanged(online bool) {
	msg := "Offline"
	if online {
		msg = "Online"
	}
	if err := j.tray.ShowNotification("JWToggle", fmt.Sprintf("Client is %s", msg)); err != nil {
		log.Fatal(err)
	}
	j.statusBar.SetOnlineStatus(online)
}
