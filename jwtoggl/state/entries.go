package state

//go:generate bash -c "sed 's/Foos/TimeEntries/g;s/Foo/TimeEntry/g;s/foo/gttimeentry/g' foo.tpl > entries_gen.go"

import (
	"errors"
	"fmt"
	"go-jw-toggl/jwtoggl/dto"
	"go-jw-toggl/jwtoggl/gtoggl/gttimeentry"
	"go-jw-toggl/jwtoggl/gtoggl/gtwebsocket"
	"go-jw-toggl/jwtoggl/store"
	"go-jw-toggl/jwtoggl/tools"
	"log"
	"time"
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
	storeEntry := entries.FromTimeEntry(timeEntry)
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
		if entries.CreateOrUpdate(&storeEntry) {
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

func (db *TimeEntriesDB) UpdateCurrentEntry(entry dto.TimeEntrySaver) error {
	entries := db.store().TimeEntries()
	storeEntry := entries.GetByPK(entry.GetPK())
	if storeEntry == nil {
		panic("Can't find entry to update")
	}
	project := db.store().Projects().GetByPK(entry.GetProjectPK())
	if project == nil {
		panic("Can't find project for entry")
	}
	storeEntry.SetTimeEntry(entry, *project)
	storeEntry.MarkModified()
	if entries.CreateOrUpdate(storeEntry) {
		db.onCurrentEntry(entries.Current(), false)
		if !debugSyncChanged {
			go db.updateEntry(storeEntry)
		}
	}
	return nil
}

func (db *TimeEntriesDB) updateEntry(entry *store.TimeEntry) error {
	timeEntry := &entry.TimeEntry
	timeEntry.WId = db.wid()
	log.Printf("Updating entry: %+v\n", timeEntry)

	if timeEntry, err := db.toggl().TimeentryClient.Update(timeEntry); err != nil {
		log.Printf("Update entry error: %+v\n", timeEntry)
		return err
	} else {
		log.Printf("Update entry result: %+v\n", timeEntry)

		entry.ComputeUpdated()
		entry.MarkFetched()

		db.store().TimeEntries().Update(entry)
		return nil
	}
}

func (db *TimeEntriesDB) CreateEntry(entry dto.TimeEntrySaver) error {
	project := db.store().Projects().GetByPK(entry.GetProjectPK())
	if project == nil {
		panic("Can't find project for entry")
	}
	storeEntry := store.TimeEntry{}
	storeEntry.SetTimeEntry(entry, *project)
	storeEntry.MarkModified()
	entries := db.store().TimeEntries()
	old := entries.Current()
	entries.Create(&storeEntry)
	if !old.Nil() && old.PK != storeEntry.PK {
		old.Stop = storeEntry.Start
		entries.Update(&old)
		db.onTimeEntriesChanged(entries.List())
	}
	db.onCurrentEntry(storeEntry, false)

	if !debugSyncChanged {
		go db.createEntry(storeEntry)
	}

	return nil
}

func (db *TimeEntriesDB) createEntry(entry store.TimeEntry) error {
	entry.CreatedWith = "go-jw-toggl"
	entry.WId = db.wid()
	timeEntry := &entry.TimeEntry
	log.Printf("Start create entry: %+v\n", timeEntry)
	if timeEntry, err := db.toggl().TimeentryClient.Create(timeEntry); err != nil {
		log.Printf("Create entry error: %+v\n", err)
		return err
	} else {
		log.Printf("Create entry result: %+v\n", timeEntry)
		entry.TimeEntry = *timeEntry

		entry.ComputeUpdated()
		entry.MarkFetched()

		db.store().TimeEntries().Update(&entry)
		return nil
	}
}

func (db *TimeEntriesDB) StopEntry(pk uint) error {
	entries := db.store().TimeEntries()
	current := entries.Current()
	if current.Nil() || current.PK != pk {
		return errors.New("current entry de-sync")
	}
	current.MarkStopped()
	entries.Update(&current)
	db.onCurrentEntry(store.TimeEntry{}, false)
	db.onTimeEntriesChanged(entries.List())
	if !debugSyncChanged {
		go db.stopEntry(current)
	}
	return nil
}

func (db *TimeEntriesDB) stopEntry(entry store.TimeEntry) {
	timeEntry := &entry.TimeEntry
	log.Printf("Send stop: %+v\n", timeEntry)
	if timeEntry, err := db.toggl().TimeentryClient.Update(timeEntry); err != nil {
		log.Printf("TimeEntry stop error: %+v\n", err)
	} else {
		log.Printf("Stop result: %+v\n", entry)
		entry.TimeEntry = *timeEntry

		entry.ComputeUpdated()
		entry.MarkFetched()

		db.store().TimeEntries().Update(&entry)
	}
}

func (db *TimeEntriesDB) ContinueEntry(pk uint) error {
	var storeEntry *store.TimeEntry
	if storeEntry = db.store().TimeEntries().GetByPK(pk); storeEntry == nil {
		panic("Can't continue entry")
	}

	storeEntry.Start = time.Now().Local()
	return db.CreateEntry(storeEntry)
}

func (db *TimeEntriesDB) DeleteEntry(pk uint) error {
	entries := db.store().TimeEntries()
	storeEntry := entries.GetByPK(pk)
	if storeEntry == nil {
		panic("Can't find entry to delete")
	}
	entries.MarkDeleted(pk)
	db.onTimeEntriesChanged(entries.List())
	if !debugSyncChanged {
		go db.deleteEntry(storeEntry)
	}
	return nil
}

func (db *TimeEntriesDB) deleteEntry(storeEntry *store.TimeEntry) error {
	log.Printf("Start delete entry: %d\n", storeEntry.Id)
	if err := db.toggl().TimeentryClient.Delete(storeEntry.Id); err != nil {
		log.Printf("Delete error: %+v\n", err)
		return err
	} else {
		log.Printf("Delete %d successfully\n", storeEntry.Id)
		db.store().TimeEntries().Delete(storeEntry.PK)
		return nil
	}
}

func (db *TimeEntriesDB) SyncCurrentEntry(force bool) error {
	entries := db.store().TimeEntries()
	old := entries.Current()
	entry, err := db.toggl().TimeentryClient.Current()
	if err != nil {
		return err
	}
	if entry == nil {
		return nil
	}
	storeEntry := entries.FromTimeEntry(*entry)
	if entries.CreateOrUpdate(&storeEntry) || force {
		current := entries.Current()
		if old.Nil() && current.Nil() {
			// prevent notify change nil to nil
			log.Println("Skip notify change current entry nil to nil")
			return nil
		}
		if !current.Equals(old) || force {
			db.onCurrentEntry(current, true)
		}
	}
	return nil
}

func (db *TimeEntriesDB) SyncLastEntries(force bool) error {
	end := time.Now()
	start := end.Add(-7 * 24 * time.Hour)
	entries := db.store().TimeEntries()
	old := entries.Current()
	timeEntries, err := db.toggl().TimeentryClient.GetRange(start, end)
	if err != nil {
		return err
	}
	if db.saveTimeEntries(timeEntries) || force {
		current := entries.Current()
		if !current.Equals(old) {
			db.onCurrentEntry(current, true)
		}
		db.onTimeEntriesChanged(entries.List())
	}
	return nil
}

func (db *TimeEntriesDB) SyncChanges() error {
	entries := db.store().TimeEntries()
	for _, entry := range entries.DeletedEntries() {
		if err := db.deleteEntry(&entry); err != nil {
			return err
		}
	}
	for _, entry := range entries.NewEntries() {
		if err := db.createEntry(entry); err != nil {
			return err
		}
	}
	for _, entry := range entries.ChangedEntries() {
		if err := db.updateEntry(&entry); err != nil {
			return err
		}
	}
	return nil
}

func (db *TimeEntriesDB) GetCurrentEntry() store.TimeEntry {
	return db.store().TimeEntries().Current()
}
