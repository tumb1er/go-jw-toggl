package gtk_history

import (
	"fmt"
	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/gtk"
	"github.com/gotk3/gotk3/pango"
	"go-jw-toggl/jwtoggl/builder"
	"go-jw-toggl/jwtoggl/dto"
	"go-jw-toggl/jwtoggl/tools"
	"log"
	"strconv"
	"time"
)

const (
	popupMenuId        = "history_menu"
	menuContinue       = "menu_continue"
	menuDelete         = "menu_delete"
	columnId           = 0
	columnDescription  = 1
	columnTimings      = 2
	columnGroupCount   = 3
	columnBackground   = 4
	columnGroupXPad    = 5
	columnGroupWidth   = 6
	dayBackground      = "#eeeeee"
	attrMarkup         = "markup"
	attrBackground     = "background"
	attrXPad           = "xpad"
	attrWidth          = "width"
	propXPad           = "xpad"
	propYPad           = "ypad"
	propAlignment      = "alignment"
	propXAlign         = "xalign"
	signalActivate     = "activate"
	signalRowActivated = "row-activated"
	signalButtonPress  = "button-press-event"
)

type Callback func(entryId uint64)

type entriesContext struct {
	Entries  []dto.TimeEntry
	Projects map[uint64]dto.Project
	Clients  map[uint64]dto.Client
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
	entries []dto.TimeEntry,
	projects map[uint64]dto.Project,
	clients map[uint64]dto.Client) {
	tools.Schedule(h.setEntries, entriesContext{
		Entries:  entries,
		Projects: projects,
		Clients:  clients,
	})
}

func (h *GtkHistory) setEntries(c entriesContext) {
	h.store.Clear()

	cache := make(dayGroup)

	for _, entry := range c.Entries {
		duration := entry.GetDuration() // entry.Stop.Sub(entry.Start)
		ds := cache.add(h.store, entry)
		ds.duration += duration
		ts := ds.add(h.store, entry)
		if ts.count == 0 {
			h.insertEntry(c, ts.it, entry, false)
		} else {
			it := h.store.Append(ts.it)
			h.insertEntry(c, it, entry, true)
		}
		ts.count += 1
	}

	for day, ds := range cache {
		h.insertDayEntry(day.Format("2006-01-02"), ds.it, ds.duration)
		for _, ts := range ds.tasks {
			if ts.count > 1 {
				h.insertTaskSummary(ts.it, ts.count)
			}
		}
	}
}

func (h *GtkHistory) insertTaskSummary(it *gtk.TreeIter, count int) {
	value := RenderGroup(count)
	if err := h.store.SetValue(it, columnGroupCount, value); err != nil {
		log.Fatal(err)
	}
	xpad := 0
	width := 0
	if count > 0 {
		xpad = 8
		width = 32
	}
	if err := h.store.SetValue(it, columnGroupXPad, xpad); err != nil {
		log.Fatal(err)
	}
	if err := h.store.SetValue(it, columnGroupWidth, width); err != nil {
		log.Fatal(err)
	}
}

func (h *GtkHistory) insertDayEntry(day string, it *gtk.TreeIter, duration time.Duration) {
	value := RenderDate(day)
	if err := h.store.SetValue(it, columnDescription, value); err != nil {
		log.Fatal(err)
	}
	value = RenderDate(duration.Round(time.Second).String())
	if err := h.store.SetValue(it, columnTimings, value); err != nil {
		log.Fatal(err)
	}
	if err := h.store.SetValue(it, columnBackground, dayBackground); err != nil {
		log.Fatal(err)
	}
	if path, err := h.store.GetPath(it); err != nil {
		log.Fatal(err)
	} else {
		h.treeView.ExpandRow(path, false)
	}
}

func (h *GtkHistory) insertEntry(entries entriesContext, it *gtk.TreeIter,
	entry dto.TimeEntry, grouped bool) {
	var ctx Ctx
	ctx.Description = entry.GetName()
	if project, ok := entries.Projects[entry.GetProjectTogglID()]; ok {
		ctx.Project = project.GetName()
		ctx.Color = project.GetHexColor()
		if client, ok := entries.Clients[project.GetClientTogglID()]; ok {
			ctx.Client = client.GetName()
		}
	}
	ctx.Tag = entry.GetTag()

	start := entry.GetStart()
	if !start.IsZero() {
		ctx.Start = start.Local().Format("15:04")
	}
	stop := entry.GetStop()
	if !stop.IsZero() {
		ctx.Stop = stop.Local().Format("15:04")
	}
	duration := entry.GetDuration()
	if duration > 0 {
		ctx.Duration = duration.Truncate(time.Second).String()
	}

	if err := h.store.SetValue(it, columnId, fmt.Sprintf("%d", entry.GetTogglID())); err != nil {
		log.Fatal(err)
	}
	if err := h.store.SetValue(it, columnDescription, RenderDescription(ctx)); err != nil {
		log.Fatal(err)
	}
	if err := h.store.SetValue(it, columnTimings, RenderTimings(ctx)); err != nil {
		log.Fatal(err)
	}

	width := 0
	if grouped {
		width = 32
	}
	if err := h.store.SetValue(it, columnGroupWidth, width); err != nil {
		log.Fatal(err)
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
		h.treeView.ExpandToPath(path)
	}
}

func (h *GtkHistory) onMenuContinue() {
	if h.onEntryContinue != nil {
		if entry, err := h.getSelectedEntryId(); err == nil {
			h.onEntryContinue(entry)
		}
	}
}

func (h *GtkHistory) onMenuDelete() {
	if h.onEntryDelete != nil {
		if entry, err := h.getSelectedEntryId(); err == nil {
			h.onEntryDelete(entry)
		}
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
		if _, err = menuItem.Connect(signalActivate, h.onMenuContinue); err != nil {
			return nil, err
		}
	}

	if menuItem, err := b.GetMenuItem(menuDelete); err != nil {
		return nil, err
	} else {
		if _, err = menuItem.Connect(signalActivate, h.onMenuDelete); err != nil {
			return nil, err
		}
	}

	if _, err = h.treeView.Connect(signalRowActivated, h.onRowActivated); err != nil {
		return nil, err
	}

	area, err := CellAreaBoxNew()
	if err != nil {
		log.Fatal(err)
	}
	column, err := TreeViewColumnNewWithArea(area)
	if err != nil {
		return nil, err
	} else {
		column.SetTitle("Recent entries")
	}

	widgets := []func(*gtk.TreeViewColumn, *CellAreaBox) error{
		setupGroupCountRenderer,
		setupDescriptionRenderer,
		setupTimingsRenderer,
	}
	for _, widget := range widgets {
		if err := widget(column, area); err != nil {
			return nil, err
		}
	}
	h.treeView.AppendColumn(column)
	h.treeView.SetSearchEqualSubstringMatch()

	if _, err := h.treeView.Connect(signalButtonPress, h.OnButtonClick); err != nil {
		return nil, err
	}
	return h, nil
}

func setupTimingsRenderer(column *gtk.TreeViewColumn, area *CellAreaBox) error {
	if r, err := CellRendererTextNew(); err != nil {
		return err
	} else {
		if err := SetProperties(r.Object, map[string]interface{}{
			propXPad:      8,
			propYPad:      4,
			propAlignment: pango.ALIGN_RIGHT,
			propXAlign:    1.0,
		}); err != nil {
			return err
		}
		area.PackStart(r, true, false, false)
		AddAttributes(column, r, map[string]int{
			attrMarkup:     columnTimings,
			attrBackground: columnBackground,
		})
		return nil
	}
}

func setupGroupCountRenderer(column *gtk.TreeViewColumn, area *CellAreaBox) error {
	if r, err := CellRendererTextNew(); err != nil {
		return err
	} else {
		area.PackStart(r, false, false, false)
		AddAttributes(column, r, map[string]int{
			attrMarkup:     columnGroupCount,
			attrBackground: columnBackground,
			attrXPad:       columnGroupXPad,
			attrWidth:      columnGroupWidth,
		})
		return nil
	}
}

func setupDescriptionRenderer(column *gtk.TreeViewColumn, area *CellAreaBox) error {
	if r, err := CellRendererTextNew(); err != nil {
		return err
	} else {
		if err := r.SetProperty(propYPad, 4); err != nil {
			return err
		}
		area.PackStart(r, true, false, false)
		AddAttributes(column, r, map[string]int{
			attrMarkup:     columnDescription,
			attrBackground: columnBackground,
		})
		return nil
	}
}
