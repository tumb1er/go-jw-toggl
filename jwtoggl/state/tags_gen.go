package state

// gttag.tpl предоставляет общие методы для работы со всеми DTO Toggle.

import (
	"go-jw-toggl/jwtoggl/gtoggl/gttag"
	"go-jw-toggl/jwtoggl/gtoggl/gtuser"
	"go-jw-toggl/jwtoggl/store"
	"log"
	"time"
)

func (db *TagsDB) OnTagsChanged(cb OnTagsChangedCallback) {
	db.onTagsChanged = cb
}

func (db *TagsDB) Subscribe() {
	db.subscriber().OnTagAction(db.onTagAction)
}

func (db *TagsDB) saveUserData(user *gtuser.User, force bool) {
	if db.saveTags(user.Tags) || force {
		db.onTagsChanged(db.store().Tags().List())
	}
}

func (db *TagsDB) saveTags(items gttag.Tags) bool {
	if changed, err := db.store().Tags().Sync(items, time.Now(), db.wid()); err != nil {
		log.Fatal("Tags sync error: ", err)
		return false
	} else {
		log.Println("Tags changed: ", changed)
		return changed
	}
}

func (db *TagsDB) ListTags() store.Tags {
	return db.store().Tags().List()
}

// bar.tpl предоставляет общие методы для справочников (которых нет у TimeEntriesDB)

func (db *TagsDB) onTagAction(action string, item gttag.Tag) error {
	log.Printf("Tag %s %s, sync user", item.Name, action)
	return db.SyncUser(false)
}
