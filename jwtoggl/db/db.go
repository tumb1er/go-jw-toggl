package db

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
type OnProjectsChangedCallback func(projects gtproject.Projects)
type OnTagsChangedCallback func(tags gttag.Tags)
type OnClientsChangedCallback func(clients gtclient.Clients)
type OnEntriesChangedCallback func(entries gttimeentry.TimeEntries)
type OnOnlineStatusChangedCallback func(online bool)

type DB struct {
	client     *gtoggl.TogglClient
	httpClient *http.Client
	sub        *gtwebsocket.TogglWebsocketClient
	defaultWId uint64
	store      Store

	onCurrentEntry    OnCurrentEntryCallback
	onProjectsChanged OnProjectsChangedCallback
	onTagsChanged     OnTagsChangedCallback
	onEntriesChanged  OnEntriesChangedCallback
	onClientsChanged  OnClientsChangedCallback
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

func (db *DB) OnProjectsChanged(cb OnProjectsChangedCallback) {
	db.onProjectsChanged = cb
}

func (db *DB) OnTagsChanged(cb OnTagsChangedCallback) {
	db.onTagsChanged = cb
}

func (db *DB) OnClientsChanged(cb OnClientsChangedCallback) {
	db.onClientsChanged = cb
}

func (db *DB) OnEntriesChanged(cb OnEntriesChangedCallback) {
	db.onEntriesChanged = cb
}

func (db *DB) UpdateCurrentEntry(entry *gttimeentry.TimeEntry) error {
	entry.Wid = db.defaultWId
	entry.CreatedWith = "go-jw-toggl"
	entry, err := db.client.TimeentryClient.Update(entry)
	if err != nil {
		return err
	}
	if db.update(*entry) {
		db.onCurrentEntry(db.store.Entries().Current(), false)
	}
	return nil
}

func (db *DB) CreateNewEntry(entry *gttimeentry.TimeEntry) error {
	return db.createNewEntry(entry, false)
}

func (db *DB) createNewEntry(entry *gttimeentry.TimeEntry, notify bool) error {
	entry.Wid = db.defaultWId
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
	old := db.store.Entries().Current()
	db.update(*entry)
	if old != nil && old.Id != entry.Id && tools.IsCurrent(*entry) {
		old.Stop = entry.Start
		old.Duration = int64(old.Stop.Sub(old.Start).Seconds())
		db.update(*old)
		db.onEntriesChanged(db.store.Entries().List())
	}
	db.onCurrentEntry(entry, notify)
	return nil
}

func (db *DB) StopEntry(entry *gttimeentry.TimeEntry) error {
	if entry == nil {
		return nil
	}
	current := db.store.Entries().Current()
	if current == nil || current.Id != entry.Id {
		return errors.New("current entry de-sync")
	}
	entry, err := db.client.TimeentryClient.Stop(entry)
	if err != nil {
		return err
	}
	db.update(*entry)
	db.onCurrentEntry(nil, false)
	db.onEntriesChanged(db.store.Entries().List())
	return nil
}

func (db *DB) ContinueEntry(id uint64) error {
	var entry *gttimeentry.TimeEntry
	var err error
	if prev := db.store.Entries().Get(id); prev != nil {
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
	entries := db.store.Entries()
	entries.Delete(id)
	db.onEntriesChanged(entries.List())
	return nil
}

func (db *DB) SyncCurrentEntry(force bool) error {
	old := db.store.Entries().Current()
	entry, err := db.client.TimeentryClient.Current()
	if err != nil {
		return err
	}
	if entry != nil && db.update(*entry) || force {
		current := db.store.Entries().Current()
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
	old := db.store.Entries().Current()
	entries, err := db.client.TimeentryClient.GetRange(start, end)
	if err != nil {
		return err
	}
	if db.saveEntries(entries) || force {
		current := db.store.Entries().Current()
		if !tools.TimeEntriesEqual(current, old) {
			db.onCurrentEntry(current, true)
		}
		db.onEntriesChanged(db.store.Entries().List())
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

func (db *DB) onProjectAction(action string, project gtproject.Project) error {
	log.Printf("project %s %s, sync user", project.Name, action)
	return db.SyncUser(false)
}

func (db *DB) onClientAction(action string, client gtclient.Client) error {
	log.Printf("client %s %s, sync user", client.Name, action)
	return db.SyncUser(false)
}

func (db *DB) onTagAction(action string, tag gttag.Tag) error {
	log.Printf("tag %s %s, sync user", tag.Name, action)
	return db.SyncUser(false)
}

func (db *DB) onTimeEntryAction(action string, timeEntry gttimeentry.TimeEntry) error {
	entries := db.store.Entries()
	old := entries.Current()
	if action == gtwebsocket.ActionDelete {
		entries.Delete(timeEntry.Id)
		current := entries.Current()
		if !tools.TimeEntriesEqual(old, current) {
			// Обновился текущий
			db.onCurrentEntry(current, true)
		} else {
			db.onEntriesChanged(entries.List())
		}

	} else if action == gtwebsocket.ActionUpdate || action == gtwebsocket.ActionInsert {
		if db.update(timeEntry) {
			current := entries.Current()
			if !tools.TimeEntriesEqual(old, current) {
				// Обновился текущий
				db.onCurrentEntry(current, true)
			}
			db.onEntriesChanged(entries.List())

		}
	} else {
		return fmt.Errorf("unhandled action: %s", action)
	}
	return nil
}

func (db *DB) saveTags(tags gttag.Tags) bool {
	c := len(tags)
	for i := 0; i < c; {
		if tags[i].WId != db.defaultWId {
			tags[i] = tags[c-1]
			c--
		} else {
			i++
		}
	}
	tags = tags[:c]

	if changed, err := db.store.Tags().Sync(tags, time.Now()); err != nil {
		log.Fatal("Tags sync error: ", err)
		return false
	} else {
		log.Println("Tags changed: ", changed)
		return changed
	}
}

func (db *DB) saveClients(clients gtclient.Clients) bool {
	c := len(clients)
	for i := 0; i < c; {
		if clients[i].WId != db.defaultWId {
			clients[i] = clients[c-1]
			c--
		} else {
			i++
		}
	}
	clients = clients[:c]

	if changed, err := db.store.Clients().Sync(clients, time.Now()); err != nil {
		log.Fatal("Clients sync error: ", err)
		return false
	} else {
		log.Println("Clients changed: ", changed)
		return changed
	}
}

func (db *DB) saveProjects(projects gtproject.Projects) bool {
	c := len(projects)
	for i := 0; i < c; {
		if projects[i].WId != db.defaultWId {
			projects[i] = projects[c-1]
			c--
		} else {
			i++
		}
	}
	projects = projects[:c]

	if changed, err := db.store.Projects().Sync(projects, time.Now()); err != nil {
		log.Fatal("Projects sync error: ", err)
		return false
	} else {
		log.Println("Projects changed: ", changed)
		return changed
	}
}

func (db *DB) saveEntries(entries gttimeentry.TimeEntries) bool {
	c := len(entries)
	for i := 0; i < c; {
		if entries[i].Wid != db.defaultWId {
			entries[i] = entries[c-1]
			c--
		} else {
			i++
		}
	}
	entries = entries[:c]

	if changed, err := db.store.Entries().Sync(entries, time.Now()); err != nil {
		log.Fatal("Entries sync error: ", err)
		return false
	} else {
		return changed
	}
}

func (db *DB) update(timeEntry gttimeentry.TimeEntry) bool {
	if timeEntry.Id == 0 {
		log.Fatal("empty timeEntry id")
	}
	entries := db.store.Entries()
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

func (db *DB) ListProjects() gtproject.Projects {
	return db.store.Projects().List()
}

func (db *DB) ListTags() gttag.Tags {
	return db.store.Tags().List()
}

func (db *DB) ListClients() gtclient.Clients {
	return db.store.Clients().List()
}

func (db *DB) GetCurrentEntry() *gttimeentry.TimeEntry {
	return db.store.Entries().Current()
}

func (db *DB) ListEntries() gttimeentry.TimeEntries {
	return db.store.Entries().List()
}
