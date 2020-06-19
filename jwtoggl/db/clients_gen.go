package db

import (
	"go-jw-toggl/jwtoggl/gtoggl/gtclient"
	"log"
	"time"
)

type OnClientsChangedCallback func(gtclient.Clients)

func (db *DB) OnClientsChanged(cb OnClientsChangedCallback) {
	db.onClientsChanged = cb
}

func (db *DB) saveClients(items gtclient.Clients) bool {
	if changed, err := db.store.Clients().Sync(items, time.Now(), db.defaultWId); err != nil {
		log.Fatal("Clients sync error: ", err)
		return false
	} else {
		log.Println("Clients changed: ", changed)
		return changed
	}
}

func (db *DB) ListClients() gtclient.Clients {
	return db.store.Clients().List()
}
