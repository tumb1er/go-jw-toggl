package state

import (
	"context"
	"errors"
	"go-jw-toggl/jwtoggl/dto"
	"go-jw-toggl/jwtoggl/gtoggl"
	"go-jw-toggl/jwtoggl/gtoggl/gthttp"
	"go-jw-toggl/jwtoggl/gtoggl/gtwebsocket"
	"go-jw-toggl/jwtoggl/store"
	"log"
	"net/http"
	"time"
)

const (
	debugSyncChanged = true
)

type DB struct {
	client     *gtoggl.TogglClient
	httpClient *http.Client
	sub        *gtwebsocket.TogglWebsocketClient
	defaultWId uint64
	s          Store

	ClientsDB
	ProjectsDB
	TagsDB
	TimeEntriesDB
}

func NewDB(token string, store Store) (*DB, error) {
	httpClient := &http.Client{Timeout: 10 * time.Second}
	client, err := gtoggl.NewClient(token, gthttp.SetHttpClient(httpClient))
	if err != nil {
		return nil, err
	}
	sub := gtwebsocket.NewClient(token)
	db := &DB{
		s:          store,
		client:     client,
		httpClient: httpClient,
		sub:        sub,
	}
	db.ClientsDB = ClientsDB{common: db}
	db.ProjectsDB = ProjectsDB{common: db}
	db.TagsDB = TagsDB{common: db}
	db.TimeEntriesDB = TimeEntriesDB{common: db}
	db.s.Migrate()
	return db, nil
}

func (db *DB) store() Store {
	return db.s
}

func (db *DB) subscriber() Subscriber {
	return db.sub
}

func (db *DB) wid() uint64 {
	return db.defaultWId
}

func (db *DB) UpdateCurrentEntry(entry dto.TimeEntrySaver) error {
	storeEntry := db.s.TimeEntries().GetByPK(entry.GetPK())
	if storeEntry == nil {
		panic("Can't find entry to update")
	}
	project := db.s.Projects().GetByPK(entry.GetProjectPK())
	if project == nil {
		panic("Can't find project for entry")
	}
	storeEntry.SetTimeEntry(entry, *project)
	storeEntry.MarkModified()
	if db.update(storeEntry) {
		db.onCurrentEntry(db.s.TimeEntries().Current(), false)
		if !debugSyncChanged {
			go db.updateEntry(storeEntry)
		}
	}

	return nil
}

func (db *DB) updateEntry(entry *store.TimeEntry) error {
	timeEntry := &entry.TimeEntry
	timeEntry.WId = db.defaultWId
	log.Printf("Updating entry: %+v\n", timeEntry)

	if timeEntry, err := db.client.TimeentryClient.Update(timeEntry); err != nil {
		log.Printf("Update entry error: %+v\n", timeEntry)
		return err
	} else {
		log.Printf("Update entry result: %+v\n", timeEntry)
		db.update(entry)
		return nil
	}
}

func (db *DB) CreateEntry(entry dto.TimeEntrySaver) error {
	project := db.s.Projects().GetByPK(entry.GetProjectPK())
	if project == nil {
		panic("Can't find project for entry")
	}
	storeEntry := store.TimeEntry{}
	storeEntry.SetTimeEntry(entry, *project)
	storeEntry.MarkModified()
	old := db.s.TimeEntries().Current()
	db.update(&storeEntry)
	if !old.Nil() && old.PK != storeEntry.PK {
		old.Stop = storeEntry.Start
		old.ComputeDuration()
		db.update(&old)
		db.onTimeEntriesChanged(db.s.TimeEntries().List())
	}
	db.onCurrentEntry(storeEntry, false)

	if !debugSyncChanged {
		go db.createEntry(storeEntry)
	}

	return nil
}

func (db *DB) createEntry(entry store.TimeEntry) error {
	entry.ComputeDuration()
	entry.CreatedWith = "go-jw-toggl"
	entry.WId = db.defaultWId
	timeEntry := &entry.TimeEntry
	log.Printf("Start create entry: %+v\n", timeEntry)
	if timeEntry, err := db.client.TimeentryClient.Create(timeEntry); err != nil {
		log.Printf("Create entry error: %+v\n", err)
		return err
	} else {
		log.Printf("Create entry result: %+v\n", timeEntry)
		entry.TimeEntry = *timeEntry

		entry.ComputeUpdated()
		entry.MarkFetched()

		db.update(&entry)
		return nil
	}
}

func (db *DB) StopEntry(pk uint) error {
	current := db.s.TimeEntries().Current()
	if current.Nil() || current.PK != pk {
		return errors.New("current entry de-sync")
	}
	current.MarkModified()
	current.Stop = *current.Modified
	db.update(&current)
	db.onCurrentEntry(store.TimeEntry{}, false)
	db.onTimeEntriesChanged(db.s.TimeEntries().List())
	if !debugSyncChanged {
		go db.stopEntry(current)
	}
	return nil
}

func (db *DB) stopEntry(entry store.TimeEntry) {
	timeEntry := &entry.TimeEntry
	log.Printf("Send stop: %+v\n", timeEntry)
	entry.ComputeDuration()
	timeEntry, err := db.client.TimeentryClient.Update(timeEntry)
	if err != nil {
		log.Printf("TimeEntry stop error: %+v\n", err)
	} else {
		log.Printf("Stop result: %+v\n", entry)
		entry.TimeEntry = *timeEntry
		db.update(&entry)
	}
}

func (db *DB) ContinueEntry(pk uint) error {
	var storeEntry *store.TimeEntry
	if storeEntry = db.s.TimeEntries().GetByPK(pk); storeEntry == nil {
		panic("Can't continue entry")
	}

	storeEntry.Id = 0
	storeEntry.Start = time.Now().Local()
	storeEntry.Stop = time.Time{}
	storeEntry.Duration = 0
	return db.CreateEntry(storeEntry)
}

func (db *DB) DeleteEntry(pk uint) error {
	storeEntry := db.s.TimeEntries().GetByPK(pk)
	if storeEntry == nil {
		panic("Can't find entry to delete")
	}
	entries := db.s.TimeEntries()
	entries.MarkDeleted(pk)
	db.onTimeEntriesChanged(entries.List())
	if !debugSyncChanged {
		go db.deleteEntry(storeEntry)
	}
	return nil
}

func (db *DB) deleteEntry(storeEntry *store.TimeEntry) error {
	log.Printf("Start delete entry: %d\n", storeEntry.Id)
	if err := db.client.TimeentryClient.Delete(storeEntry.Id); err != nil {
		log.Printf("Delete error: %+v\n", err)
		return err
	} else {
		log.Printf("Delete %d successfully\n", storeEntry.Id)
		db.s.TimeEntries().Delete(storeEntry.PK)
		return nil
	}
}

func (db *DB) SyncCurrentEntry(force bool) error {
	old := db.s.TimeEntries().Current()
	entry, err := db.client.TimeentryClient.Current()
	if err != nil {
		return err
	}
	if entry == nil {
		return nil
	}
	storeEntry := db.s.TimeEntries().FromTimeEntry(*entry)
	if storeEntry != nil && db.update(storeEntry) || force {
		current := db.s.TimeEntries().Current()
		if old.Nil() && current.Nil() {
			// prevent notify change nil to nil
			log.Println("Skip notify change current entry nil to nil")
			return nil
		}
		if current.Equals(old) || force {
			db.onCurrentEntry(current, true)
		}
	}
	return nil
}

func (db *DB) SyncLastEntries(force bool) error {
	end := time.Now()
	start := end.Add(-7 * 24 * time.Hour)
	old := db.s.TimeEntries().Current()
	entries, err := db.client.TimeentryClient.GetRange(start, end)
	if err != nil {
		return err
	}
	if db.saveTimeEntries(entries) || force {
		current := db.s.TimeEntries().Current()
		if !current.Equals(old) {
			db.onCurrentEntry(current, true)
		}
		db.onTimeEntriesChanged(db.s.TimeEntries().List())
	}
	return nil
}

func (db *DB) SyncUser(force bool) error {
	user, err := db.client.UserClient.Get(true)
	if err != nil {
		return err
	}
	db.defaultWId = user.DefaultWId
	db.ClientsDB.saveUserData(user, force)
	db.ProjectsDB.saveUserData(user, force)
	db.TagsDB.saveUserData(user, force)
	db.TimeEntriesDB.saveUserData(user, force)
	return nil
}

func (db *DB) SyncChanges() error {
	log.Println("Start sync changes")
	defer log.Println("Finish sync changes")
	for _, entry := range db.s.TimeEntries().DeletedEntries() {
		if err := db.deleteEntry(&entry); err != nil {
			return err
		}
	}
	for _, entry := range db.s.TimeEntries().NewEntries() {
		if err := db.createEntry(entry); err != nil {
			return err
		}
	}
	for _, entry := range db.s.TimeEntries().ChangedEntries() {
		if err := db.updateEntry(&entry); err != nil {
			return err
		}
	}
	return nil
}

func (db *DB) StartListen(ctx context.Context) {
	db.ClientsDB.Subscribe()
	db.ProjectsDB.Subscribe()
	db.TagsDB.Subscribe()
	db.TimeEntriesDB.Subscribe()
	go func() {
		err := db.sub.Listen(ctx)
		if err != nil {
			log.Fatal("WebSocket listen error:", err)
		}
	}()
}

func (db *DB) update(timeEntry *store.TimeEntry) bool {
	entries := db.s.TimeEntries()
	var existing *store.TimeEntry
	if timeEntry.Id > 0 {
		existing = entries.Get(timeEntry.Id)
	}
	if existing == nil {
		existing = entries.GetByPK(timeEntry.PK)
	}
	if existing == nil {
		entries.Create(timeEntry)
		return true
	}

	if existing.Equals(*timeEntry) {
		// Нет изменений
		return false
	}

	entries.Update(timeEntry)
	return true
}

func (db *DB) GetCurrentEntry() store.TimeEntry {
	return db.s.TimeEntries().Current()
}
