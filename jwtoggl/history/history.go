package history

import (
	"fmt"
	"github.com/gotk3/gotk3/gtk"
	"go-jw-toggl/jwtoggl/builder"
	"go-jw-toggl/jwtoggl/gtoggl/gttimentry"
	"go-jw-toggl/jwtoggl/tools"
	"log"
)

type History struct {
	store    *gtk.TreeStore
	treeView *gtk.TreeView
}

func (h *History) SetEntries(entries []*gttimeentry.TimeEntry) {
	tools.Schedule(h.setEntries, entries)
}
func (h *History) setEntries(entries []*gttimeentry.TimeEntry) {
	h.store.Clear()
	for _, entry := range entries {
		it := h.store.Append(nil)
		if err := h.store.SetValue(it, 0, fmt.Sprintf("%d", entry.Id)); err != nil {
			log.Fatal(err)
		}
		if err := h.store.SetValue(it, 1, entry.Description); err != nil {
			log.Fatal(err)
		}
	}
}

func New(b *builder.Builder) *History {
	h := &History{
		store:    b.GetHistoryTreeStore(),
		treeView: b.GetHistoryView(),
	}
	column := h.treeView.GetColumn(0)
	r, err := CellRendererTextNew()
	if err != nil {
		log.Fatal(err)
	}
	column.PackStart(r, true)
	column.AddAttribute(r, "text", 1)
	return h
}
