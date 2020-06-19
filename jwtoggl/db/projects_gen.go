package db

import (
	"go-jw-toggl/jwtoggl/gtoggl/gtproject"
	"log"
	"time"
)

type OnProjectsChangedCallback func(gtproject.Projects)

func (db *DB) OnProjectsChanged(cb OnProjectsChangedCallback) {
	db.onProjectsChanged = cb
}

func (db *DB) saveProjects(items gtproject.Projects) bool {
	if changed, err := db.store.Projects().Sync(items, time.Now(), db.defaultWId); err != nil {
		log.Fatal("Projects sync error: ", err)
		return false
	} else {
		log.Println("Projects changed: ", changed)
		return changed
	}
}

func (db *DB) ListProjects() gtproject.Projects {
	return db.store.Projects().List()
}
