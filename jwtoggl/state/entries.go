package state

//go:generate bash -c "sed 's/Foos/TimeEntries/g;s/Foo/TimeEntry/g;s/foo/gttimeentry/g' foo.tpl > entries_gen.go"

import (
	"fmt"
	"go-jw-toggl/jwtoggl/dto"
	"go-jw-toggl/jwtoggl/gtoggl/gttimeentry"
	"go-jw-toggl/jwtoggl/gtoggl/gtwebsocket"
	"go-jw-toggl/jwtoggl/store"
	"go-jw-toggl/jwtoggl/tools"
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
	if action == gtwebsocket.ActionDelete {
		entries.Delete(timeEntry.Id)
		current := entries.Current()
		if !tools.TimeEntriesEqual(&old.TimeEntry, &current.TimeEntry) {
			// Обновился текущий
			db.onCurrentEntry(current, true)
		} else {
			db.onTimeEntriesChanged(entries.List())
		}

	} else if action == gtwebsocket.ActionUpdate || action == gtwebsocket.ActionInsert {
		storeEntry := store.NewTimeEntry(&timeEntry)
		if db.update(&storeEntry) {
			current := entries.Current()
			if !tools.TimeEntriesEqual(&old.TimeEntry, &current.TimeEntry) {
				// Обновился текущий
				db.onCurrentEntry(current, true)
			}
			db.onTimeEntriesChanged(entries.List())

		}
	} else {
		return fmt.Errorf("unhandled action: %s", action)
	}
	return nil
}
