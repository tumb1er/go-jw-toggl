package state

//go:generate bash -c "sed 's/Foos/TimeEntries/g;s/Foo/TimeEntry/g;s/foo/gttimeentry/g' foo.tpl > entries_gen.go"

import (
	"fmt"
	"go-jw-toggl/jwtoggl/dto"
	"go-jw-toggl/jwtoggl/gtoggl/gttimeentry"
	"go-jw-toggl/jwtoggl/gtoggl/gtwebsocket"
	"go-jw-toggl/jwtoggl/tools"
	"log"
)

type OnCurrentEntryCallback func(entry dto.TimeEntry, external bool)

type OnTimeEntriesChangedCallback func(dto.TimeEntries)

type TimeEntriesDB struct {
	common

	onCurrentEntry       OnCurrentEntryCallback
	onTimeEntriesChanged OnTimeEntriesChangedCallback
}

func (db *DB) OnCurrentEntry(cb OnCurrentEntryCallback) {
	db.onCurrentEntry = cb
}

func (db *TimeEntriesDB) onTimeEntryAction(action string, timeEntry gttimeentry.TimeEntry) error {
	entries := db.store().TimeEntries()
	old := entries.Current()
	storeEntry := db.store().TimeEntries().FromTimeEntry(timeEntry)
	if action == gtwebsocket.ActionDelete {
		entries.Delete(storeEntry.PK)
		current := entries.Current()
		if !tools.TimeEntriesEqual(&old.TimeEntry, &current.TimeEntry) {
			// Обновился текущий
			db.onCurrentEntry(current, true)
		} else {
			db.onTimeEntriesChanged(entries.List())
		}

	} else if action == gtwebsocket.ActionUpdate || action == gtwebsocket.ActionInsert {
		if db.update(storeEntry) {
			log.Printf("onTimeEntryAction update: %+v\n", old)
			current := entries.Current()
			log.Printf("onTimeEntryAction current: %+v\n", current)
			if !tools.TimeEntriesEqual(&old.TimeEntry, &current.TimeEntry) {
				// Обновился текущий
				log.Printf("onCurrentEntry: %+v\n", current)
				db.onCurrentEntry(current, true)
			} else {
				log.Println("Current entry not changed.")
			}
			db.onTimeEntriesChanged(entries.List())

		}
	} else {
		return fmt.Errorf("unhandled action: %s", action)
	}
	return nil
}
