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
		db.onCurrentEntry(db.store.Current(), false)
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
	old := db.store.Current()
	db.store.Update(*entry)
	if old != nil && old.Id != entry.Id && tools.IsCurrent(*entry) {
		old.Stop = entry.Start
		old.Duration = int64(old.Stop.Sub(old.Start).Seconds())
		db.store.Update(*old)
		db.onEntriesChanged(db.store.List())
	}
	db.onCurrentEntry(entry, notify)
	return nil
}

func (db *DB) StopEntry(entry *gttimeentry.TimeEntry) error {
	if entry == nil {
		return nil
	}
	current := db.store.Current()
	if current == nil || current.Id != entry.Id {
		return errors.New("current entry de-sync")
	}
	entry, err := db.client.TimeentryClient.Stop(entry)
	if err != nil {
		return err
	}
	db.store.Update(*entry)
	db.onCurrentEntry(nil, false)
	db.onEntriesChanged(db.store.List())
	return nil
}

func (db *DB) ContinueEntry(id uint64) error {
	var entry *gttimeentry.TimeEntry
	var err error
	if prev, ok := db.store.Get(id); ok {
		entry = &prev
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
	entry := gttimeentry.TimeEntry{Id: id}
	db.store.Delete(entry)
	db.onEntriesChanged(db.store.List())
	return nil
}

func (db *DB) SyncCurrentEntry() error {
	old := db.store.Current()
	entry, err := db.client.TimeentryClient.Current()
	if err != nil {
		return err
	}
	if entry != nil && db.store.Update(*entry) {
		current := db.store.Current()
		if !tools.EntriesEqual(old, current) {
			db.onCurrentEntry(current, true)
		}
	}
	return nil
}

func (db *DB) SyncLastEntries() error {
	end := time.Now()
	start := end.Add(-72 * time.Hour)
	old := db.store.Current()
	entries, err := db.client.TimeentryClient.GetRange(start, end)
	if err != nil {
		return err
	} else {
		if db.store.BulkUpdate(entries) {
			current := db.store.Current()
			if !tools.EntriesEqual(current, old) {
				db.onCurrentEntry(current, true)
			}
			db.onEntriesChanged(db.store.List())
		}
	}
	return nil
}

func (db *DB) SyncUser() error {
	user, err := db.client.UserClient.Get(true)
	if err != nil {
		return err
	}
	db.defaultWId = user.DefaultWId
	db.saveTags(user.Tags)
	db.saveClients(user.Clients)
	db.saveProjects(user.Projects)
	var changed bool
	projects := make(gtproject.Projects, 0, len(user.Projects))
	for _, project := range user.Projects {
		if project.WId != user.DefaultWId || !project.Active {
			continue
		} else {
			projects = append(projects, project)
		}
		existing, ok := db.projects[project.Id]
		if !ok || !tools.ProjectsEqual(existing, &project) {
			db.projects[project.Id] = &project
			changed = true
		}
	}
	if changed && db.onProjectsChanged != nil {
		db.onProjectsChanged(projects)
	}

	changed = false
	tags := make(gttag.Tags, 0, len(user.Tags))
	for _, tag := range user.Tags {
		if tag.WId != user.DefaultWId {
			continue
		} else {
			tags = append(tags, tag)
		}
		existing, ok := db.tags[tag.Id]
		if !ok || !tools.TagsEqual(existing, &tag) {
			db.tags[tag.Id] = &tag
			changed = true
		}
	}
	if changed && db.onTagsChanged != nil {
		db.onTagsChanged(tags)
	}

	changed = false
	clients := make(gtclient.Clients, 0, len(user.Clients))
	for _, client := range user.Clients {
		if client.WId != user.DefaultWId {
			continue
		} else {
			clients = append(clients, client)
		}
		existing, ok := db.clients[client.Id]
		if !ok || !tools.ClientsEqual(existing, &client) {
			db.clients[client.Id] = &client
			changed = true
		}
	}
	if changed && db.onClientsChanged != nil {
		db.onClientsChanged(clients)
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
	return db.SyncUser()
}

func (db *DB) onClientAction(action string, client gtclient.Client) error {
	log.Printf("client %s %s, sync user", client.Name, action)
	return db.SyncUser()
}

func (db *DB) onTagAction(action string, tag gttag.Tag) error {
	log.Printf("tag %s %s, sync user", tag.Name, action)
	return db.SyncUser()
}

func (db *DB) onTimeEntryAction(action string, entry gttimeentry.TimeEntry) error {
	old := db.store.Current()
	if action == gtwebsocket.ActionDelete {
		if db.store.Delete(entry) {
			current := db.store.Current()
			if !tools.EntriesEqual(old, current) {
				// Обновился текущий
				db.onCurrentEntry(current, true)
			} else {
				db.onEntriesChanged(db.store.List())
			}
		}

	} else if action == gtwebsocket.ActionUpdate || action == gtwebsocket.ActionInsert {
		if db.store.Update(entry) {
			current := db.store.Current()
			if !tools.EntriesEqual(old, current) {
				// Обновился текущий
				db.onCurrentEntry(current, true)
			}
			db.onEntriesChanged(db.store.List())

		}
	} else {
		return fmt.Errorf("unhandled action: %s", action)
	}
	return nil
}

func (db *DB) Migrate() {
	// Миграция схем
	db.db.AutoMigrate(&Tag{}, &Client{}, &Project{})
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

func (db *DB) ListProjects() gtproject.Projects {
	queryset := make(Projects, 0)
	db.db.Find(&queryset)
	result := make(gtproject.Projects, len(queryset))
	for i, p := range queryset {
		result[i] = p.Project
	}
	return result
}

func (db *DB) ListTags() gttag.Tags {
	queryset := make(Tags, 0)
	db.db.Find(&queryset)
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
