package state

import (
	"context"
	"go-jw-toggl/jwtoggl/gtoggl"
	"go-jw-toggl/jwtoggl/gtoggl/gthttp"
	"go-jw-toggl/jwtoggl/gtoggl/gtwebsocket"
	"log"
	"net/http"
	"time"
)

const (
	debugSyncChanged = true
)

type DB struct {
	client     *gtoggl.TogglClient
	httpClient *http.Client
	sub        *gtwebsocket.TogglWebsocketClient
	defaultWId uint64
	s          Store

	ClientsDB
	ProjectsDB
	TagsDB
	TimeEntriesDB
}

func NewDB(token string, store Store) (*DB, error) {
	httpClient := &http.Client{Timeout: 10 * time.Second}
	client, err := gtoggl.NewClient(token, gthttp.SetHttpClient(httpClient))
	if err != nil {
		return nil, err
	}
	sub := gtwebsocket.NewClient(token)
	db := &DB{
		s:          store,
		client:     client,
		httpClient: httpClient,
		sub:        sub,
	}
	db.ClientsDB = ClientsDB{common: db}
	db.ProjectsDB = ProjectsDB{common: db}
	db.TagsDB = TagsDB{common: db}
	db.TimeEntriesDB = TimeEntriesDB{common: db}
	db.s.Migrate()
	return db, nil
}

func (db *DB) store() Store {
	return db.s
}

func (db *DB) subscriber() Subscriber {
	return db.sub
}

func (db *DB) wid() uint64 {
	return db.defaultWId
}

func (db *DB) toggl() *gtoggl.TogglClient {
	return db.client
}

func (db *DB) SyncUser(force bool) error {
	user, err := db.client.UserClient.Get(true)
	if err != nil {
		return err
	}
	db.defaultWId = user.DefaultWId
	db.ClientsDB.saveUserData(user, force)
	db.ProjectsDB.saveUserData(user, force)
	db.TagsDB.saveUserData(user, force)
	db.TimeEntriesDB.saveUserData(user, force)
	return nil
}

func (db *DB) SyncChanges() error {
	log.Println("Start sync changes")
	defer log.Println("Finish sync changes")
	return db.TimeEntriesDB.SyncChanges()
}

func (db *DB) StartListen(ctx context.Context) {
	db.ClientsDB.Subscribe()
	db.ProjectsDB.Subscribe()
	db.TagsDB.Subscribe()
	db.TimeEntriesDB.Subscribe()
	go func() {
		err := db.sub.Listen(ctx)
		if err != nil {
			log.Fatal("WebSocket listen error:", err)
		}
	}()
}
