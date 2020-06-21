package gtk_entry_form

import (
	"fmt"
	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"
	"go-jw-toggl/jwtoggl/builder"
	"go-jw-toggl/jwtoggl/dto"
	"go-jw-toggl/jwtoggl/tools"
	"log"
	"strconv"
	"time"
)

const (
	textDescription = "text_entry"
	textStart       = "entry_start"
	textStop        = "entry_stop"
	textDuration    = "entry_duration"
	textProject     = "text_project_entry"
	textTag         = "text_tag_entry"
	comboProject    = "text_project"
	comboTag        = "text_tag"
	buttonStartStop = "button_start_stop"
	imageStart      = "icon_start"
	imageStop       = "icon_stop"
	storeProjects   = "projects_list_store"
	storeTags       = "tags_list_store"
	iconWarning     = "gtk-dialog-warning"
)

type GtkEntryForm struct {
	description       *gtk.Entry
	start             *gtk.Entry
	stop              *gtk.Entry
	duration          *gtk.Entry
	projectEntry      *gtk.Entry
	tagEntry          *gtk.Entry
	project           *gtk.ComboBox
	tag               *gtk.ComboBox
	button            *gtk.Button
	imageStop         *gtk.Image
	imageStart        *gtk.Image
	projectsListStore *gtk.ListStore
	tagsListStore     *gtk.ListStore

	handler UIHandler
}

type UIHandler interface {
	OnButtonClicked()
	OnStartTimeChanged(*time.Time) bool
	OnProjectChanged(uint64) bool
	OnTagChanged(string) bool
	OnDescriptionChanged(string) bool
	OnEditCompleted()
}

func MarkValid(entry *gtk.Entry, valid bool) bool {
	var icon string

	text, err := entry.GetText()
	if err != nil {
		log.Fatal(err)
	}
	if !valid && text != "" {
		icon = iconWarning
	}
	if err := entry.SetProperty("secondary-icon-stock", icon); err != nil {
		log.Fatalf("%+v mark valid error: %+v\n", entry, err)
	}
	return valid
}

func New(b *builder.Builder, handler UIHandler) (f *GtkEntryForm, err error) {
	f = &GtkEntryForm{handler: handler}
	if f.description, err = b.GetEntry(textDescription); err != nil {
		return f, err
	}
	if f.start, err = b.GetEntry(textStart); err != nil {
		return f, err
	}
	if f.stop, err = b.GetEntry(textStop); err != nil {
		return f, err
	}
	if f.duration, err = b.GetEntry(textDuration); err != nil {
		return f, err
	}
	if f.projectEntry, err = b.GetEntry(textProject); err != nil {
		return f, err
	}
	if f.tagEntry, err = b.GetEntry(textTag); err != nil {
		return f, err
	}
	if f.project, err = b.GetComboBox(comboProject); err != nil {
		return f, err
	}
	if f.tag, err = b.GetComboBox(comboTag); err != nil {
		return f, err
	}
	if f.button, err = b.GetButton(buttonStartStop); err != nil {
		return f, err
	}
	if f.imageStart, err = b.GetImage(imageStart); err != nil {
		return f, err
	}
	if f.imageStop, err = b.GetImage(imageStop); err != nil {
		return f, err
	}
	if f.projectsListStore, err = b.GetListStore(storeProjects); err != nil {
		return f, err
	}
	if f.tagsListStore, err = b.GetListStore(storeTags); err != nil {
		return f, err
	}
	if r, err := gtk.CellRendererTextNew(); err != nil {
		return f, err
	} else {
		// Удаляем автоматически добавленную колонку, которая появилась из-за EntryTextColumn
		CellLayoutClear(&f.project.CellLayout)
		// Добавляем свою колонку с Markup
		f.project.PackStart(r, true)
		f.project.AddAttribute(r, "markup", 2)
	}
	return f, nil
}

type sender interface {
	Connect(string, interface{}, ...interface{}) (glib.SignalHandle, error)
}

func connectSignals(obj sender, handler interface{}) error {
	for i, signal := range []string{"changed", "focus-out-event"} {
		if _, err := obj.Connect(signal, handler, i == 1); err != nil {
			return err
		}
	}
	return nil
}

func (f *GtkEntryForm) ConnectSignals() error {
	if err := connectSignals(f.description, f.notifyDescription); err != nil {
		return err
	}
	if err := connectSignals(f.projectEntry, f.notifyProject); err != nil {
		return err
	}
	if err := connectSignals(f.project, f.notifyProject); err != nil {
		return err
	}
	if err := connectSignals(f.tagEntry, f.notifyTag); err != nil {
		return err
	}
	if err := connectSignals(f.tag, f.notifyTag); err != nil {
		return err
	}
	if err := connectSignals(f.start, f.notifyStart); err != nil {
		return err
	}
	if _, err := f.button.Connect("clicked", f.notifyClicked); err != nil {
		return err
	}
	return nil
}

func (f *GtkEntryForm) notifyDescription(_ *gtk.Entry, event interface{}) {
	if text, err := f.description.GetText(); err != nil {
		log.Fatal(err)
	} else {
		MarkValid(f.description, f.handler.OnDescriptionChanged(text))
		if _, ok := event.(*gdk.Event); ok {
			f.handler.OnEditCompleted()
		}
	}
}

func (f *GtkEntryForm) notifyProject() {
	f.SetActiveProject()
	projectId := f.project.GetActiveID()
	if projectId == "" {
		MarkValid(f.projectEntry, f.handler.OnProjectChanged(0))
		return
	}
	if pid, err := strconv.ParseUint(projectId, 10, 64); err != nil {
		log.Fatal("projectId parse error:", err)
	} else {

		MarkValid(f.projectEntry, f.handler.OnProjectChanged(pid))
	}
}

func (f *GtkEntryForm) notifyTag() {
	f.SetActiveTag()
	tag := f.tag.GetActiveID()
	MarkValid(f.projectEntry, f.handler.OnTagChanged(tag))
}

func (f *GtkEntryForm) notifyStart() {
	if startTime, err := f.parseStart(); err != nil {
		MarkValid(f.start, f.handler.OnStartTimeChanged(nil))
	} else {
		MarkValid(f.start, f.handler.OnStartTimeChanged(&startTime))
	}
}

func (f *GtkEntryForm) notifyClicked() {
	f.handler.OnButtonClicked()
}

func (f *GtkEntryForm) SetProjects(projects []dto.Project) {
	tools.Schedule(func() {
		if err := f.setProjects(projects); err != nil {
			log.Fatal(err)
		}
	})
}

func (f *GtkEntryForm) setProjects(projects []dto.Project) error {
	ls := f.projectsListStore
	ls.Clear()
	for _, project := range projects {
		it := ls.Append()
		html := fmt.Sprintf("<span foreground=\"%s\">%s</span>",
			project.GetHexColor(), project.GetName())
		if err := ls.SetValue(it, 0, project.GetTogglID()); err != nil {
			return err
		}
		if err := ls.SetValue(it, 1, project.GetName()); err != nil {
			return err
		}
		if err := ls.SetValue(it, 2, html); err != nil {
			return err
		}
	}
	return nil
}

func (f *GtkEntryForm) SetTags(tags []dto.Tag) {
	tools.Schedule(func() {
		if err := f.setTags(tags); err != nil {
			log.Fatal(err)
		}
	})
}

func (f *GtkEntryForm) setTags(tags []dto.Tag) error {
	ls := f.tagsListStore
	ls.Clear()
	for _, tag := range tags {
		it := ls.Append()
		if err := ls.SetValue(it, 0, tag.GetName()); err != nil {
			return err
		}
	}
	return nil
}

func (f *GtkEntryForm) parseStart() (time.Time, error) {
	var start string
	var err error
	if start, err = f.start.GetText(); err != nil {
		log.Fatal(err)
	}
	now := time.Now().Local()
	if start != "" {
		local := fmt.Sprintf("%sT%s%s",
			now.Format("2006-01-02"), start, now.Format("-0700"))
		var parsed time.Time
		if parsed, err = time.Parse("2006-01-02T15:04:05-0700", local); err != nil {
			parsed, err = time.Parse("2006-01-02T15:04-0700", local)
		}
		if err != nil {
			return parsed, err
		} else {
			if now.Before(parsed) {
				parsed = parsed.AddDate(0, 0, -1)
			}
			return parsed, nil
		}
	} else {
		return time.Time{}, nil
	}
}

func (f *GtkEntryForm) setTimeEntry(entry dto.TimeEntry) {
	log.Println("SetTimeEntry:", entry)
	if entry.Nil() {
		f.description.SetText("")
		MarkValid(f.description, true)
		f.project.SetActive(-1)
		f.tag.SetActive(-1)
		f.button.SetImage(f.imageStart)
		f.start.SetText("")
		f.stop.SetText("")
		f.projectEntry.SetText("")
		MarkValid(f.projectEntry, true)
		f.tagEntry.SetText("")
		MarkValid(f.tagEntry, true)
		f.button.SetSensitive(false)
		f.duration.SetText("")
		f.stop.SetText("")
	} else {
		f.description.SetText(entry.GetDescription())
		tag := entry.GetTag()
		if tag != "" {
			f.tag.SetActiveID(tag)
		} else {
			f.tag.SetActive(-1)
		}
		f.project.SetActiveID(strconv.FormatUint(entry.GetProjectTogglID(), 10))
		f.button.SetImage(f.imageStop)
		if entry.GetStart().IsZero() {
			f.start.SetText("")
		} else {
			f.start.SetText(entry.GetStart().Local().Format("15:04:05"))
		}
		if entry.GetStop().IsZero() {
			f.stop.SetText("")
			d := time.Now().Sub(entry.GetStart())
			f.duration.SetText(d.Truncate(time.Second).String())
		} else {
			f.stop.SetText(entry.GetStop().Format("15:04:05"))
			d := entry.GetStop().Sub(entry.GetStart())
			f.duration.SetText(d.Truncate(time.Second).String())
		}
	}
}

func (f *GtkEntryForm) SetTimeEntry(entry dto.TimeEntry) {
	log.Printf("Schedule setTimeEntry %+v\n", entry)
	tools.Schedule(f.setTimeEntry, entry)
}

func (f *GtkEntryForm) SetButtonSensitive(s bool) {
	tools.Schedule(f.button.SetSensitive, s)
}

func (f *GtkEntryForm) uiSetActiveProject() error {
	if project, err := f.projectEntry.GetText(); err != nil {
		return err
	} else {
		filter := func(
			model *gtk.TreeModel,
			path *gtk.TreePath,
			iter *gtk.TreeIter,
			userData interface{},
		) bool {
			name, err := tools.GetString(model, iter, 1)
			if err != nil {
				return true
			}
			if name == project {
				projectId, err := tools.GetString(model, iter, 0)
				if err != nil {
					return true
				}
				f.project.SetActiveID(projectId)
				return true
			}
			return false
		}
		return f.projectsListStore.ForEach(filter)
	}
}

func (f *GtkEntryForm) SetActiveProject() {
	tools.Schedule(func() {
		if f.uiSetActiveProject() != nil {
			log.Println("Project not matched")
		}
	})
}

func (f *GtkEntryForm) uiSetActiveTag() error {
	if tag, err := f.tagEntry.GetText(); err != nil {
		return err
	} else {
		f.tag.SetActiveID(tag)
	}
	return nil
}

func (f *GtkEntryForm) SetActiveTag() {
	tools.Schedule(func() {
		if f.uiSetActiveTag() != nil {
			log.Println("Tag not matched")
		}
	})
}

func (f *GtkEntryForm) UpdateDuration() {
	tools.Schedule(func() {
		startTime, err := f.parseStart()
		if err != nil {
			startTime = time.Time{}
		}
		var duration string
		if !startTime.IsZero() {
			n := time.Now().Local()
			if startTime.Before(n.Add(-time.Second)) {
				d := n.Sub(startTime).Truncate(time.Second)
				duration = d.String()
			}
		}
		f.duration.SetText(duration)
	})
}
