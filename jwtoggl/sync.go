package jwtoggl

import (
	"context"
	"go-jw-toggl/jwtoggl/gtoggl/gtclient"
	"go-jw-toggl/jwtoggl/gtoggl/gtproject"
	"go-jw-toggl/jwtoggl/gtoggl/gttag"
	"go-jw-toggl/jwtoggl/gtoggl/gttimentry"
	"log"
	"time"
)

func (j *jApp) StartSync(ctx context.Context) {
	j.db.OnCurrentEntry(j.OnCurrentEntryChanged)
	j.db.OnProjectsChanged(j.OnProjectsChanged)
	j.db.OnTagsChanged(j.OnTagsChanged)
	j.db.OnClientsChanged(j.OnClientsChanged)
	j.db.OnEntriesChanged(j.entries.SetEntries)
	if err := j.db.SyncUser(); err != nil {
		log.Fatal("Project sync error:", err)
	}
	log.Println("Start periodic sync")
	j.periodicSync(ctx)
}

func (j *jApp) OnProjectsChanged(projects gtproject.Projects) {
	j.form.SetProjects(projects)
	j.entries.SetProjects(projects)
}

func (j *jApp) OnTagsChanged(tags gttag.Tags) {
	j.form.SetTags(tags)
}

func (j *jApp) OnClientsChanged(clients gtclient.Clients) {
	j.entries.SetClients(clients)
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

func (j *jApp) OnCurrentEntryChanged(entry *gttimeentry.TimeEntry, external bool) {
	log.Println("OnCurrentEntryChanged")
	j.tray.SetCurrentEntry(entry)
	j.form.SetTimeEntry(entry)
	if external {
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
