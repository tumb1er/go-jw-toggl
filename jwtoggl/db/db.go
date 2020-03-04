package db

import (
	"context"
	"errors"
	"fmt"
	"go-jw-toggl/jwtoggl/gtoggl"
	"go-jw-toggl/jwtoggl/gtoggl/gtproject"
	"go-jw-toggl/jwtoggl/gtoggl/gttag"
	"go-jw-toggl/jwtoggl/gtoggl/gttimentry"
	"go-jw-toggl/jwtoggl/gtoggl/gtwebsocket"
	"go-jw-toggl/jwtoggl/tools"
	"log"
	"time"
)

type OnCurrentEntryCallback func(entry *gttimeentry.TimeEntry, external bool)
type OnProjectsChangedCallback func(projects gtproject.Projects)
type OnTagsChangedCallback func(tags gttag.Tags)

type DB struct {
	client     *gtoggl.TogglClient
	sub        *gtwebsocket.TogglWebsocketClient
	Entry      *gttimeentry.TimeEntry
	projects   map[uint64]*gtproject.Project
	tags       map[uint64]*gttag.Tag
	defaultWId uint64

	onCurrentEntry    OnCurrentEntryCallback
	onProjectsChanged OnProjectsChangedCallback
	onTagsChanged     OnTagsChangedCallback
}

func NewDB(token string) (*DB, error) {
	client, err := gtoggl.NewClient(token)
	if err != nil {
		return nil, err
	}
	sub := gtwebsocket.NewClient(token)
	return &DB{
		client:   client,
		sub:      sub,
		projects: make(map[uint64]*gtproject.Project),
		tags:     make(map[uint64]*gttag.Tag),
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

func (db *DB) UpdateEntry(entry *gttimeentry.TimeEntry) error {
	// FIXME: compare with current entry
	entry.Wid = db.defaultWId
	entry.CreatedWith = "go-jw-toggl"
	entry, err := db.client.TimeentryClient.Update(entry)
	if err != nil {
		return err
	}
	db.Entry = entry
	db.onCurrentEntry(db.Entry, false)
	return nil
}

func (db *DB) CreateNewEntry(entry *gttimeentry.TimeEntry) error {
	entry.Wid = db.defaultWId
	entry.CreatedWith = "go-jw-toggl"
	entry.Start = time.Now()
	entry, err := db.client.TimeentryClient.Start(entry)
	if err != nil {
		return err
	}
	db.Entry = entry
	db.onCurrentEntry(db.Entry, false)
	return nil
}

func (db *DB) StopEntry(entry *gttimeentry.TimeEntry) error {
	if entry == nil {
		return nil
	}
	if db.Entry.Id != entry.Id {
		return errors.New("current entry de-sync")
	}
	_, err := db.client.TimeentryClient.Stop(entry)
	if err != nil {
		return err
	}
	db.onCurrentEntry(nil, false)
	return nil
}

func (db *DB) SyncCurrentEntry() error {
	entry, err := db.client.TimeentryClient.Current()
	if err != nil {
		return err
	}
	if !tools.EntriesEqual(db.Entry, entry) {
		db.Entry = entry
		db.onCurrentEntry(db.Entry, true)
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
	if action == gtwebsocket.ActionInsert {

		if entry.Stop.IsZero() {
			log.Println("INSERT new running entry")
			// New time entry started, replacing current.
			db.Entry = &entry
			db.onCurrentEntry(db.Entry, true)
			return nil
		} else {
			// Inserted entry directly to history - is it possible?
			return fmt.Errorf("can't handle insert to history: %+v", entry)
		}
	}

	if action == gtwebsocket.ActionDelete {
		if entry.Stop.IsZero() {
			log.Println("DELETE running entry")
			// Currently running entry is completely deleted.
			db.Entry = nil
			db.onCurrentEntry(nil, true)
			return nil
		} else {
			// Deleted historical entry
			// TODO: delete entry from local history
			return nil
		}
	}

	if action == gtwebsocket.ActionUpdate {
		if entry.Stop.IsZero() {
			// Current entry is updated
			db.Entry = &entry
			db.onCurrentEntry(db.Entry, true)
			return nil
		}
		if db.Entry != nil && db.Entry.Id == entry.Id {
			// Updating current entry, and it's stopped
			db.Entry = nil
			db.onCurrentEntry(nil, true)
			return nil
		} else {
			// TODO: Update entry in local history
			return nil
		}
	}
	return fmt.Errorf("unhandled action: %s", action)
}
