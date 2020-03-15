package gtk_history

import (
	"fmt"
	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/gtk"
	"go-jw-toggl/jwtoggl/builder"
	"go-jw-toggl/jwtoggl/gtoggl/gtclient"
	"go-jw-toggl/jwtoggl/gtoggl/gtproject"
	"go-jw-toggl/jwtoggl/gtoggl/gttimentry"
	"go-jw-toggl/jwtoggl/tools"
	"log"
	"strconv"
	"time"
)

const (
	popupMenuId  = "history_menu"
	menuContinue = "menu_continue"
	menuDelete   = "menu_delete"
)

type Callback func(entryId uint64)

type entriesContext struct {
	Entries  gttimeentry.TimeEntries
	Projects map[uint64]gtproject.Project
	Clients  map[uint64]gtclient.Client
}

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

func (h *GtkHistory) SetEntries(
	entries gttimeentry.TimeEntries,
	projects map[uint64]gtproject.Project,
	clients map[uint64]gtclient.Client) {
	tools.Schedule(h.setEntries, entriesContext{
		Entries:  entries,
		Projects: projects,
		Clients:  clients,
	})
}
func (h *GtkHistory) setEntries(c entriesContext) {
	h.store.Clear()
	for _, entry := range c.Entries {
		it := h.store.Append(nil)
		if err := h.store.SetValue(it, 0, fmt.Sprintf("%d", entry.Id)); err != nil {
			log.Fatal(err)
		}

		var ctx Ctx
		ctx.Description = entry.Description
		if project, ok := c.Projects[entry.Pid]; ok {
			ctx.Project = project.Name
			ctx.Color = project.HexColor
			if client, ok := c.Clients[project.CId]; ok {
				ctx.Client = client.Name
			}
		}
		if len(entry.Tags) >= 1 {
			ctx.Tag = entry.Tags[0]
		}
		if !entry.Start.IsZero() {
			ctx.Start = entry.Start.Format("15:04")
		}
		if !entry.Stop.IsZero() {
			ctx.Stop = entry.Stop.Format("15:04")
		}
		if entry.Duration > 0 {
			duration := time.Duration(entry.Duration) * time.Second
			ctx.Duration = duration.Truncate(time.Second).String()
		}

		if err := h.store.SetValue(it, 1, Render(ctx)); err != nil {
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
	column.AddAttribute(r, "markup", 1)

	if _, err := h.treeView.Connect("button-press-event", h.OnButtonClick); err != nil {
		return nil, err
	}
	return h, nil
}
