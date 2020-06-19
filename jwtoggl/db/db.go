package db

//go:generate bash -c "sed 's/Foo/Client/g;s/foo/gtclient/g' foo.tpl > clients_gen.go"
//go:generate bash -c "sed 's/Foos/TimeEntries/g;s/Foo/TimeEntry/g;s/foo/gttimeentry/g' foo.tpl > entries_gen.go"
//go:generate bash -c "sed 's/Foo/Project/g;s/foo/gtproject/g' foo.tpl > projects_gen.go"
//go:generate bash -c "sed 's/Foo/Tag/g;s/foo/gttag/g' foo.tpl > tags_gen.go"

import (
	"context"
	"errors"
	"fmt"
	"go-jw-toggl/jwtoggl/gtoggl"
	"go-jw-toggl/jwtoggl/gtoggl/gtclient"
	"go-jw-toggl/jwtoggl/gtoggl/gthttp"
	"go-jw-toggl/jwtoggl/gtoggl/gtproject"
	"go-jw-toggl/jwtoggl/gtoggl/gttag"
	"go-jw-toggl/jwtoggl/gtoggl/gttimeentry"
	"go-jw-toggl/jwtoggl/gtoggl/gtwebsocket"
	"go-jw-toggl/jwtoggl/tools"
	"log"
	"net/http"
	"time"
)

type OnCurrentEntryCallback func(entry *gttimeentry.TimeEntry, external bool)

type DB struct {
	client     *gtoggl.TogglClient
	httpClient *http.Client
	sub        *gtwebsocket.TogglWebsocketClient
	defaultWId uint64
	store      Store

	onCurrentEntry       OnCurrentEntryCallback
	onProjectsChanged    OnProjectsChangedCallback
	onTagsChanged        OnTagsChangedCallback
	onTimeEntriesChanged OnTimeEntriesChangedCallback
	onClientsChanged     OnClientsChangedCallback
}

func NewDB(token string, store Store) (*DB, error) {
	httpClient := &http.Client{Timeout: time.Second}
	client, err := gtoggl.NewClient(token, gthttp.SetHttpClient(httpClient))
	if err != nil {
		return nil, err
	}
	sub := gtwebsocket.NewClient(token)
	db := &DB{
		store:      store,
		client:     client,
		httpClient: httpClient,
		sub:        sub,
	}
	db.store.Migrate()
	return db, nil
}

func (db *DB) OnCurrentEntry(cb OnCurrentEntryCallback) {
	db.onCurrentEntry = cb
}

func (db *DB) UpdateCurrentEntry(entry *gttimeentry.TimeEntry) error {
	entry.WId = db.defaultWId
	entry.CreatedWith = "go-jw-toggl"
	entry, err := db.client.TimeentryClient.Update(entry)
	if err != nil {
		return err
	}
	if db.update(*entry) {
		db.onCurrentEntry(db.store.TimeEntries().Current(), false)
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
	old := db.store.TimeEntries().Current()
	db.update(*entry)
	if old != nil && old.Id != entry.Id && tools.IsCurrent(*entry) {
		old.Stop = entry.Start
		old.Duration = int64(old.Stop.Sub(old.Start).Seconds())
		db.update(*old)
		db.onTimeEntriesChanged(db.store.TimeEntries().List())
	}
	db.onCurrentEntry(entry, notify)
	return nil
}

func (db *DB) StopEntry(entry *gttimeentry.TimeEntry) error {
	if entry == nil {
		return nil
	}
	current := db.store.TimeEntries().Current()
	if current == nil || current.Id != entry.Id {
		return errors.New("current entry de-sync")
	}
	entry, err := db.client.TimeentryClient.Stop(entry)
	if err != nil {
		return err
	}
	db.update(*entry)
	db.onCurrentEntry(nil, false)
	db.onTimeEntriesChanged(db.store.TimeEntries().List())
	return nil
}

func (db *DB) ContinueEntry(id uint64) error {
	var entry *gttimeentry.TimeEntry
	var err error
	if prev := db.store.TimeEntries().Get(id); prev != nil {
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
	entries := db.store.TimeEntries()
	entries.Delete(id)
	db.onTimeEntriesChanged(entries.List())
	return nil
}

func (db *DB) SyncCurrentEntry(force bool) error {
	old := db.store.TimeEntries().Current()
	entry, err := db.client.TimeentryClient.Current()
	if err != nil {
		return err
	}
	if entry != nil && db.update(*entry) || force {
		current := db.store.TimeEntries().Current()
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
	old := db.store.TimeEntries().Current()
	entries, err := db.client.TimeentryClient.GetRange(start, end)
	if err != nil {
		return err
	}
	if db.saveTimeEntries(entries) || force {
		current := db.store.TimeEntries().Current()
		if !tools.TimeEntriesEqual(current, old) {
			db.onCurrentEntry(current, true)
		}
		db.onTimeEntriesChanged(db.store.TimeEntries().List())
	}
	return nil
}

func (db *DB) SyncUser(force bool) error {
	user, err := db.client.UserClient.Get(true)
	if err != nil {
		return err
	}
	db.defaultWId = user.DefaultWId
	if db.saveTags(user.Tags) || force {
		db.onTagsChanged(db.store.Tags().List())
	}
	if db.saveClients(user.Clients) || force {
		db.onClientsChanged(db.store.Clients().List())
	}
	if db.saveProjects(user.Projects) || force {
		db.onProjectsChanged(db.store.Projects().List())
	}
	return nil
}

func (db *DB) StartListen(ctx context.Context) {
	db.sub.OnTimeEntryAction(db.onTimeEntryAction)
	db.sub.OnProjectAction(db.onProjectAction)
	db.sub.OnClientAction(db.onClientAction)
	db.sub.OnTagAction(db.onTagAction)
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
	entries := db.store.TimeEntries()
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
	return db.store.TimeEntries().Current()
}

func (db *DB) onTimeEntryAction(action string, timeEntry gttimeentry.TimeEntry) error {
	entries := db.store.TimeEntries()
	old := entries.Current()
	if action == gtwebsocket.ActionDelete {
		entries.Delete(timeEntry.Id)
		current := entries.Current()
		if !tools.TimeEntriesEqual(old, current) {
			// Обновился текущий
			db.onCurrentEntry(current, true)
		} else {
			db.onTimeEntriesChanged(entries.List())
		}

	} else if action == gtwebsocket.ActionUpdate || action == gtwebsocket.ActionInsert {
		if db.update(timeEntry) {
			current := entries.Current()
			if !tools.TimeEntriesEqual(old, current) {
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

func (db *DB) onClientAction(action string, item gtclient.Client) error {
	log.Printf("Client %s %s, sync user", item.Name, action)
	return db.SyncUser(false)
}

func (db *DB) onProjectAction(action string, item gtproject.Project) error {
	log.Printf("Project %s %s, sync user", item.Name, action)
	return db.SyncUser(false)
}

func (db *DB) onTagAction(action string, item gttag.Tag) error {
	log.Printf("Tag %s %s, sync user", item.Name, action)
	return db.SyncUser(false)
}
