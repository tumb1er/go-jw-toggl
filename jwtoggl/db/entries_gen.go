package db

import (
	"go-jw-toggl/jwtoggl/gtoggl/gttimeentry"
	"log"
	"time"
)

type OnTimeEntriesChangedCallback func(gttimeentry.TimeEntries)

func (db *DB) OnTimeEntriesChanged(cb OnTimeEntriesChangedCallback) {
	db.onTimeEntriesChanged = cb
}

func (db *DB) saveTimeEntries(items gttimeentry.TimeEntries) bool {
	if changed, err := db.store.TimeEntries().Sync(items, time.Now(), db.defaultWId); err != nil {
		log.Fatal("TimeEntries sync error: ", err)
		return false
	} else {
		log.Println("TimeEntries changed: ", changed)
		return changed
	}
}

func (db *DB) ListTimeEntries() gttimeentry.TimeEntries {
	return db.store.TimeEntries().List()
}
