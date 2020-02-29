package jwtoggl

import (
	"context"
	"github.com/dawidd6/go-appindicator"
	"github.com/tumb1er/gtoggl-api/gtproject"
	"github.com/tumb1er/gtoggl-api/gttimentry"
	"log"
	"time"
)

func (j *jApp) StartSync(ctx context.Context) {
	j.db.OnCurrentEntry(j.onCurrentEntryChanged)
	j.db.OnProjectsChanged(j.onProjectsChanged)
	if err := j.db.SyncProjects(); err != nil {
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
	log.Printf("Projects: %+v\n", projects)
	if err := j.form.SetProjects(projects); err != nil {
		log.Fatal("Projects bind error:", err)
	}
	//j.window.ShowAll()
}

func (j *jApp) onCurrentEntryChanged(entry *gttimeentry.TimeEntry) {
	if entry == nil {
		j.UpdateCurrentEntry(CurrentEntryInfo{
			Title:  "JW Toggl",
			Body:   "Timer stopped",
			Label:  " Work!",
			Status: appindicator.StatusActive,
			Icon:   TogglIconBW,
			Menu:   "New...",
		})
	} else {
		j.UpdateCurrentEntry(CurrentEntryInfo{
			Title:  "JW Toggl",
			Body:   entry.Description,
			Label:  shorten(entry.Description),
			Status: appindicator.StatusAttention,
			Icon:   TogglIcon,
			Menu:   entry.Description,
		})
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
