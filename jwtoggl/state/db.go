package state

import (
	"context"
	"errors"
	"go-jw-toggl/jwtoggl/gtoggl"
	"go-jw-toggl/jwtoggl/gtoggl/gthttp"
	"go-jw-toggl/jwtoggl/gtoggl/gttimeentry"
	"go-jw-toggl/jwtoggl/gtoggl/gtwebsocket"
	"go-jw-toggl/jwtoggl/tools"
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

func (db *DB) UpdateCurrentEntry(entry *gttimeentry.TimeEntry) error {
	entry.WId = db.defaultWId
	entry.CreatedWith = "go-jw-toggl"
	entry, err := db.client.TimeentryClient.Update(entry)
	if err != nil {
		return err
	}
	if db.update(*entry) {
		db.onCurrentEntry(db.s.TimeEntries().Current(), false)
	}
	return nil
}

func (db *DB) CreateNewEntry(entry *gttimeentry.TimeEntry) error {
	return db.createNewEntry(entry, false)
}

func (db *DB) createNewEntry(entry *gttimeentry.TimeEntry, notify bool) error {
	entry.WId = db.defaultWId
	entry.CreatedWith = "go-jw-toggl"
	var err error
	if entry.Start.IsZero() {
		entry, err = db.client.TimeentryClient.Start(entry)
	} else {
		if entry.Stop.IsZero() {
			entry.Duration = -entry.Start.Unix()
		}
		entry, err = db.client.TimeentryClient.Create(entry)
	}
	if err != nil {
		return err
	}
	old := db.s.TimeEntries().Current()
	db.update(*entry)
	if old != nil && old.Id != entry.Id && tools.IsCurrent(*entry) {
		old.Stop = entry.Start
		old.Duration = int64(old.Stop.Sub(old.Start).Seconds())
		db.update(*old)
		db.onTimeEntriesChanged(db.s.TimeEntries().List())
	}
	db.onCurrentEntry(entry, notify)
	return nil
}

func (db *DB) StopEntry(entry *gttimeentry.TimeEntry) error {
	if entry == nil {
		return nil
	}
	current := db.s.TimeEntries().Current()
	if current == nil || current.Id != entry.Id {
		return errors.New("current entry de-sync")
	}
	entry, err := db.client.TimeentryClient.Stop(entry)
	if err != nil {
		return err
	}
	db.update(*entry)
	db.onCurrentEntry(nil, false)
	db.onTimeEntriesChanged(db.s.TimeEntries().List())
	return nil
}

func (db *DB) ContinueEntry(id uint64) error {
	var entry *gttimeentry.TimeEntry
	var err error
	if prev := db.s.TimeEntries().Get(id); prev != nil {
		entry = prev
	} else {
		entry, err = db.client.TimeentryClient.Get(id)
		if err != nil {
			return err
		}
	}
	entry.Id = 0
	entry.Start = time.Time{}
	entry.Stop = time.Time{}
	entry.Duration = 0
	return db.createNewEntry(entry, true)
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
	if entry != nil && db.update(*entry) || force {
		current := db.s.TimeEntries().Current()
		if old == nil && current == nil {
			// prevent notify change nil to nil
			log.Println("Skip notify change current entry nil to nil")
			return nil
		}
		if !tools.TimeEntriesEqual(old, current) || force {
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
		if !tools.TimeEntriesEqual(current, old) {
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

func (db *DB) update(timeEntry gttimeentry.TimeEntry) bool {
	if timeEntry.Id == 0 {
		log.Fatal("empty timeEntry id")
	}
	entries := db.s.TimeEntries()
	existing := entries.Get(timeEntry.Id)
	if existing == nil {
		entries.Create(timeEntry)
		return true
	}

	if tools.TimeEntriesEqual(existing, &timeEntry) {
		// Нет изменений
		return false
	}

	entries.Update(timeEntry)
	return true
}

func (db *DB) GetCurrentEntry() *gttimeentry.TimeEntry {
	return db.s.TimeEntries().Current()
}
