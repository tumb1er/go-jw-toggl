package entry_form

import "C"
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

type EntryForm struct {
	ui                tools.UI
	eid               uint64
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
	onNewEntryCreated NewEntryCreatedCallback
	onEntryStopped    EntryStoppedCallback
	mainThread        []byte
}

type NewEntryCreatedCallback func(entry *gttimeentry.TimeEntry)
type EntryStoppedCallback func(entry *gttimeentry.TimeEntry)

func NewEntryForm(b *builder.Builder) (*EntryForm, error) {
	f := &EntryForm{}
	var err error
	if f.description, err = b.GetEntry(textDescription); err != nil {
		return nil, err
	}
	if f.start, err = b.GetEntry(textStart); err != nil {
		return nil, err
	}
	if f.stop, err = b.GetEntry(textStop); err != nil {
		return nil, err
	}
	if f.projectEntry, err = b.GetEntry(textProject); err != nil {
		return nil, err
	}
	if f.tagEntry, err = b.GetEntry(textTag); err != nil {
		return nil, err
	}
	if f.project, err = b.GetComboBox(comboProject); err != nil {
		return nil, err
	}
	if f.tag, err = b.GetComboBox(comboTag); err != nil {
		return nil, err
	}
	if f.button, err = b.GetButton(buttonStartStop); err != nil {
		return nil, err
	}
	if f.imageStart, err = b.GetImage(imageStart); err != nil {
		return nil, err
	}
	if f.imageStop, err = b.GetImage(imageStop); err != nil {
		return nil, err
	}
	if f.projectsListStore, err = b.GetListStore(storeProjects); err != nil {
		return nil, err
	}
	if f.tagsListStore, err = b.GetListStore(storeTags); err != nil {
		return nil, err
	}
	if r, err := gtk.CellRendererTextNew(); err != nil {
		return nil, err
	} else {
		// Удаляем автоматически добавленную колонку, которая появилась из-за EntryTextColumn
		CellLayoutClear(&f.project.CellLayout)
		// Добавляем свою колонку с Markup
		f.project.PackStart(r, true)
		f.project.AddAttribute(r, "markup", 2)
	}
	if _, err = f.button.Connect("clicked", f.onButtonClicked); err != nil {
		return nil, err
	}
	if _, err = f.description.Connect("changed", f.onDescriptionChanged); err != nil {
		return nil, err
	}
	if _, err = f.projectEntry.Connect("changed", f.SetActiveProject); err != nil {
		return nil, err
	}
	if _, err = f.tagEntry.Connect("changed", f.SetActiveTag); err != nil {
		return nil, err
	}
	return f, nil
}

func (f *EntryForm) SetProjects(projects gtproject.Projects) error {
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

func (f *EntryForm) SetTags(tags gttag.Tags) error {
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

func (f *EntryForm) uiGetTimeEntry() (interface{}, error) {
	var err error
	entry := &gttimeentry.TimeEntry{Id: f.eid}

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
	return entry, nil
}

func (f *EntryForm) GetTimeEntry() *gttimeentry.TimeEntry {
	v, err := f.ui.CallValueError(f.uiGetTimeEntry)
	if err != nil {
		return nil
	}
	return v.(*gttimeentry.TimeEntry)
}

func (f *EntryForm) uiSetTimeEntry(v interface{}) {
	entry := v.(*gttimeentry.TimeEntry)
	if entry == nil {
		f.eid = 0
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
		f.eid = entry.Id
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

func (f *EntryForm) SetTimeEntry(entry *gttimeentry.TimeEntry) {
	log.Println("SetTimeEntry:", entry)
	f.ui.CallParam(f.uiSetTimeEntry, entry)
	f.Validate()
}

func (f *EntryForm) uiValidate() (interface{}, error) {
	entry, err := f.uiGetTimeEntry()
	if err != nil {
		return nil, err
	}
	return entry != nil, nil
}

func (f *EntryForm) Validate() bool {
	v, err := f.ui.CallValueError(f.uiValidate)
	if err != nil {
		return false
	}
	valid := v.(bool)
	log.Println("validate:", valid)
	f.button.SetSensitive(valid)
	return valid
}

func (f *EntryForm) uiSetActiveProject() error {
	if project, err := f.projectEntry.GetText(); err != nil {
		return err
	} else {
		getString := func(model *gtk.TreeModel, iter *gtk.TreeIter, column int) (string, error) {
			glibValue, err := model.GetValue(iter, column)
			if err != nil {
				return "", err
			}
			v, err := glibValue.GoValue()
			if err != nil {
				return "", err
			}
			result, ok := v.(string)
			if !ok {
				return "", fmt.Errorf("string convertion error: %T", v)
			}
			return result, nil
		}
		filter := func(model *gtk.TreeModel, path *gtk.TreePath, iter *gtk.TreeIter, userData interface{}) bool {
			name, err := getString(model, iter, 1)
			if err != nil {
				return true
			}
			if name == project {
				projectId, err := getString(model, iter, 0)
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

func (f *EntryForm) SetActiveProject() {
	err := f.ui.CallError(f.uiSetActiveProject)
	if err != nil {
		log.Fatal(err)
	}
	f.Validate()
}

func (f *EntryForm) uiSetActiveTag() error {
	if tag, err := f.tagEntry.GetText(); err != nil {
		return err
	} else {
		f.tag.SetActiveID(tag)
	}
	return nil
}

func (f *EntryForm) SetActiveTag() {
	err := f.ui.CallError(f.uiSetActiveTag)
	if err != nil {
		log.Fatal(err)
	}
	f.Validate()
}

func (f *EntryForm) OnNewEntryCreated(cb NewEntryCreatedCallback) {
	f.onNewEntryCreated = cb
}
func (f *EntryForm) OnEntryStopped(cb EntryStoppedCallback) {
	f.onEntryStopped = cb
}
func (f *EntryForm) onDescriptionChanged() {
	go f.Validate()
}

func (f *EntryForm) onButtonClicked() {
	if f.eid != 0 {
		f.handleStop()
	} else {
		f.handleStart()
	}
}

func (f *EntryForm) handleStart() {
	log.Println("handleStart")
	if !f.Validate() {
		return
	}
	if f.onNewEntryCreated != nil {
		f.onNewEntryCreated(f.GetTimeEntry())
	}
}

func (f *EntryForm) handleStop() {
	log.Println("handleStop")
	entry := f.GetTimeEntry()
	if entry != nil && f.onEntryStopped != nil {
		f.onEntryStopped(entry)
		f.SetTimeEntry(nil)
	}
}
