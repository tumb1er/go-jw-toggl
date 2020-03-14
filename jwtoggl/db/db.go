package db

import (
	"context"
	"errors"
	"fmt"
	"go-jw-toggl/jwtoggl/gtoggl"
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
type OnEntriesChangedCallback func(entry []*gttimeentry.TimeEntry)

type DB struct {
	client     *gtoggl.TogglClient
	store      EntryStore
	httpClient *http.Client
	sub        *gtwebsocket.TogglWebsocketClient
	projects   map[uint64]*gtproject.Project
	tags       map[uint64]*gttag.Tag
	defaultWId uint64

	onCurrentEntry    OnCurrentEntryCallback
	onProjectsChanged OnProjectsChangedCallback
	onTagsChanged     OnTagsChangedCallback
	onEntriesChanged  OnEntriesChangedCallback
}

func NewDB(token string) (*DB, error) {
	httpClient := &http.Client{Timeout: time.Second}
	client, err := gtoggl.NewClient(token, gthttp.SetHttpClient(httpClient))
	if err != nil {
		return nil, err
	}
	sub := gtwebsocket.NewClient(token)
	return &DB{
		store: EntryStore{
			index: make(map[uint64]int),
			list:  make([]*gttimeentry.TimeEntry, 0),
		},
		client:     client,
		httpClient: httpClient,
		sub:        sub,
		projects:   make(map[uint64]*gtproject.Project),
		tags:       make(map[uint64]*gttag.Tag),
	}, nil
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
	if db.store.Update(entry) {
		db.onCurrentEntry(db.store.Current(), false)
	}
	return nil
}

func (db *DB) CreateNewEntry(entry *gttimeentry.TimeEntry) error {
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
	db.store.Update(entry)
	db.onCurrentEntry(db.store.Current(), false)
	return nil
}

func (db *DB) StopEntry(entry *gttimeentry.TimeEntry) error {
	if entry == nil {
		return nil
	}
	if db.store.Current().Id != entry.Id {
		return errors.New("current entry de-sync")
	}
	entry, err := db.client.TimeentryClient.Stop(entry)
	if err != nil {
		return err
	}
	db.store.Update(entry)
	db.onCurrentEntry(nil, false)
	db.onEntriesChanged(db.store.List())
	return nil
}

func (db *DB) SyncCurrentEntry() error {
	old := db.store.Current()
	entry, err := db.client.TimeentryClient.Current()
	if err != nil {
		db.httpClient.CloseIdleConnections()
		return err
	}
	if db.store.Update(entry) {
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
	var changed bool
	projects := make(gtproject.Projects, 0, len(user.Projects))
	for _, project := range user.Projects {
		if project.WId != user.DefaultWId {
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
	return nil
}

func (db *DB) StartListen(ctx context.Context) {
	db.sub.OnTimeEntryAction(db.onTimeEntryAction)
	go func() {
		err := db.sub.Listen(ctx)
		if err != nil {
			log.Fatal("WebSocket listen error:", err)
		}
	}()
}

func (db *DB) onTimeEntryAction(action string, entry gttimeentry.TimeEntry) error {
	old := db.store.Current()
	if action == gtwebsocket.ActionDelete {
		if db.store.Delete(&entry) {
			current := db.store.Current()
			if !tools.EntriesEqual(old, current) {
				// Обновился текущий
				db.onCurrentEntry(current, true)
			} else {
				db.onEntriesChanged(db.store.List())
			}
		}

	} else if action == gtwebsocket.ActionUpdate || action == gtwebsocket.ActionInsert {
		if db.store.Update(&entry) {
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
