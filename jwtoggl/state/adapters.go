package state

import "go-jw-toggl/jwtoggl/state/store"

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
