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

func (db *DB) UpdateCurrentEntry(entry dto.TimeEntry) error {
	storeEntry := db.s.TimeEntries().GetByPK(entry.GetPK())
	if storeEntry == nil {
		panic("Can't find entry to update")
	}
	storeProject := db.s.Projects().GetByPK(entry.GetProject().GetPK())
	if storeProject == nil {
		panic("Can't find project for entry")
	}
	timeEntry := &storeEntry.TimeEntry
	timeEntry.Name = entry.GetName()
	timeEntry.Start = entry.GetStart()
	timeEntry.Stop = entry.GetStop()
	timeEntry.Tags = entry.GetTags()
	timeEntry.Pid = storeProject.Id
	timeEntry.WId = db.defaultWId
	timeEntry.CreatedWith = "go-jw-toggl"
	storeEntry.ComputeDuration()
	storeEntry.Tag = entry.GetTag()
	if db.update(storeEntry) {
		db.onCurrentEntry(db.s.TimeEntries().Current(), false)
		go func() {
			log.Printf("Updating entry: %+v\n", timeEntry)
			timeEntry, err := db.client.TimeentryClient.Update(timeEntry)
			if err != nil {
				log.Printf("Update entry error: %+v\n", timeEntry)
			} else {
				log.Printf("Update entry result: %+v\n", timeEntry)
				db.update(storeEntry)
			}
		}()
	}

	return nil
}

func (db *DB) CreateNewEntry(entry dto.TimeEntry) error {
	return db.createNewEntry(entry, false)
}

func (db *DB) createNewEntry(entry dto.TimeEntry, notify bool) error {
	storeProject := db.s.Projects().GetByPK(entry.GetProject().GetPK())
	if storeProject == nil {
		panic("Can't find project for new entry")
	}
	timeEntry := &gttimeentry.TimeEntry{
		Name:        entry.GetName(),
		Start:       entry.GetStart(),
		Stop:        entry.GetStop(),
		Duration:    -entry.GetStart().Unix(),
		Tags:        entry.GetTags(),
		Pid:         storeProject.Id,
		WId:         db.defaultWId,
		CreatedWith: "go-jw-toggl",
	}

	storeEntry := store.TimeEntry{
		TimeEntry: *timeEntry,
		Tag:       "",
	}

	old := db.s.TimeEntries().Current()
	db.update(&storeEntry)
	if !old.Nil() && old.PK != storeEntry.PK && storeEntry.IsCurrent() {
		old.Stop = storeEntry.Start
		old.Duration = int64(old.Stop.Sub(old.Start).Seconds())
		db.update(&old)
		db.onTimeEntriesChanged(db.s.TimeEntries().List())
	}
	db.onCurrentEntry(storeEntry, notify)

	go func() {
		log.Printf("Start create entry: %+v\n", timeEntry)
		var err error
		timeEntry, err = db.client.TimeentryClient.Create(timeEntry)
		if err != nil {
			log.Printf("Create entry error: %+v\n", err)
		} else {
			log.Printf("Create entry result: %+v\n", timeEntry)
			storeEntry.TimeEntry = *timeEntry
			db.update(&storeEntry)
		}
	}()
	return nil
}

func (db *DB) StopEntry(entry dto.TimeEntry) error {
	if entry.Nil() {
		return nil
	}
	current := db.s.TimeEntries().Current()
	if current.Nil() || current.PK != entry.GetPK() {
		return errors.New("current entry de-sync")
	}
	current.Stop = time.Now().Local()
	db.update(&current)
	db.onCurrentEntry(store.TimeEntry{}, false)
	db.onTimeEntriesChanged(db.s.TimeEntries().List())

	go func() {
		log.Printf("Send stop: %+v\n", current.TimeEntry)
		current.ComputeDuration()
		entry, err := db.client.TimeentryClient.Update(&current.TimeEntry)
		if err != nil {
			log.Printf("TimeEntry stop error: %+v\n", err)
		} else {
			log.Printf("Stop result: %+v\n", entry)
			current.TimeEntry = *entry
			db.update(&current)
		}
	}()
	return nil
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
	//storeEntry.Project = *db.store().Projects().Get(entry.Pid)
	//storeEntry.ProjectPK = storeEntry.Project.PK
	return db.createNewEntry(storeEntry, true)
}

func (db *DB) DeleteEntry(pk uint) error {
	storeEntry := db.s.TimeEntries().GetByPK(pk)
	if storeEntry == nil {
		panic("Can't find entry to delete")
	}
	if err := db.client.TimeentryClient.Delete(storeEntry.Id); err != nil {
		return err
	}
	entries := db.s.TimeEntries()
	entries.Delete(storeEntry.Id)
	db.onTimeEntriesChanged(entries.List())
	return nil
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
