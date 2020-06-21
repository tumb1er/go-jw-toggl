package state

import (
	"go-jw-toggl/jwtoggl/gtoggl/gtclient"
	"go-jw-toggl/jwtoggl/gtoggl/gtproject"
	"go-jw-toggl/jwtoggl/gtoggl/gttag"
	"go-jw-toggl/jwtoggl/gtoggl/gttimeentry"
	"go-jw-toggl/jwtoggl/gtoggl/gtwebsocket"
	"go-jw-toggl/jwtoggl/store"
	"time"
)

//Store предоставляет доступ к ORM-моделям
type Store interface {
	Migrate()
	TimeEntries() Entries
	Tags() Tags
	Projects() Projects
	Clients() Clients
}

//Clients предоставляет доступ ко списку клиентов в БД
type Clients interface {
	List() store.Clients
	Sync(gtclient.Clients, time.Time, uint64) (bool, error)
}

//Projects предоставляет доступ ко списку проектов в БД
type Projects interface {
	List() store.Projects
	Sync(gtproject.Projects, time.Time, uint64) (bool, error)
	Get(id uint64) *store.Project
}

//Tags предоставляет доступ ко списку тегов в БД
type Tags interface {
	List() store.Tags
	Sync(gttag.Tags, time.Time, uint64) (bool, error)
}

//Entries предоставляет доступ ко списку отметок времени в БД
type Entries interface {
	Get(id uint64) *store.TimeEntry
	Create(entry *store.TimeEntry)
	Current() store.TimeEntry
	Update(entry *store.TimeEntry)
	Sync(gttimeentry.TimeEntries, time.Time, uint64) (bool, error)
	List() store.TimeEntries
	Delete(id uint64) bool
}

//common дает миксинам TagsDB, ClientsDB и прочим доступ к необходимым им полям из общего класса.
type common interface {
	SyncUser(bool) error

	store() Store
	subscriber() Subscriber
	wid() uint64
	update(*store.TimeEntry) bool
}

// Subscriber предоставляет доступ к подписке на события WS-клиента Toggl
type Subscriber interface {
	OnTimeEntryAction(callback gtwebsocket.OnTimeEntryActionCallback)
	OnProjectAction(callback gtwebsocket.OnProjectActionCallback)
	OnClientAction(callback gtwebsocket.OnClientActionCallback)
	OnTagAction(callback gtwebsocket.OnTagActionCallback)
}
