package db

import (
	"context"
	"errors"
	"fmt"
	"github.com/jinzhu/gorm"
	"go-jw-toggl/jwtoggl/gtoggl"
	"go-jw-toggl/jwtoggl/gtoggl/gtclient"
	"go-jw-toggl/jwtoggl/gtoggl/gthttp"
	"go-jw-toggl/jwtoggl/gtoggl/gtproject"
	"go-jw-toggl/jwtoggl/gtoggl/gttag"
	"go-jw-toggl/jwtoggl/gtoggl/gttimentry"
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
	store      EntryStore
	httpClient *http.Client
	sub        *gtwebsocket.TogglWebsocketClient
	projects   map[uint64]*gtproject.Project
	tags       map[uint64]*gttag.Tag
	clients    map[uint64]*gtclient.Client
	defaultWId uint64
	db         *gorm.DB

	onCurrentEntry    OnCurrentEntryCallback
	onProjectsChanged OnProjectsChangedCallback
	onTagsChanged     OnTagsChangedCallback
	onEntriesChanged  OnEntriesChangedCallback
	onClientsChanged  OnClientsChangedCallback
}

func NewDB(token string, database *gorm.DB) (*DB, error) {
	httpClient := &http.Client{Timeout: time.Second}
	client, err := gtoggl.NewClient(token, gthttp.SetHttpClient(httpClient))
	if err != nil {
		return nil, err
	}
	sub := gtwebsocket.NewClient(token)
	db := &DB{
		db: database,
		store: EntryStore{
			index: make(map[uint64]int),
			list:  make(gttimeentry.TimeEntries, 0),
		},
		client:     client,
		httpClient: httpClient,
		sub:        sub,
		projects:   make(map[uint64]*gtproject.Project),
		tags:       make(map[uint64]*gttag.Tag),
		clients:    make(map[uint64]*gtclient.Client),
	}
	db.Migrate()
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
	if db.store.Update(*entry) {
		db.onCurrentEntry(db.getCurrentEntry(), false)
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
	old := db.getCurrentEntry()
	db.store.Update(*entry)
	if old != nil && old.Id != entry.Id && tools.IsCurrent(*entry) {
		old.Stop = entry.Start
		old.Duration = int64(old.Stop.Sub(old.Start).Seconds())
		db.store.Update(*old)
		db.onEntriesChanged(db.ListEntries())
	}
	db.onCurrentEntry(entry, notify)
	return nil
}

func (db *DB) StopEntry(entry *gttimeentry.TimeEntry) error {
	if entry == nil {
		return nil
	}
	current := db.getCurrentEntry()
	if current == nil || current.Id != entry.Id {
		return errors.New("current entry de-sync")
	}
	entry, err := db.client.TimeentryClient.Stop(entry)
	if err != nil {
		return err
	}
	db.store.Update(*entry)
	db.onCurrentEntry(nil, false)
	db.onEntriesChanged(db.ListEntries())
	return nil
}

func (db *DB) ContinueEntry(id uint64) error {
	var entry *gttimeentry.TimeEntry
	var err error
	if prev, ok := db.getEntryById(id); ok {
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
	entry := Entry{TimeEntry: gttimeentry.TimeEntry{Id: id}}
	entry.Delete(db.db)
	db.onEntriesChanged(db.ListEntries())
	return nil
}

func (db *DB) SyncCurrentEntry() error {
	old := db.getCurrentEntry()
	entry, err := db.client.TimeentryClient.Current()
	if err != nil {
		return err
	}
	if entry != nil && db.store.Update(*entry) {
		current := db.getCurrentEntry()
		if !tools.EntriesEqual(old, current) {
			db.onCurrentEntry(current, true)
		}
	}
	return nil
}

func (db *DB) getCurrentEntry() *gttimeentry.TimeEntry {
	var entry Entry
	if res := db.db.Where("stop = ?", time.Time{}).First(&entry); res.RecordNotFound() {
		return nil
	} else if res.Error != nil {
		log.Fatal(res.Error)
	}
	return &entry.TimeEntry
}

func (db *DB) SyncLastEntries() error {
	end := time.Now()
	start := end.Add(-72 * time.Hour)
	old := db.getCurrentEntry()
	entries, err := db.client.TimeentryClient.GetRange(start, end)
	if err != nil {
		return err
	}
	if db.saveEntries(entries) {
		current := db.getCurrentEntry()
		if !tools.EntriesEqual(current, old) {
			db.onCurrentEntry(current, true)
		}
		db.onEntriesChanged(db.ListEntries())
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
		db.onTagsChanged(db.ListTags())
	}
	if db.saveClients(user.Clients) || force {
		db.onClientsChanged(db.ListClients())
	}
	if db.saveProjects(user.Projects) || force {
		db.onProjectsChanged(db.ListProjects())
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
	old := db.getCurrentEntry()
	entry := Entry{TimeEntry: timeEntry}
	if action == gtwebsocket.ActionDelete {
		entry.Delete(db.db)
		current := db.getCurrentEntry()
		if !tools.EntriesEqual(old, current) {
			// Обновился текущий
			db.onCurrentEntry(current, true)
		} else {
			db.onEntriesChanged(db.ListEntries())
		}

	} else if action == gtwebsocket.ActionUpdate || action == gtwebsocket.ActionInsert {
		if db.store.Update(timeEntry) {
			current := db.getCurrentEntry()
			if !tools.EntriesEqual(old, current) {
				// Обновился текущий
				db.onCurrentEntry(current, true)
			}
			db.onEntriesChanged(db.ListEntries())

		}
	} else {
		return fmt.Errorf("unhandled action: %s", action)
	}
	return nil
}

func (db *DB) Migrate() {
	// Миграция схем
	db.db.AutoMigrate(&Tag{}, &Client{}, &Project{}, &Entry{})
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

	queryset := NewTags(tags)
	if changed, err := queryset.Sync(db.db, time.Now()); err != nil {
		log.Fatal("Tags sync error: ", err)
		return false
	} else {
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

	queryset := NewClients(clients)
	if changed, err := queryset.Sync(db.db, time.Now()); err != nil {
		log.Fatal("Clients sync error: ", err)
		return false
	} else {
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

	queryset := NewProjects(projects)
	if changed, err := queryset.Sync(db.db, time.Now()); err != nil {
		log.Fatal("Projects sync error: ", err)
		return false
	} else {
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

	queryset := NewEntries(entries)
	if changed, err := queryset.Sync(db.db, time.Now()); err != nil {
		log.Fatal("Entries sync error: ", err)
		return false
	} else {
		return changed
	}
}

func (db *DB) ListProjects() gtproject.Projects {
	queryset := make(Projects, 0)
	db.db.Order("name ASC").Find(&queryset)
	result := make(gtproject.Projects, len(queryset))
	for i, p := range queryset {
		result[i] = p.Project
	}
	return result
}

func (db *DB) ListTags() gttag.Tags {
	queryset := make(Tags, 0)
	db.db.Order("name ASC").Find(&queryset)
	result := make(gttag.Tags, len(queryset))
	for i, t := range queryset {
		result[i] = t.Tag
	}
	return result
}

func (db *DB) ListClients() gtclient.Clients {
	queryset := make(Clients, 0)
	db.db.Find(&queryset)
	result := make(gtclient.Clients, len(queryset))
	for i, c := range queryset {
		result[i] = c.Client
	}
	return result
}

func (db *DB) ListEntries() gttimeentry.TimeEntries {
	queryset := make(Entries, 0)
	db.db.Find(&queryset)
	result := make(gttimeentry.TimeEntries, len(queryset))
	for i, c := range queryset {
		result[i] = c.TimeEntry
	}
	return result
}

func (db *DB) getEntryById(id uint64) (*gttimeentry.TimeEntry, bool) {
	var entry Entry
	if res := db.db.Where("id = ?", id).First(&entry); res.RecordNotFound() {
		return nil, false
	} else if res.Error != nil {
		log.Fatal(res.Error)
	}
	return &entry.TimeEntry, true
}
