package db

import (
	"context"
	"fmt"
	"github.com/tumb1er/gtoggl"
	"github.com/tumb1er/gtoggl-api/gttimentry"
	"github.com/tumb1er/gtoggl-api/gtwebsocket"
	"log"
)

type OnCurrentEntryCallback func(entry *gttimeentry.TimeEntry)

type DB struct {
	client *gtoggl.TogglClient
	sub    *gtwebsocket.TogglWebsocketClient
	Entry  *gttimeentry.TimeEntry

	onCurrentEntry OnCurrentEntryCallback
}

func NewDB(token string) (*DB, error) {
	client, err := gtoggl.NewClient(token)
	if err != nil {
		return nil, err
	}
	sub, err := gtwebsocket.NewClient(token)
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
