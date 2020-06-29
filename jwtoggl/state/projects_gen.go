package state

// gtproject.tpl предоставляет общие методы для работы со всеми DTO Toggle.

import (
	"go-jw-toggl/jwtoggl/gtoggl/gtproject"
	"go-jw-toggl/jwtoggl/gtoggl/gtuser"
	"go-jw-toggl/jwtoggl/store"
	"log"
)

func (db *ProjectsDB) OnProjectsChanged(cb OnProjectsChangedCallback) {
	db.onProjectsChanged = cb
}

func (db *ProjectsDB) Subscribe() {
	db.subscriber().OnProjectAction(db.onProjectAction)
}

func (db *ProjectsDB) saveUserData(user *gtuser.User, force bool) {
	if db.saveProjects(user.Projects) || force {
		db.onProjectsChanged(db.store().Projects().List())
	}
}

func (db *ProjectsDB) saveProjects(items gtproject.Projects) bool {
	if changed, err := db.store().Projects().Sync(items, db.wid()); err != nil {
		log.Fatal("Projects sync error: ", err)
		return false
	} else {
		log.Println("Projects changed: ", changed)
		return changed
	}
}

func (db *ProjectsDB) ListProjects() store.Projects {
	return db.store().Projects().List()
}

// bar.tpl предоставляет общие методы для справочников (которых нет у TimeEntriesDB)

func (db *ProjectsDB) onProjectAction(action string, item gtproject.Project) error {
	log.Printf("Project %s %s, sync user", item.Name, action)
	return db.SyncUser(false)
}
