package db

import (
	"go-jw-toggl/jwtoggl/gtoggl/gttag"
	"go-jw-toggl/jwtoggl/gtoggl/gtuser"
	"log"
	"time"
)

func (db *TagsDB) OnTagsChanged(cb OnTagsChangedCallback) {
	db.onTagsChanged = cb
}

func (db *TagsDB) Subscribe() {
	db.Sub().OnTagAction(db.onTagAction)
}

func (db *TagsDB) saveUserData(user *gtuser.User, force bool) {
	if db.saveTags(user.Tags) || force {
		db.onTagsChanged(db.Store().Tags().List())
	}
}

func (db *TagsDB) saveTags(items gttag.Tags) bool {
	if changed, err := db.Store().Tags().Sync(items, time.Now(), db.WId()); err != nil {
		log.Fatal("Tags sync error: ", err)
		return false
	} else {
		log.Println("Tags changed: ", changed)
		return changed
	}
}

func (db *TagsDB) ListTags() gttag.Tags {
	return db.Store().Tags().List()
}

func (db *TagsDB) onTagAction(action string, item gttag.Tag) error {
	log.Printf("Tag %s %s, sync user", item.Name, action)
	return db.SyncUser(false)
}
