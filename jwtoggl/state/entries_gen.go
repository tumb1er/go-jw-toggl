package state

// gttimeentry.tpl предоставляет общие методы для работы со всеми DTO Toggle.

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
	db.subscriber().OnTimeEntryAction(db.onTimeEntryAction)
}

func (db *TimeEntriesDB) saveUserData(user *gtuser.User, force bool) {
	if db.saveTimeEntries(user.TimeEntries) || force {
		db.onTimeEntriesChanged(db.store().TimeEntries().List())
	}
}

func (db *TimeEntriesDB) saveTimeEntries(items gttimeentry.TimeEntries) bool {
	if changed, err := db.store().TimeEntries().Sync(items, time.Now(), db.wid()); err != nil {
		log.Fatal("TimeEntries sync error: ", err)
		return false
	} else {
		log.Println("TimeEntries changed: ", changed)
		return changed
	}
}

func (db *TimeEntriesDB) ListTimeEntries() gttimeentry.TimeEntries {
	return db.store().TimeEntries().List()
}
