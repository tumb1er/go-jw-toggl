package db

import (
	"go-jw-toggl/jwtoggl/gtoggl/gtproject"
	"go-jw-toggl/jwtoggl/gtoggl/gtuser"
	"log"
	"time"
)

func (db *ProjectsDB) OnProjectsChanged(cb OnProjectsChangedCallback) {
	db.onProjectsChanged = cb
}

func (db *ProjectsDB) Subscribe() {
	db.Sub().OnProjectAction(db.onProjectAction)
}

func (db *ProjectsDB) saveUserData(user *gtuser.User, force bool) {
	if db.saveProjects(user.Projects) || force {
		db.onProjectsChanged(db.Store().Projects().List())
	}
}

func (db *ProjectsDB) saveProjects(items gtproject.Projects) bool {
	if changed, err := db.Store().Projects().Sync(items, time.Now(), db.WId()); err != nil {
		log.Fatal("Projects sync error: ", err)
		return false
	} else {
		log.Println("Projects changed: ", changed)
		return changed
	}
}

func (db *ProjectsDB) ListProjects() gtproject.Projects {
	return db.Store().Projects().List()
}

func (db *ProjectsDB) onProjectAction(action string, item gtproject.Project) error {
	log.Printf("Project %s %s, sync user", item.Name, action)
	return db.SyncUser(false)
}
