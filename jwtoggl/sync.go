package jwtoggl

import (
	"context"
	"go-jw-toggl/jwtoggl/dto"
	"log"
	"time"
)

func (j *jApp) StartSync(ctx context.Context) {
	j.db.OnCurrentEntry(j.OnCurrentEntryChanged)
	j.db.OnProjectsChanged(j.OnProjectsChanged)
	j.db.OnTagsChanged(j.OnTagsChanged)
	j.db.OnClientsChanged(j.OnClientsChanged)
	j.db.OnTimeEntriesChanged(j.OnTimeEntriesChanged)
	if err := j.db.SyncUser(true); err != nil {
		j.handleError(err, "Sync error")
		j.OnProjectsChanged(j.db.ListProjects())
		j.OnTagsChanged(j.db.ListTags())
		j.OnTimeEntriesChanged(j.db.ListTimeEntries())
	}
	if err := j.db.SyncCurrentEntry(true); err != nil {
		j.OnCurrentEntryChanged(j.db.GetCurrentEntry(), false)
	}
	log.Println("Start periodic sync")
	j.periodicSync(ctx)
}

func (j *jApp) syncUser() {
	j.handleError(j.db.SyncUser(false), "Sync error")
}

func (j *jApp) syncChanges() {
	j.handleError(j.db.SyncChanges(), "Changes sync error")
}

func (j *jApp) OnProjectsChanged(projects dto.Projects) {
	j.form.SetProjects(projects.DTO())
	// В истории проекты берутся из TimeEntry.Project, форсируем
	// синхронизацию записей.
	j.entries.SetEntries(j.db.ListTimeEntries().DTO())
}

func (j *jApp) OnTagsChanged(tags dto.Tags) {
	j.form.SetTags(tags.DTO())
}

func (j *jApp) OnClientsChanged(_ dto.Clients) {
	// Клиенты отображаются только в истории, а там они берутся из TimeEntry.Project.Client,
	// поэтому форсируем синхронизацию записей.
	j.entries.SetEntries(j.db.ListTimeEntries().DTO())
}

func (j *jApp) OnTimeEntriesChanged(entries dto.TimeEntries) {
	j.entries.SetEntries(entries.DTO())
}

func (j *jApp) periodicSync(ctx context.Context) {
	t := time.After(0)
	t2 := time.NewTicker(time.Second)
	j.db.StartListen(ctx)
	for {
		select {
		case <-t:
			t = j.doSync()
			log.Println("Sync time")
		case <-t2.C:
			j.syncDuration()
		case <-ctx.Done():
			log.Println("Stop periodic sync")
			return
		}
	}
}

func (j *jApp) doSync() <-chan time.Time {
	log.Println("start sync")
	defer log.Println("sync done")
	errCount := 0
	if err := j.db.SyncCurrentEntry(false); err != nil {
		j.handleError(err, "Sync current entry error")
		errCount += 1
	}
	if err := j.db.SyncLastEntries(false); err != nil {
		j.handleError(err, "Sync last entries error")
		errCount += 1
	}
	if err := j.db.SyncChanges(); err != nil {
		j.handleError(err, "Sync changes error")
		errCount += 1
	}

	if errCount == 0 {
		j.backoff = 0
	} else {
		j.backoff += 1
		log.Println("Sync backoff", time.Duration(15*j.backoff)*time.Minute)
	}
	return time.After(time.Duration(15*j.backoff) * time.Minute)
}

func (j *jApp) syncDuration() {
	j.form.UpdateDuration()
}

func (j *jApp) OnCurrentEntryChanged(entry dto.TimeEntry, external bool) {
	log.Printf("OnCurrentEntryChanged %+v\n", entry)
	j.tray.SetCurrentEntry(entry)
	j.form.SetTimeEntry(entry)
	if external {
		var text string
		if !entry.Nil() {
			text = entry.GetName()
		} else {
			text = "Timer stopped"
		}
		if err := j.tray.ShowNotification("JW Toggl", text); err != nil {
			log.Fatal(" notification error:", err)
		}
	}
}
