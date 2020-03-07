package history

import (
	"github.com/gotk3/gotk3/gtk"
	"go-jw-toggl/jwtoggl/builder"
	"log"
	"strconv"
)

type History struct {
	store    *gtk.TreeStore
	treeView *gtk.TreeView
}

func (h *History) Fill() {
	for i := 0; i < 5; i++ {
		it := h.store.Append(nil)
		if err := h.store.SetValue(it, 0, strconv.Itoa(i)); err != nil {
			log.Fatal(err)
		}
		if err := h.store.SetValue(it, 1, "entry "+strconv.Itoa(i)); err != nil {
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
