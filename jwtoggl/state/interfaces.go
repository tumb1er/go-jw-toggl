package state

import (
	"go-jw-toggl/jwtoggl/gtoggl/gtclient"
	"go-jw-toggl/jwtoggl/gtoggl/gtproject"
	"go-jw-toggl/jwtoggl/gtoggl/gttag"
	"go-jw-toggl/jwtoggl/gtoggl/gttimeentry"
	"go-jw-toggl/jwtoggl/gtoggl/gtwebsocket"
	"go-jw-toggl/jwtoggl/store"
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
	Sync(gtclient.Clients, uint64) (bool, error)
}

//Projects предоставляет доступ ко списку проектов в БД
type Projects interface {
	List() store.Projects
	Sync(gtproject.Projects, uint64) (bool, error)
	Get(id uint64) *store.Project
	GetByPK(pk uint) *store.Project
}

//TagSet предоставляет доступ ко списку тегов в БД
type Tags interface {
	List() store.Tags
	Sync(gttag.Tags, uint64) (bool, error)
}

//Entries предоставляет доступ ко списку отметок времени в БД
type Entries interface {
	Get(id uint64) *store.TimeEntry
	Create(entry *store.TimeEntry)
	Current() store.TimeEntry
	Update(entry *store.TimeEntry)
	Sync(gttimeentry.TimeEntries, uint64) (bool, error)
	List() store.TimeEntries
	Delete(pk uint) bool
	MarkDeleted(pk uint) bool
	FromTimeEntry(entry gttimeentry.TimeEntry) store.TimeEntry
	GetByPK(pk uint) *store.TimeEntry
	DeletedEntries() store.TimeEntries
	NewEntries() store.TimeEntries
	ChangedEntries() store.TimeEntries
	CreateOrUpdate(entry *store.TimeEntry) bool
}

//common дает миксинам TagsDB, ClientsDB и прочим доступ к необходимым им полям из общего класса.
type common interface {
	SyncUser(bool) error

	store() Store
	subscriber() Subscriber
	wid() uint64
}

// Subscriber предоставляет доступ к подписке на события WS-клиента Toggl
type Subscriber interface {
	OnTimeEntryAction(callback gtwebsocket.OnTimeEntryActionCallback)
	OnProjectAction(callback gtwebsocket.OnProjectActionCallback)
	OnClientAction(callback gtwebsocket.OnClientActionCallback)
	OnTagAction(callback gtwebsocket.OnTagActionCallback)
}
