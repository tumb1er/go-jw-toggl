package jwtoggl

import (
	"context"
	"go-jw-toggl/jwtoggl/gtoggl/gttimentry"
	"log"
	"time"
)

func (j *jApp) StartSync(ctx context.Context) {
	j.db.OnCurrentEntry(j.onCurrentEntryChanged)
	j.db.OnProjectsChanged(j.form.SetProjects)
	j.db.OnTagsChanged(j.form.SetTags)
	j.db.OnEntriesChanged(j.entries.SetEntries)
	if err := j.db.SyncUser(); err != nil {
		log.Fatal("Project sync error:", err)
	}
	log.Println("Start periodic sync")
	j.periodicSync(ctx)
}

func (j *jApp) periodicSync(ctx context.Context) {
	t := time.NewTicker(time.Second * 10)
	t2 := time.NewTicker(time.Second)
	j.db.StartListen(ctx)
	j.doSync()
	for {
		select {
		case <-t.C:
			j.doSync()
		case <-t2.C:
			j.syncDuration()
		case <-ctx.Done():
			log.Println("Stop periodic sync")
			return
		}
	}
}

func (j *jApp) doSync() {
	log.Println("start sync")
	defer log.Println("sync done")
	if err := j.db.SyncCurrentEntry(); err != nil {
		log.Println("Sync current entry error:", err)
	}
	if err := j.db.SyncLastEntries(); err != nil {
		log.Println("Sync last entries error:", err)
	}
}

func (j *jApp) syncDuration() {
	j.form.UpdateDuration()
}

func (j *jApp) onCurrentEntryChanged(entry *gttimeentry.TimeEntry, external bool) {
	log.Println("OnCurrentEntryChanged")
	j.tray.SetCurrentEntry(entry)
	if external {
		j.form.SetTimeEntry(entry)
		var text string
		if entry != nil {
			text = entry.Description
		} else {
			text = "Timer stopped"
		}
		if err := j.tray.ShowNotification("JW Toggl", text); err != nil {
			log.Fatal(" notification error:", err)
		}

	}
}
