package db

import (
	"go-jw-toggl/jwtoggl/gtoggl/gttag"
	"log"
	"time"
)

type OnTagsChangedCallback func(gttag.Tags)

func (db *DB) OnTagsChanged(cb OnTagsChangedCallback) {
	db.onTagsChanged = cb
}

func (db *DB) saveTags(items gttag.Tags) bool {
	if changed, err := db.store.Tags().Sync(items, time.Now(), db.defaultWId); err != nil {
		log.Fatal("Tags sync error: ", err)
		return false
	} else {
		log.Println("Tags changed: ", changed)
		return changed
	}
}

func (db *DB) ListTags() gttag.Tags {
	return db.store.Tags().List()
}
