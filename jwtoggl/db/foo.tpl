package db

import (
	"go-jw-toggl/jwtoggl/gtoggl/foo"
	"go-jw-toggl/jwtoggl/gtoggl/gtuser"
	"log"
	"time"
)

func (db *FoosDB) OnFoosChanged(cb OnFoosChangedCallback) {
	db.onFoosChanged = cb
}

func (db *FoosDB) Subscribe() {
    db.sub.OnFooAction(db.onFooAction)
}

func (db *FoosDB) saveUserData(user *gtuser.User, force bool) {
    if db.saveFoos(user.Foos) || force {
        db.onFoosChanged(db.store.Foos().List())
    }
}

func (db *FoosDB) saveFoos(items foo.Foos) bool {
	if changed, err := db.store.Foos().Sync(items, time.Now(), db.defaultWId); err != nil {
		log.Fatal("Foos sync error: ", err)
		return false
	} else {
		log.Println("Foos changed: ", changed)
		return changed
	}
}

func (db *FoosDB) ListFoos() foo.Foos {
	return db.store.Foos().List()
}

