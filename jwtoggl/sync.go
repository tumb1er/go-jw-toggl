package jwtoggl

import (
	"context"
	"github.com/dawidd6/go-appindicator"
	"log"
	"time"
)

func (j *jApp) periodicSync(ctx context.Context) {
	log.Println("Start periodic sync")
	t := time.NewTicker(time.Second * 10)
	j.syncCurrentEntry()
	for {
		select {
		case <-t.C:
			j.syncCurrentEntry()
		case <-ctx.Done():
			log.Println("Stop periodic sync")
			return
		}
	}
}

func (j *jApp) syncCurrentEntry() {
	log.Println("start sync")
	defer log.Println("sync done")
	prev := j.db.GetCurrentEntry()
	if err := j.db.SyncCurrentEntry(); err != nil {
		log.Println("Sync current entry error:", err)
	} else {
		e := j.db.GetCurrentEntry()
		if e == nil {
			if prev != nil {
				j.showNotification("JW Toggl", "Timer stopped")
				j.indicator.SetLabel("", "")
				j.indicator.SetStatus(appindicator.StatusActive)
				j.indicator.SetIcon("/home/tumbler/GolandProjects/go-jw-toggl/toggl_bw.png")
			}
			return
		} else {
			if prev == nil || prev.Description != e.Description {
				j.indicator.SetStatus(appindicator.StatusAttention)
				j.indicator.SetLabel(shorten(e.Description), "Guid")
				j.indicator.SetIconFull("/home/tumbler/GolandProjects/go-jw-toggl/toggl.png", "Descr")
				j.showNotification("JW Toggl", e.Description)
			}
		}

	}
}

func shorten(desc string) string {
	l := len(desc)
	if l > 10 {
		desc = " " + desc[:10] + "..."
	} else {
		desc = " " + desc
	}
	return desc
}
