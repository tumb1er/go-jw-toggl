package db

import (
	"context"
	"fmt"
	"go-jw-toggl/jwtoggl/gtoggl"
	"go-jw-toggl/jwtoggl/gtoggl/gtproject"
	"go-jw-toggl/jwtoggl/gtoggl/gttag"
	"go-jw-toggl/jwtoggl/gtoggl/gttimentry"
	"go-jw-toggl/jwtoggl/gtoggl/gtwebsocket"
	"log"
)

type OnCurrentEntryCallback func(entry *gttimeentry.TimeEntry)
type OnProjectsChangedCallback func(projects gtproject.Projects)
type OnTagsChangedCallback func(tags gttag.Tags)

type DB struct {
	client   *gtoggl.TogglClient
	sub      *gtwebsocket.TogglWebsocketClient
	Entry    *gttimeentry.TimeEntry
	projects map[uint64]*gtproject.Project
	tags     map[uint64]*gttag.Tag

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

func (db *DB) SyncCurrentEntry() error {
	entry, err := db.client.TimeentryClient.Current()
	if err != nil {
		return err
	}
	if !entriesEqual(db.Entry, entry) {
		db.Entry = entry
		db.onCurrentEntry(db.Entry)
	}
	return nil
}

func (db *DB) SyncUser() error {
	user, err := db.client.UserClient.Get(true)
	if err != nil {
		return err
	}
	var changed bool
	for _, project := range user.Projects {
		existing, ok := db.projects[project.Id]
		if !ok || !projectsEqual(existing, &project) {
			db.projects[project.Id] = &project
			changed = true
		}
	}
	if changed && db.onProjectsChanged != nil {
		db.onProjectsChanged(user.Projects)
	}

	changed = false
	for _, tag := range user.Tags {
		existing, ok := db.tags[tag.Id]
		if !ok || !tagsEqual(existing, &tag) {
			db.tags[tag.Id] = &tag
			changed = true
		}
	}
	if changed && db.onTagsChanged != nil {
		db.onTagsChanged(user.Tags)
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
			db.onCurrentEntry(db.Entry)
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
			db.onCurrentEntry(nil)
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
			db.onCurrentEntry(db.Entry)
			return nil
		}
		if db.Entry != nil && db.Entry.Id == entry.Id {
			// Updating current entry, and it's stopped
			db.Entry = nil
			db.onCurrentEntry(nil)
			return nil
		} else {
			// TODO: Update entry in local history
			return nil
		}
	}
	return fmt.Errorf("unhandled action: %s", action)
}
