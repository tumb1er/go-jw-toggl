package gtk_entry_form

import (
	"fmt"
	"github.com/gotk3/gotk3/gtk"
	"go-jw-toggl/jwtoggl/builder"
	"go-jw-toggl/jwtoggl/gtoggl/gtproject"
	"go-jw-toggl/jwtoggl/gtoggl/gttag"
	"go-jw-toggl/jwtoggl/gtoggl/gttimeentry"
	"go-jw-toggl/jwtoggl/tools"
	"log"
	"strconv"
	"strings"
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

	entryId uint64
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

func New(b *builder.Builder) (f *GtkEntryForm, err error) {
	f = &GtkEntryForm{}
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

func (f *GtkEntryForm) ConnectSignals(
	onButtonClicked interface{},
	onProjectEntryChanged interface{},
	onTagEntryChanged interface{},
	onEditDone interface{},
) (err error) {

	if _, err = f.button.Connect("clicked", onButtonClicked); err != nil {
		return err
	}
	if _, err = f.start.Connect("changed", onEditDone); err != nil {
		return err
	}
	if _, err = f.description.Connect("changed", onEditDone); err != nil {
		return err
	}
	if _, err = f.projectEntry.Connect("changed", onProjectEntryChanged); err != nil {
		return err
	}
	if _, err = f.tagEntry.Connect("changed", onTagEntryChanged); err != nil {
		return err
	}
	if _, err = f.start.Connect("focus-out-event", onEditDone); err != nil {
		return err
	}
	if _, err = f.description.Connect("focus-out-event", onEditDone); err != nil {
		return err
	}
	if _, err = f.projectEntry.Connect("focus-out-event", onEditDone); err != nil {
		return err
	}
	if _, err = f.tagEntry.Connect("focus-out-event", onEditDone); err != nil {
		return err
	}
	return nil
}

func (f *GtkEntryForm) SetProjects(projects gtproject.Projects) {
	tools.Schedule(func() {
		if err := f.setProjects(projects); err != nil {
			log.Fatal(err)
		}
	})
}

func (f *GtkEntryForm) setProjects(projects gtproject.Projects) error {
	ls := f.projectsListStore
	ls.Clear()
	for _, project := range projects {
		it := ls.Append()
		html := fmt.Sprintf("<span foreground=\"%s\">%s</span>", project.HexColor, project.Name)
		if err := ls.SetValue(it, 0, project.Id); err != nil {
			return err
		}
		if err := ls.SetValue(it, 1, project.Name); err != nil {
			return err
		}
		if err := ls.SetValue(it, 2, html); err != nil {
			return err
		}
	}
	return nil
}

func (f *GtkEntryForm) SetTags(tags gttag.Tags) {
	tools.Schedule(func() {
		if err := f.setTags(tags); err != nil {
			log.Fatal(err)
		}
	})
}

func (f *GtkEntryForm) setTags(tags gttag.Tags) error {
	ls := f.tagsListStore
	ls.Clear()
	for _, tag := range tags {
		it := ls.Append()
		if err := ls.SetValue(it, 0, tag.Name); err != nil {
			return err
		}
	}
	return nil
}

func (f *GtkEntryForm) getUITimeEntry() *gttimeentry.TimeEntry {
	var err error
	entry := &gttimeentry.TimeEntry{Id: f.entryId}

	if entry.Description, err = f.description.GetText(); err != nil {
		log.Fatal(err)
	}
	valid := true
	valid = MarkValid(f.description, strings.HasPrefix(entry.Description, "#")) && valid

	projectId := f.project.GetActiveID()
	if projectId != "" {
		if entry.Pid, err = strconv.ParseUint(projectId, 10, 64); err != nil {
			log.Fatal("projectId parse error:", err)
		}
	}

	valid = MarkValid(f.projectEntry, entry.Pid != 0) && valid

	var tag string
	if tag = f.tag.GetActiveID(); tag != "" {
		entry.Tags = []string{tag}
	}

	valid = MarkValid(f.tagEntry, len(entry.Tags) != 0) && valid

	startTime, err := f.parseStart()
	if err == nil && startTime.IsZero() {
		startTime = time.Now().Local()
	}
	if err != nil {
		startTime = time.Time{}
	}
	valid = MarkValid(f.start, !startTime.IsZero()) && valid
	entry.Start = startTime

	if !startTime.IsZero() {
		entry.Duration = -entry.Start.Unix()
	} else {
		entry.Duration = 0
	}
	if valid {
		return entry
	} else {
		return nil
	}
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

func (f *GtkEntryForm) uiSetTimeEntry(entry *gttimeentry.TimeEntry) {
	log.Println("SetTimeEntry:", entry)
	if entry == nil {
		f.entryId = 0
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
		f.entryId = entry.Id
		f.description.SetText(entry.Description)
		if len(entry.Tags) > 0 {
			f.tag.SetActiveID(entry.Tags[0])
		} else {
			f.tag.SetActive(-1)
		}
		f.project.SetActiveID(strconv.FormatUint(entry.Pid, 10))
		f.button.SetImage(f.imageStop)
		if entry.Start.IsZero() {
			f.start.SetText("")
		} else {
			f.start.SetText(entry.Start.Local().Format("15:04:05"))
		}
		if entry.Stop.IsZero() {
			f.stop.SetText("")
			d := time.Now().Sub(entry.Start)
			f.duration.SetText(d.Truncate(time.Second).String())
		} else {
			f.stop.SetText(entry.Stop.Format("15:04:05"))
			d := entry.Stop.Sub(entry.Start)
			f.duration.SetText(d.Truncate(time.Second).String())
		}
	}
}

func (f *GtkEntryForm) SetTimeEntry(entry *gttimeentry.TimeEntry) {
	tools.Schedule(f.uiSetTimeEntry, entry)
}

func (f *GtkEntryForm) GetTimeEntry(cb func(entry *gttimeentry.TimeEntry)) {
	tools.Schedule(func() {
		cb(f.getUITimeEntry())
	})
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
