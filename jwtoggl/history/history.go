package history

import (
	"go-jw-toggl/jwtoggl/builder"
	"go-jw-toggl/jwtoggl/gtoggl/gttimentry"
	"go-jw-toggl/jwtoggl/history/gtk_history"
)

type History struct {
	history *gtk_history.GtkHistory
}

func (h *History) SetEntries(entries []*gttimeentry.TimeEntry) {
	h.history.SetEntries(entries)
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
		history: history,
	}
	return h, nil
}
