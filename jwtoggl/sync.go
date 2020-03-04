package jwtoggl

import (
	"context"
	"go-jw-toggl/jwtoggl/gtoggl/gtproject"
	"go-jw-toggl/jwtoggl/gtoggl/gttag"
	"go-jw-toggl/jwtoggl/gtoggl/gttimentry"
	"log"
	"time"
)

func (j *jApp) StartSync(ctx context.Context) {
	j.db.OnCurrentEntry(j.onCurrentEntryChanged)
	j.db.OnProjectsChanged(j.onProjectsChanged)
	j.db.OnTagsChanged(j.onTagsChanged)
	if err := j.db.SyncUser(); err != nil {
		log.Fatal("Project sync error:", err)
	}
	j.periodicSync(ctx)
}

func (j *jApp) periodicSync(ctx context.Context) {
	log.Println("Start periodic sync")
	t := time.NewTicker(time.Second * 10)
	j.db.StartListen(ctx)
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

func (j *jApp) onProjectsChanged(projects gtproject.Projects) {
	if err := j.form.SetProjects(projects); err != nil {
		log.Fatal("Projects bind error:", err)
	}
}

func (j *jApp) onTagsChanged(tags gttag.Tags) {
	if err := j.form.SetTags(tags); err != nil {
		log.Fatal("Tags bind error:", err)
	}
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
