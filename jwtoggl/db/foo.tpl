package db

import (
	"go-jw-toggl/jwtoggl/gtoggl/foo"
	"log"
	"time"
)

type OnFoosChangedCallback func(foo.Foos)

func (db *DB) OnFoosChanged(cb OnFoosChangedCallback) {
	db.onFoosChanged = cb
}

func (db *DB) saveFoos(items foo.Foos) bool {
	if changed, err := db.store.Foos().Sync(items, time.Now(), db.defaultWId); err != nil {
		log.Fatal("Foos sync error: ", err)
		return false
	} else {
		log.Println("Foos changed: ", changed)
		return changed
	}
}

func (db *DB) ListFoos() foo.Foos {
	return db.store.Foos().List()
}

