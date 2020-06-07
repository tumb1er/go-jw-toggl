package history

import (
	"go-jw-toggl/jwtoggl/builder"
	"go-jw-toggl/jwtoggl/gtoggl/gtclient"
	"go-jw-toggl/jwtoggl/gtoggl/gtproject"
	"go-jw-toggl/jwtoggl/gtoggl/gttimeentry"
	"go-jw-toggl/jwtoggl/history/gtk_history"
)

type History struct {
	history  *gtk_history.GtkHistory
	projects map[uint64]gtproject.Project
	clients  map[uint64]gtclient.Client
	entries  gttimeentry.TimeEntries
}

func (h *History) SetProjects(projects gtproject.Projects) {
	h.projects = make(map[uint64]gtproject.Project)
	for _, project := range projects {
		h.projects[project.Id] = project
	}
	h.setEntries()
}

func (h *History) SetClients(clients gtclient.Clients) {
	h.clients = make(map[uint64]gtclient.Client)
	for _, client := range clients {
		h.clients[client.Id] = client
	}
	h.setEntries()
}

func (h *History) setEntries() {
	h.history.SetEntries(h.entries, h.projects, h.clients)
}

func (h *History) SetEntries(entries gttimeentry.TimeEntries) {
	h.entries = entries
	h.setEntries()
}

func (h *History) InitCallbacks(onEntryContinue, onEntryDelete gtk_history.Callback) {
	h.history.OnEntryContinue(onEntryContinue)
	h.history.OnEntryDelete(onEntryDelete)
}

func New(b *builder.Builder) (*History, error) {
	history, err := gtk_history.New(b)
	if err != nil {
		return nil, err
	}
	h := &History{
		history:  history,
		projects: make(map[uint64]gtproject.Project),
		clients:  make(map[uint64]gtclient.Client),
	}
	return h, nil
}
