package gtk_entry_form

import (
	"errors"
	"fmt"
	"github.com/gotk3/gotk3/gtk"
	"go-jw-toggl/jwtoggl/builder"
	"go-jw-toggl/jwtoggl/gtoggl/gtproject"
	"go-jw-toggl/jwtoggl/gtoggl/gttag"
	"go-jw-toggl/jwtoggl/gtoggl/gttimentry"
	"go-jw-toggl/jwtoggl/tools"
	"log"
	"strconv"
	"time"
)

const (
	textDescription = "text_entry"
	textStart       = "entry_start"
	textStop        = "entry_stop"
	textProject     = "text_project_entry"
	textTag         = "text_tag_entry"
	comboProject    = "text_project"
	comboTag        = "text_tag"
	buttonStartStop = "button_start_stop"
	imageStart      = "icon_start"
	imageStop       = "icon_stop"
	storeProjects   = "projects_list_store"
	storeTags       = "tags_list_store"
)

type GtkEntryForm struct {
	ui                tools.UI
	description       *gtk.Entry
	start             *gtk.Entry
	stop              *gtk.Entry
	projectEntry      *gtk.Entry
	tagEntry          *gtk.Entry
	project           *gtk.ComboBox
	tag               *gtk.ComboBox
	button            *gtk.Button
	imageStop         *gtk.Image
	imageStart        *gtk.Image
	projectsListStore *gtk.ListStore
	tagsListStore     *gtk.ListStore
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
	onDescriptionChanged interface{},
	onProjectEntryChanged interface{},
	onTagEntryChanged interface{},
	onEditDone interface{},
) (err error) {

	if _, err = f.button.Connect("clicked", onButtonClicked); err != nil {
		return err
	}
	if _, err = f.description.Connect("changed", onDescriptionChanged); err != nil {
		return err
	}
	if _, err = f.projectEntry.Connect("changed", onProjectEntryChanged); err != nil {
		return err
	}
	if _, err = f.tagEntry.Connect("changed", onTagEntryChanged); err != nil {
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

func (f *GtkEntryForm) SetProjects(projects gtproject.Projects) error {
	return f.ui.CallParamError(f.setProjects, projects)
}

func (f *GtkEntryForm) setProjects(v interface{}) error {
	projects := v.(gtproject.Projects)
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

func (f *GtkEntryForm) SetTags(tags gttag.Tags) error {
	return f.ui.CallParamError(f.setTags, tags)
}

func (f *GtkEntryForm) setTags(v interface{}) error {
	tags := v.(gttag.Tags)
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

func (f *GtkEntryForm) uiGetTimeEntry(eid interface{}) (interface{}, error) {
	var err error
	entry := &gttimeentry.TimeEntry{Id: eid.(uint64)}

	if entry.Description, err = f.description.GetText(); err != nil {
		return nil, err
	}
	if entry.Pid, err = strconv.ParseUint(f.project.GetActiveID(), 10, 64); err != nil {
		return nil, err
	}
	var tag string
	if tag = f.tag.GetActiveID(); tag == "" {
		return nil, errors.New("no tags")
	} else {
		entry.Tags = []string{tag}
	}
	// FIXME: respect timezone!
	now := time.Now()
	if start, err := f.start.GetText(); err != nil {
		return nil, err

	} else if start == "" {
		entry.Start = now
	} else {
		today := now.Format("2006-01-02")
		tz := now.Format("-0700")
		entry.Start, err = time.Parse("2006-01-02 15:04:05", today+" "+start+tz)
		if err != nil {
			return nil, err
		}
	}
	return entry, nil
}

func (f *GtkEntryForm) GetTimeEntry(eid uint64) *gttimeentry.TimeEntry {
	v, err := f.ui.CallParamValueError(f.uiGetTimeEntry, eid)
	if err != nil {
		return nil
	}
	return v.(*gttimeentry.TimeEntry)
}

func (f *GtkEntryForm) uiSetTimeEntry(v interface{}) {
	entry := v.(*gttimeentry.TimeEntry)
	if entry == nil {
		f.description.SetText("")
		f.project.SetActive(-1)
		f.tag.SetActive(-1)
		f.button.SetImage(f.imageStart)
		f.start.SetText("")
		f.stop.SetText("")
		f.projectEntry.SetText("")
		f.tagEntry.SetText("")
		f.button.SetSensitive(false)
	} else {
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
		} else {
			f.stop.SetText(entry.Stop.Format("15:04:05"))
		}
	}
}

func (f *GtkEntryForm) SetTimeEntry(entry *gttimeentry.TimeEntry) {
	log.Println("SetTimeEntry:", entry)
	f.ui.CallParam(f.uiSetTimeEntry, entry)
}

func (f *GtkEntryForm) SetButtonSensitive(s bool) {
	f.ui.CallParam(f.uiSetButtonSensitive, s)
}

func (f *GtkEntryForm) uiSetButtonSensitive(s interface{}) {
	f.button.SetSensitive(s.(bool))
}

func (f *GtkEntryForm) uiSetActiveProject() error {
	if project, err := f.projectEntry.GetText(); err != nil {
		return err
	} else {
		filter := func(model *gtk.TreeModel, path *gtk.TreePath, iter *gtk.TreeIter, userData interface{}) bool {
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
	err := f.ui.CallError(f.uiSetActiveProject)
	if err != nil {
		log.Fatal(err)
	}
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
	err := f.ui.CallError(f.uiSetActiveTag)
	if err != nil {
		log.Fatal(err)
	}
}
