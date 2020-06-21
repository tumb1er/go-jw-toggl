package state

// gtclient.tpl предоставляет общие методы для работы со всеми DTO Toggle.

import (
	"go-jw-toggl/jwtoggl/gtoggl/gtclient"
	"go-jw-toggl/jwtoggl/gtoggl/gtuser"
	"go-jw-toggl/jwtoggl/store"
	"log"
	"time"
)

func (db *ClientsDB) OnClientsChanged(cb OnClientsChangedCallback) {
	db.onClientsChanged = cb
}

func (db *ClientsDB) Subscribe() {
	db.subscriber().OnClientAction(db.onClientAction)
}

func (db *ClientsDB) saveUserData(user *gtuser.User, force bool) {
	if db.saveClients(user.Clients) || force {
		db.onClientsChanged(db.store().Clients().List())
	}
}

func (db *ClientsDB) saveClients(items gtclient.Clients) bool {
	if changed, err := db.store().Clients().Sync(items, time.Now(), db.wid()); err != nil {
		log.Fatal("Clients sync error: ", err)
		return false
	} else {
		log.Println("Clients changed: ", changed)
		return changed
	}
}

func (db *ClientsDB) ListClients() store.Clients {
	return db.store().Clients().List()
}

// bar.tpl предоставляет общие методы для справочников (которых нет у TimeEntriesDB)

func (db *ClientsDB) onClientAction(action string, item gtclient.Client) error {
	log.Printf("Client %s %s, sync user", item.Name, action)
	return db.SyncUser(false)
}
