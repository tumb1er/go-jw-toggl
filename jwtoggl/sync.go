package jwtoggl

import (
	"context"
	"errors"
	"go-jw-toggl/jwtoggl/dto"
	"go-jw-toggl/jwtoggl/gtoggl/gthttp"
	"log"
	"time"
)

const normalSyncInterval = 2 * time.Second
const quotaFallbackDelay = time.Hour

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
	quotaLimited := false
	quotaDelay := time.Duration(0)
	handleSyncError := func(err error, msg string) {
		j.handleError(err, msg)
		if retryAfter, limited := quotaRetryAfter(err); limited {
			quotaLimited = true
			if retryAfter > quotaDelay {
				quotaDelay = retryAfter
			}
			return
		}
		errCount++
	}
	if err := j.db.SyncCurrentEntry(false); err != nil {
		handleSyncError(err, "Sync current entry error")
	}
	if err := j.db.SyncLastEntries(false); err != nil {
		handleSyncError(err, "Sync last entries error")
	}
	if err := j.db.SyncChanges(); err != nil {
		handleSyncError(err, "Sync changes error")
	}
	if quotaLimited {
		delay := syncDelayForFailures(errCount, true, quotaDelay)
		log.Println("Sync quota backoff", delay)
		return time.After(delay)
	}

	if errCount == 0 {
		j.backoff = 0
	} else {
		j.backoff += 1
		log.Println("Sync backoff", time.Duration(15*j.backoff)*time.Minute)
	}
	return time.After(syncDelayForFailures(j.backoff, false, 0))
}

func quotaRetryAfter(err error) (time.Duration, bool) {
	var togglErr *gthttp.TogglError
	if errors.As(err, &togglErr) && togglErr.Code == 402 {
		return togglErr.RetryAfter, true
	}
	return 0, false
}

func syncDelayForFailures(errorCount int, quotaLimited bool, quotaRetryAfter time.Duration) time.Duration {
	if quotaLimited {
		if quotaRetryAfter > 0 {
			return quotaRetryAfter
		}
		return quotaFallbackDelay
	}
	return syncDelayForErrors(errorCount)
}

func syncDelayForErrors(errorCount int) time.Duration {
	if errorCount == 0 {
		return normalSyncInterval
	}
	return time.Duration(15*errorCount) * time.Minute
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
