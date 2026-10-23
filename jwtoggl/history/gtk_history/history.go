package gtk_history

import (
	"fmt"
	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"
	"github.com/gotk3/gotk3/pango"
	"go-jw-toggl/jwtoggl/builder"
	"go-jw-toggl/jwtoggl/dto"
	"go-jw-toggl/jwtoggl/gtoggl"
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
	signalStyleUpdated = "style-updated"
	dayBackgroundBlend = 0.08
)

type Callback func(entryPK uint)

type GtkHistory struct {
	store           *gtk.TreeStore
	treeView        *gtk.TreeView
	menu            *gtk.Menu
	onEntryDelete   Callback
	onEntryContinue Callback
	dayBackground   string
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

func (h *GtkHistory) SetEntries(entries []dto.TimeEntry) {
	tools.Schedule(
		func() {
			h.store.Clear()

			cache := make(dayGroup)

			for _, entry := range entries {
				duration := entry.GetDuration() // entry.Stop.Sub(entry.Start)
				ds := cache.add(h.store, entry)
				ds.duration += duration
				ts := ds.add(h.store, entry)
				if ts.count == 0 {
					h.insertEntry(ts.it, entry, false)
				} else {
					it := h.store.Append(ts.it)
					h.insertEntry(it, entry, true)
				}
				ts.count += 1
			}

			for day, ds := range cache {
				h.insertDayEntry(day, ds.it, ds.duration)
				for _, ts := range ds.tasks {
					if ts.count > 1 {
						h.insertTaskSummary(ts.it, ts.count)
					}
				}
			}
		})
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
	if h.dayBackground != "" {
		if err := h.store.SetValue(it, columnBackground, h.dayBackground); err != nil {
			log.Fatal(err)
		}
	}
	if path, err := h.store.GetPath(it); err != nil {
		log.Fatal(err)
	} else {
		h.treeView.ExpandRow(path, false)
	}
}

func (h *GtkHistory) insertEntry(it *gtk.TreeIter, entry dto.TimeEntry, grouped bool) {
	var ctx Ctx
	ctx.Description = entry.GetName()
	project := entry.GetProject()
	ctx.Project = project.GetName()
	ctx.Color = project.GetHexColor()
	ctx.Client = project.GetClient().GetName()
	ctx.Tag = gtoggl.Tag(entry.GetTags())

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

	if err := h.store.SetValue(it, columnId, fmt.Sprintf("%d", entry.GetPK())); err != nil {
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

func (h *GtkHistory) getSelectedEntryPK() (uint, error) {
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
				return uint(id), nil
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
		if entry, err := h.getSelectedEntryPK(); err == nil {
			h.onEntryContinue(entry)
		}
	}
}

func (h *GtkHistory) onMenuDelete() {
	if h.onEntryDelete != nil {
		if entry, err := h.getSelectedEntryPK(); err == nil {
			h.onEntryDelete(entry)
		}
	}
}

func New(b *builder.Builder) (*GtkHistory, error) {
	h := &GtkHistory{
		store:    b.GetHistoryTreeStore(),
		treeView: b.GetHistoryView(),
	}
	h.dayBackground = treeViewDayBackground(h.treeView)
	h.treeView.ConnectAfter(signalStyleUpdated, func(_ *gtk.TreeView) {
		h.refreshDayBackground()
	})
	var err error
	if h.menu, err = b.GetMenu(popupMenuId); err != nil {
		return nil, err
	}
	if menuItem, err := b.GetMenuItem(menuContinue); err != nil {
		return nil, err
	} else {
		menuItem.Connect(signalActivate, h.onMenuContinue)
	}

	if menuItem, err := b.GetMenuItem(menuDelete); err != nil {
		return nil, err
	} else {
		menuItem.Connect(signalActivate, h.onMenuDelete)
	}

	h.treeView.Connect(signalRowActivated, h.onRowActivated)

	area, err := gtk.CellAreaBoxNew()
	if err != nil {
		log.Fatal(err)
	}
	column, err := gtk.TreeViewColumnNewWithArea(area)
	if err != nil {
		return nil, err
	} else {
		column.SetTitle("Recent entries")
	}

	widgets := []func(*gtk.TreeViewColumn, *gtk.CellAreaBox) error{
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

	h.treeView.Connect(signalButtonPress, h.OnButtonClick)
	return h, nil
}

func setupTimingsRenderer(column *gtk.TreeViewColumn, area *gtk.CellAreaBox) error {
	if r, err := gtk.CellRendererTextNew(); err != nil {
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

func setupGroupCountRenderer(column *gtk.TreeViewColumn, area *gtk.CellAreaBox) error {
	if r, err := gtk.CellRendererTextNew(); err != nil {
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

func setupDescriptionRenderer(column *gtk.TreeViewColumn, area *gtk.CellAreaBox) error {
	if r, err := gtk.CellRendererTextNew(); err != nil {
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

func dayBackgroundFromTheme(lookup func(string) (string, bool)) string {
	background, backgroundOK := lookup("theme_base_color")
	foreground, foregroundOK := lookup("theme_fg_color")
	if backgroundOK && foregroundOK {
		return blendThemeColors(background, foreground, dayBackgroundBlend)
	}
	return ""
}

func blendThemeColors(background, foreground string, ratio float64) string {
	backgroundRed, backgroundGreen, backgroundBlue, backgroundOK := parseCellRendererColor(background)
	foregroundRed, foregroundGreen, foregroundBlue, foregroundOK := parseCellRendererColor(foreground)
	if !backgroundOK || !foregroundOK {
		return ""
	}
	blend := func(background, foreground int) int {
		return int(float64(background)*(1-ratio) + float64(foreground)*ratio + 0.5)
	}
	return fmt.Sprintf("#%02x%02x%02x",
		blend(backgroundRed, foregroundRed),
		blend(backgroundGreen, foregroundGreen),
		blend(backgroundBlue, foregroundBlue))
}

func parseCellRendererColor(value string) (int, int, int, bool) {
	if len(value) != 7 || value[0] != '#' {
		return 0, 0, 0, false
	}
	channels := [3]int{}
	for index := range channels {
		channel, err := strconv.ParseUint(value[1+index*2:3+index*2], 16, 8)
		if err != nil {
			return 0, 0, 0, false
		}
		channels[index] = int(channel)
	}
	return channels[0], channels[1], channels[2], true
}

func treeViewDayBackground(treeView *gtk.TreeView) string {
	style, err := treeView.GetStyleContext()
	if err != nil {
		return ""
	}
	return dayBackgroundFromTheme(func(name string) (string, bool) {
		color, ok := style.LookupColor(name)
		if !ok {
			return "", false
		}
		return cellRendererColor(color), true
	})
}

func (h *GtkHistory) refreshDayBackground() {
	background := treeViewDayBackground(h.treeView)
	if background == h.dayBackground {
		return
	}
	h.dayBackground = background
	iter, ok := h.store.GetIterFirst()
	for ok {
		if err := h.store.SetValue(iter, columnBackground, background); err != nil {
			log.Fatal(err)
		}
		ok = h.store.IterNext(iter)
	}
}

func cellRendererColor(color *gdk.RGBA) string {
	return fmt.Sprintf("#%02x%02x%02x",
		int(color.GetRed()*255+0.5),
		int(color.GetGreen()*255+0.5),
		int(color.GetBlue()*255+0.5))
}

func SetProperties(obj *glib.Object, values map[string]interface{}) error {
	for name, value := range values {
		if err := obj.SetProperty(name, value); err != nil {
			return err
		}
	}
	return nil
}

func AddAttributes(column *gtk.TreeViewColumn, r *gtk.CellRendererText, attributes map[string]int) {
	renderer := gtk.CellRenderer{InitiallyUnowned: r.InitiallyUnowned}
	for name, columnId := range attributes {
		column.AddAttribute(&renderer, name, columnId)
	}
}
