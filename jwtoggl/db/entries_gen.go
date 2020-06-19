package db

import (
	"go-jw-toggl/jwtoggl/gtoggl/gttimeentry"
	"go-jw-toggl/jwtoggl/gtoggl/gtuser"
	"log"
	"time"
)

func (db *TimeEntriesDB) OnTimeEntriesChanged(cb OnTimeEntriesChangedCallback) {
	db.onTimeEntriesChanged = cb
}

func (db *TimeEntriesDB) Subscribe() {
	db.Sub().OnTimeEntryAction(db.onTimeEntryAction)
}

func (db *TimeEntriesDB) saveUserData(user *gtuser.User, force bool) {
	if db.saveTimeEntries(user.TimeEntries) || force {
		db.onTimeEntriesChanged(db.Store().TimeEntries().List())
	}
}

func (db *TimeEntriesDB) saveTimeEntries(items gttimeentry.TimeEntries) bool {
	if changed, err := db.Store().TimeEntries().Sync(items, time.Now(), db.WId()); err != nil {
		log.Fatal("TimeEntries sync error: ", err)
		return false
	} else {
		log.Println("TimeEntries changed: ", changed)
		return changed
	}
}

func (db *TimeEntriesDB) ListTimeEntries() gttimeentry.TimeEntries {
	return db.Store().TimeEntries().List()
}
