package gtk_history

import (
	"fmt"
	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/gtk"
	"go-jw-toggl/jwtoggl/builder"
	"go-jw-toggl/jwtoggl/gtoggl/gttimentry"
	"go-jw-toggl/jwtoggl/tools"
	"log"
	"strconv"
)

const (
	popupMenuId  = "history_menu"
	menuContinue = "menu_continue"
	menuDelete   = "menu_delete"
)

type Callback func(entryId uint64)

type GtkHistory struct {
	store           *gtk.TreeStore
	treeView        *gtk.TreeView
	menu            *gtk.Menu
	onEntryDelete   Callback
	onEntryContinue Callback
}

func (h *GtkHistory) OnEntryContinue(cb Callback) {
	h.onEntryContinue = cb
}

func (h *GtkHistory) OnEntryDelete(cb Callback) {
	h.onEntryDelete = cb
}

func (h *GtkHistory) OnButtonClick(_ *gtk.TreeView, e *gdk.Event) {
	b := gdk.EventButtonNewFromEvent(e)
	if b.Button() == 3 {
		log.Println("Right click ")
		h.menu.PopupAtPointer(e)
	} else if b.Button() == 1 && b.Type() == gdk.EVENT_2BUTTON_PRESS {
		log.Println("Left double click ")
	}
}

func (h *GtkHistory) SetEntries(entries gttimeentry.TimeEntries) {
	tools.Schedule(h.setEntries, entries)
}
func (h *GtkHistory) setEntries(entries gttimeentry.TimeEntries) {
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

func (h *GtkHistory) getSelectedEntryId() (uint64, error) {
	s, err := h.treeView.GetSelection()
	if err != nil {
		return 0, err
	}
	defer s.UnselectAll()
	_, it, ok := s.GetSelected()
	if ok {
		v, err := h.store.GetValue(it, 0)
		if err != nil {
			return 0, err
		}
		value, err := v.GoValue()
		if err != nil {
			return 0, err
		}
		if s, ok := value.(string); !ok {
			return 0, fmt.Errorf("type conversion error: %T", value)
		} else {
			if id, err := strconv.ParseUint(s, 10, 64); err != nil {
				return 0, err
			} else {
				return id, nil
			}
		}
	} else {
		return 0, fmt.Errorf("nothing selected")
	}
}

func (h *GtkHistory) onMenuContinue() {
	if h.onEntryContinue != nil {
		entry, err := h.getSelectedEntryId()
		if err != nil {
			log.Fatal(err)
		}
		h.onEntryContinue(entry)
	}
}

func (h *GtkHistory) onMenuDelete() {
	if h.onEntryDelete != nil {
		entry, err := h.getSelectedEntryId()
		if err != nil {
			log.Fatal(err)
		}
		h.onEntryDelete(entry)
	}
}

func New(b *builder.Builder) (*GtkHistory, error) {
	h := &GtkHistory{
		store:    b.GetHistoryTreeStore(),
		treeView: b.GetHistoryView(),
	}
	var err error
	if h.menu, err = b.GetMenu(popupMenuId); err != nil {
		return nil, err
	}
	if menuItem, err := b.GetMenuItem(menuContinue); err != nil {
		return nil, err
	} else {
		if _, err = menuItem.Connect("activate", h.onMenuContinue); err != nil {
			return nil, err
		}
	}

	if menuItem, err := b.GetMenuItem(menuDelete); err != nil {
		return nil, err
	} else {
		if _, err = menuItem.Connect("activate", h.onMenuDelete); err != nil {
			return nil, err
		}
	}

	column := h.treeView.GetColumn(0)
	r, err := gtk.CellRendererTextNew()
	if err != nil {
		return nil, err
	}
	column.PackStart(r, true)
	column.AddAttribute(r, "text", 1)

	if _, err := h.treeView.Connect("button-press-event", h.OnButtonClick); err != nil {
		return nil, err
	}
	return h, nil
}
