package jwtoggl

import (
	"context"
	"github.com/dawidd6/go-appindicator"
	gttimeentry "github.com/tumb1er/gtoggl-api/gttimentry"
	"log"
	"time"
)

func (j *jApp) periodicSync(ctx context.Context) {
	log.Println("Start periodic sync")
	t := time.NewTicker(time.Second * 10)
	j.db.Subscribe(ctx)
	j.db.OnCurrentEntry(j.onCurrentEntryChanged)
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
	if err := j.db.SyncCurrentEntry(); err != nil {
		log.Println("Sync current entry error:", err)
	}
}

func (j *jApp) onCurrentEntryChanged(entry *gttimeentry.TimeEntry) {
	if entry == nil {
		j.showNotification("JW Toggl", "Timer stopped")
		j.indicator.SetLabel("", "")
		j.indicator.SetStatus(appindicator.StatusActive)
		j.indicator.SetIcon("/home/tumbler/GolandProjects/go-jw-toggl/toggl_bw.png")
	} else {
		j.indicator.SetStatus(appindicator.StatusAttention)
		j.indicator.SetLabel(shorten(entry.Description), "Guid")
		j.indicator.SetIconFull("/home/tumbler/GolandProjects/go-jw-toggl/toggl.png", "Descr")
		j.showNotification("JW Toggl", entry.Description)
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
