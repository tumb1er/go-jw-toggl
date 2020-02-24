package db

import (
	"context"
	"github.com/tumb1er/gtoggl"
	"github.com/tumb1er/gtoggl-api/gttimentry"
	"go-jw-toggl/jwtoggl/gtws"
	"log"
)

type OnCurrentEntryCallback func(entry *gttimeentry.TimeEntry)

type DB struct {
	client *gtoggl.TogglClient
	sub    *gtws.TogglWebsocketClient
	Entry  *gttimeentry.TimeEntry

	onCurrentEntry OnCurrentEntryCallback
}

func NewDB(token string) (*DB, error) {
	client, err := gtoggl.NewClient(token)
	if err != nil {
		return nil, err
	}
	sub, err := gtws.NewClient(token)
	if err != nil {
		return nil, err
	}
	return &DB{client: client, sub: sub}, nil
}

func (db *DB) OnCurrentEntry(cb OnCurrentEntryCallback) {
	db.onCurrentEntry = cb
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

func (db *DB) Subscribe(ctx context.Context) {
	db.sub.OnTimeEntryAction(db.onTimeEntryAction)
	go db.sub.Listen(ctx)
}

func (db *DB) onTimeEntryAction(action string, entry gttimeentry.TimeEntry) error {
	if action == gtws.ActionInsert {
		log.Println("INSERT")
		if entry.Stop.IsZero() {
			// New time entry started, replacing current.
			db.Entry = &entry
			db.onCurrentEntry(db.Entry)
		}
	}
	return nil
}
