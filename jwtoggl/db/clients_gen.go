package db

import (
	"go-jw-toggl/jwtoggl/gtoggl/gtclient"
	"go-jw-toggl/jwtoggl/gtoggl/gtuser"
	"log"
	"time"
)

func (db *ClientsDB) OnClientsChanged(cb OnClientsChangedCallback) {
	db.onClientsChanged = cb
}

func (db *ClientsDB) Subscribe() {
	db.sub.OnClientAction(db.onClientAction)
}

func (db *ClientsDB) saveUserData(user *gtuser.User, force bool) {
	if db.saveClients(user.Clients) || force {
		db.onClientsChanged(db.store.Clients().List())
	}
}

func (db *ClientsDB) saveClients(items gtclient.Clients) bool {
	if changed, err := db.store.Clients().Sync(items, time.Now(), db.defaultWId); err != nil {
		log.Fatal("Clients sync error: ", err)
		return false
	} else {
		log.Println("Clients changed: ", changed)
		return changed
	}
}

func (db *ClientsDB) ListClients() gtclient.Clients {
	return db.store.Clients().List()
}

func (db *ClientsDB) onClientAction(action string, item gtclient.Client) error {
	log.Printf("Client %s %s, sync user", item.Name, action)
	return db.SyncUser(false)
}
