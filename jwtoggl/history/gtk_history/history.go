package gtk_history

import (
	"fmt"
	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/gtk"
	"github.com/gotk3/gotk3/pango"
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
	popupMenuId       = "history_menu"
	menuContinue      = "menu_continue"
	menuDelete        = "menu_delete"
	columnId          = 0
	columnDescription = 1
	columnTimings     = 2
	columnGroupCount  = 3
	columnBackground  = 4
	dayBackground     = "#eeeeee"
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

	type parent struct {
		it *gtk.TreeIter
		c  int
	}
	taskNodes := make(map[string]*parent)
	type dayNode struct {
		it       *gtk.TreeIter
		day      string
		duration time.Duration
	}
	dayNodes := make(map[string]*dayNode)
	for _, entry := range c.Entries {
		day := entry.Start.Local().Format("2006-01-02")
		var dayIt *gtk.TreeIter
		if d, ok := dayNodes[day]; ok {
			dayIt = d.it
			d.duration += entry.Stop.Sub(entry.Start)
			value := RenderDate(d.duration.String())
			if err := h.store.SetValue(dayIt, columnTimings, value); err != nil {
				log.Fatal(err)
			}
		} else {
			dayIt = h.store.Append(nil)
			d = &dayNode{
				it:       dayIt,
				day:      day,
				duration: entry.Stop.Sub(entry.Start),
			}
			dayNodes[day] = d
			value := RenderDate(day)
			if err := h.store.SetValue(dayIt, columnDescription, value); err != nil {
				log.Fatal(err)
			}
			value = RenderDate(d.duration.String())
			if err := h.store.SetValue(dayIt, columnTimings, value); err != nil {
				log.Fatal(err)
			}
			if err := h.store.SetValue(dayIt, columnBackground, dayBackground); err != nil {
				log.Fatal(err)
			}
		}
		key := tools.TaskKey(entry)
		var it *gtk.TreeIter
		if p, ok := taskNodes[key]; ok {
			it = h.store.Append(p.it)
			p.c++
			value := fmt.Sprintf("%dx", p.c)
			if err := h.store.SetValue(p.it, columnGroupCount, value); err != nil {
				log.Fatal(err)
			}
		} else {
			it = h.store.Append(dayIt)
			taskNodes[key] = &parent{
				it: it,
				c:  1,
			}
		}
		if err := h.store.SetValue(it, columnId, fmt.Sprintf("%d", entry.Id)); err != nil {
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
			ctx.Start = entry.Start.Local().Format("15:04")
		}
		if !entry.Stop.IsZero() {
			ctx.Stop = entry.Stop.Local().Format("15:04")
		}
		if entry.Duration > 0 {
			duration := time.Duration(entry.Duration) * time.Second
			ctx.Duration = duration.Truncate(time.Second).String()
		}

		if err := h.store.SetValue(it, columnDescription, RenderDescription(ctx)); err != nil {
			log.Fatal(err)
		}
		if err := h.store.SetValue(it, columnTimings, RenderTimings(ctx)); err != nil {
			log.Fatal(err)
		}
	}

	for _, d := range dayNodes {
		if path, err := h.store.GetPath(d.it); err != nil {
			log.Fatal(err)
		} else {
			h.treeView.ExpandRow(path, false)
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

func (h *GtkHistory) onRowActivated(_ *gtk.TreeView, path *gtk.TreePath) {
	// Entry key press - expand/collapse
	if h.treeView.RowExpanded(path) {
		h.treeView.CollapseRow(path)
	} else {
		//h.treeView.CollapseAll()
		h.treeView.ExpandToPath(path)
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

	if _, err = h.treeView.Connect("row-activated", h.onRowActivated); err != nil {
		return nil, err
	}

	column, err := gtk.TreeViewColumnNew()
	if err != nil {
		return nil, err
	} else {
		column.SetTitle("Recent entries")
	}
	if err := setupGroupCountRenderer(column); err != nil {
		return nil, err
	}
	if err := setupDescriptionRenderer(column); err != nil {
		return nil, err
	}
	if err := setupTimingsRenderer(column); err != nil {
		return nil, err
	}
	h.treeView.AppendColumn(column)
	h.treeView.SetSearchEqualSubstringMatch()

	if _, err := h.treeView.Connect("button-press-event", h.OnButtonClick); err != nil {
		return nil, err
	}
	return h, nil
}

func setupTimingsRenderer(column *gtk.TreeViewColumn) error {
	if r, err := gtk.CellRendererTextNew(); err != nil {
		return err
	} else {
		if err := r.SetProperty("ypad", 4); err != nil {
			return err
		}
		if err := r.SetProperty("xpad", 8); err != nil {
			return err
		}
		if err := r.SetProperty("alignment", pango.ALIGN_RIGHT); err != nil {
			return err
		}
		if err := r.SetProperty("xalign", 1.0); err != nil {
			return err
		}
		column.PackStart(r, true)
		column.AddAttribute(r, "markup", columnTimings)
		column.AddAttribute(r, "background", columnBackground)
		return nil
	}
}

func setupGroupCountRenderer(column *gtk.TreeViewColumn) error {
	if r, err := gtk.CellRendererTextNew(); err != nil {
		return err
	} else {
		column.PackStart(r, false)
		column.AddAttribute(r, "markup", columnGroupCount)
		column.AddAttribute(r, "background", columnBackground)
		return nil
	}
}

func setupDescriptionRenderer(column *gtk.TreeViewColumn) error {
	if r, err := gtk.CellRendererTextNew(); err != nil {
		return err
	} else {
		if err := r.SetProperty("ypad", 4); err != nil {
			return err
		}
		column.PackStart(r, true)
		column.AddAttribute(r, "markup", columnDescription)
		column.AddAttribute(r, "background", columnBackground)
		return nil
	}
}
