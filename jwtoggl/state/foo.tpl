package state

// foo.tpl предоставляет общие методы для работы со всеми DTO Toggle.

import (
	"go-jw-toggl/jwtoggl/gtoggl/foo"
	"go-jw-toggl/jwtoggl/gtoggl/gtuser"
	"go-jw-toggl/jwtoggl/store"
	"log"
)

func (db *FoosDB) OnFoosChanged(cb OnFoosChangedCallback) {
	db.onFoosChanged = cb
}

func (db *FoosDB) Subscribe() {
	db.subscriber().OnFooAction(db.onFooAction)
}

func (db *FoosDB) saveUserData(user *gtuser.User, force bool) {
	if db.saveFoos(user.Foos) || force {
		db.onFoosChanged(db.store().Foos().List())
	}
}

func (db *FoosDB) saveFoos(items foo.Foos) bool {
	if changed, err := db.store().Foos().Sync(items, db.wid()); err != nil {
		log.Fatal("Foos sync error: ", err)
		return false
	} else {
		log.Println("Foos changed: ", changed)
		return changed
	}
}

func (db *FoosDB) ListFoos() store.Foos {
	return db.store().Foos().List()
}
