package db

import (
	"go-jw-toggl/jwtoggl/db/store"
	"go-jw-toggl/jwtoggl/gtoggl/gtclient"
	"go-jw-toggl/jwtoggl/gtoggl/gtproject"
	"go-jw-toggl/jwtoggl/gtoggl/gttag"
	"go-jw-toggl/jwtoggl/gtoggl/gttimeentry"
	"go-jw-toggl/jwtoggl/gtoggl/gtwebsocket"
	"time"
)

type Store interface {
	Migrate()
	TimeEntries() Entries
	Tags() Tags
	Projects() Projects
	Clients() Clients
}

type Clients interface {
	List() gtclient.Clients
	Sync(gtclient.Clients, time.Time, uint64) (bool, error)
}

type Projects interface {
	List() gtproject.Projects
	Sync(gtproject.Projects, time.Time, uint64) (bool, error)
}

type Tags interface {
	List() gttag.Tags
	Sync(gttag.Tags, time.Time, uint64) (bool, error)
}

type Entries interface {
	Get(id uint64) *gttimeentry.TimeEntry
	Create(entry gttimeentry.TimeEntry)
	Current() *gttimeentry.TimeEntry
	Update(entry gttimeentry.TimeEntry)
	Sync(gttimeentry.TimeEntries, time.Time, uint64) (bool, error)
	List() gttimeentry.TimeEntries
	Delete(id uint64) bool
}

//StoreAdapter changes signatures of methods returning structs to methods returning interfaces.
type StoreAdapter struct {
	*store.Store
}

func (w *StoreAdapter) Clients() Clients {
	return w.Store.Clients()
}

func (w *StoreAdapter) TimeEntries() Entries {
	return w.Store.Entries()
}

func (w *StoreAdapter) Projects() Projects {
	return w.Store.Projects()
}

func (w *StoreAdapter) Tags() Tags {
	return w.Store.Tags()
}

//CommonDB дает миксинам TagsDB, ClientsDB и прочим доступ к необходимым им полям из общего класса.
type CommonDB interface {
	Store() Store
	Sub() Subscriber
	SyncUser(bool) error
	WId() uint64
	update(gttimeentry.TimeEntry) bool
}

// Subscriber предоставляет доступ к подписке на события WS-клиента Toggl
type Subscriber interface {
	OnTimeEntryAction(callback gtwebsocket.OnTimeEntryActionCallback)
	OnProjectAction(callback gtwebsocket.OnProjectActionCallback)
	OnClientAction(callback gtwebsocket.OnClientActionCallback)
	OnTagAction(callback gtwebsocket.OnTagActionCallback)
}
