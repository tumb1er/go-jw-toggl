package history

import (
	"go-jw-toggl/jwtoggl/builder"
	"go-jw-toggl/jwtoggl/dto"
	"go-jw-toggl/jwtoggl/history/gtk_history"
)

type History struct {
	history  *gtk_history.GtkHistory
	projects map[uint64]dto.Project
	entries  []dto.TimeEntry
}

func (h *History) SetProjects(projects []dto.Project) {
	h.projects = make(map[uint64]dto.Project)
	for _, project := range projects {
		h.projects[project.GetTogglID()] = project
	}
	h.setEntries()
}

func (h *History) setEntries() {
	h.history.SetEntries(h.entries, h.projects)
}

func (h *History) SetEntries(entries []dto.TimeEntry) {
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
		projects: make(map[uint64]dto.Project),
	}
	return h, nil
}
