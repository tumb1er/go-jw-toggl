package state

import (
	"context"
	"errors"
	"go-jw-toggl/jwtoggl/dto"
	"go-jw-toggl/jwtoggl/gtoggl"
	"go-jw-toggl/jwtoggl/gtoggl/gthttp"
	"go-jw-toggl/jwtoggl/gtoggl/gttimeentry"
	"go-jw-toggl/jwtoggl/gtoggl/gtwebsocket"
	"go-jw-toggl/jwtoggl/store"
	"log"
	"net/http"
	"time"
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
	httpClient := &http.Client{Timeout: time.Second}
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

func (db *DB) UpdateCurrentEntry(entry dto.TimeEntry) error {
	timeEntry := &gttimeentry.TimeEntry{
		Id:          entry.GetTogglID(),
		Description: entry.GetDescription(),
		Start:       entry.GetStart(),
		Stop:        entry.GetStop(),
		Tags:        entry.GetTags(),
		Pid:         entry.GetProjectTogglID(),
		WId:         db.defaultWId,
		CreatedWith: "go-jw-toggl",
	}
	if timeEntry.Stop.IsZero() {
		timeEntry.Duration = -timeEntry.Start.Unix()
	}
	timeEntry, err := db.client.TimeentryClient.Update(timeEntry)
	if err != nil {
		return err
	}
	storeEntry := store.NewTimeEntry(timeEntry)
	if db.update(&storeEntry) {
		db.onCurrentEntry(db.s.TimeEntries().Current(), false)
	}
	return nil
}

func (db *DB) CreateNewEntry(entry dto.TimeEntry) error {
	return db.createNewEntry(entry, false)
}

func (db *DB) createNewEntry(entry dto.TimeEntry, notify bool) error {
	timeEntry := &gttimeentry.TimeEntry{
		Description: entry.GetDescription(),
		Start:       entry.GetStart(),
		Stop:        entry.GetStop(),
		Duration:    -entry.GetStart().Unix(),
		Tags:        entry.GetTags(),
		Pid:         entry.GetProjectTogglID(),
		WId:         db.defaultWId,
		CreatedWith: "go-jw-toggl",
	}
	var err error
	timeEntry, err = db.client.TimeentryClient.Create(timeEntry)
	if err != nil {
		return err
	}

	storeEntry := store.TimeEntry{
		TimeEntry: *timeEntry,
		Tag:       "",
	}

	old := db.s.TimeEntries().Current()
	db.update(&storeEntry)
	if !old.Nil() && old.Id != storeEntry.Id && storeEntry.IsCurrent() {
		old.Stop = storeEntry.Start
		old.Duration = int64(old.Stop.Sub(old.Start).Seconds())
		db.update(&old)
		db.onTimeEntriesChanged(db.s.TimeEntries().List())
	}
	db.onCurrentEntry(storeEntry, notify)
	return nil
}

func (db *DB) StopEntry(entry dto.TimeEntry) error {
	if entry.Nil() {
		return nil
	}
	current := db.s.TimeEntries().Current()
	if current.Nil() || current.Id != entry.GetTogglID() {
		return errors.New("current entry de-sync")
	}
	timeEntry := &gttimeentry.TimeEntry{Id: entry.GetTogglID()}
	timeEntry, err := db.client.TimeentryClient.Stop(timeEntry)
	if err != nil {
		return err
	}
	storeEntry := store.NewTimeEntry(timeEntry)
	db.update(&storeEntry)
	db.onCurrentEntry(store.TimeEntry{}, false)
	db.onTimeEntriesChanged(db.s.TimeEntries().List())
	return nil
}

func (db *DB) ContinueEntry(id uint64) error {
	var entry *gttimeentry.TimeEntry
	var err error
	if prev := db.s.TimeEntries().Get(id); prev != nil {
		entry = &prev.TimeEntry
	} else {
		entry, err = db.client.TimeentryClient.Get(id)
		if err != nil {
			return err
		}
	}
	entry.Id = 0
	entry.Start = time.Now().Local()
	entry.Stop = time.Time{}
	entry.Duration = 0
	return db.createNewEntry(store.NewTimeEntry(entry), true)
}

func (db *DB) DeleteEntry(id uint64) error {
	if err := db.client.TimeentryClient.Delete(id); err != nil {
		return err
	}
	entries := db.s.TimeEntries()
	entries.Delete(id)
	db.onTimeEntriesChanged(entries.List())
	return nil
}

func (db *DB) SyncCurrentEntry(force bool) error {
	old := db.s.TimeEntries().Current()
	entry, err := db.client.TimeentryClient.Current()
	if err != nil {
		return err
	}
	storeEntry := store.NewTimeEntry(entry)
	if entry != nil && db.update(&storeEntry) || force {
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
	start := end.Add(-72 * time.Hour)
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
	if timeEntry.Id == 0 {
		log.Fatal("empty timeEntry id")
	}
	entries := db.s.TimeEntries()
	existing := entries.Get(timeEntry.Id)
	if existing == nil {
		entries.Create(timeEntry)
		return true
	}

	if existing.Equals(timeEntry) {
		// Нет изменений
		return false
	}

	entries.Update(timeEntry)
	return true
}

func (db *DB) GetCurrentEntry() store.TimeEntry {
	return db.s.TimeEntries().Current()
}
