package db

import (
	"github.com/tumb1er/gtoggl"
	"github.com/tumb1er/gtoggl-api/gttimentry"
)

type DB struct {
	client *gtoggl.TogglClient
	Entry  *gttimeentry.TimeEntry
}

func NewDB(token string) (*DB, error) {
	client, err := gtoggl.NewClient(token)
	if err != nil {
		return nil, err
	}
	return &DB{client: client}, nil
}

func (db *DB) SyncCurrentEntry() error {
	entry, err := db.client.TimeentryClient.Current()
	if err != nil {
		return err
	}
	db.Entry = entry
	return nil
}

func (db DB) GetCurrentEntry() *gttimeentry.TimeEntry {
	return db.Entry
}
