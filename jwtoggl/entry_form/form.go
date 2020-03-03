package entry_form

import "C"
import (
	"fmt"
	"github.com/gotk3/gotk3/gtk"
	"go-jw-toggl/jwtoggl/builder"
	"go-jw-toggl/jwtoggl/gtoggl/gtproject"
	"go-jw-toggl/jwtoggl/gtoggl/gttag"
	"go-jw-toggl/jwtoggl/gtoggl/gttimentry"
	"log"
	"strconv"
)

const (
	textDescription = "text_entry"
	textStart       = "entry_start"
	textStop        = "entry_stop"
	comboProject    = "text_project"
	comboTag        = "text_tag"
	buttonStartStop = "button_start_stop"
	imageStart      = "icon_start"
	imageStop       = "icon_stop"
	storeProjects   = "projects_list_store"
	storeTags       = "tags_list_store"
)

type EntryForm struct {
	description       *gtk.Entry
	start             *gtk.Entry
	stop              *gtk.Entry
	project           *gtk.ComboBox
	tag               *gtk.ComboBox
	button            *gtk.Button
	imageStop         *gtk.Image
	imageStart        *gtk.Image
	projectsListStore *gtk.ListStore
	tagsListStore     *gtk.ListStore
	onNewEntryCreated NewEntryCreatedCallback
}

type NewEntryCreatedCallback func(entry *gttimeentry.TimeEntry)

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

func (f *EntryForm) GetTimeEntry() *gttimeentry.TimeEntry {
	var err error
	entry := &gttimeentry.TimeEntry{}
	if entry.Description, err = f.description.GetText(); err != nil {
		return nil
	}
	log.Println(1)
	if entry.Pid, err = strconv.ParseUint(f.project.GetActiveID(), 10, 64); err != nil {
		return nil
	}
	log.Println(2)
	var tag string
	if tag = f.tag.GetActiveID(); tag == "" {
		return nil
	} else {
		entry.Tags = []string{tag}
	}
	log.Println(3)

	return entry
}

func (f *EntryForm) SetTimeEntry(entry *gttimeentry.TimeEntry) {
	if entry == nil {
		f.description.SetText("")
		f.project.SetActive(-1)
		f.tag.SetActive(-1)
		f.button.SetImage(f.imageStart)
		f.start.SetText("")
		f.stop.SetText("")
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

func (f *EntryForm) Validate() bool {
	valid := f.GetTimeEntry() != nil
	if err := f.button.SetProperty("enabled", valid); err != nil {
		log.Fatal("validate error:", err)
	}
	return valid
}

func (f *EntryForm) OnNewEntryCreated(cb NewEntryCreatedCallback) {
	f.onNewEntryCreated = cb
}

func (f *EntryForm) onButtonClicked() {
	if !f.Validate() {
		return
	}
	if f.onNewEntryCreated != nil {
		f.onNewEntryCreated(f.GetTimeEntry())
	}
}

//
//func (j *jwtoggl.jApp) onStartStopClicked() {
//	var err error
//	var projectId, tagId int
//	var description string
//	log.Println("Project active id:", j.form.project.GetActiveID())
//	if projectId, err = strconv.Atoi(j.form.project.GetActiveID()); err != nil {
//		log.Fatal("Project cast error:", err)
//	}
//
//	if tagId, err = strconv.Atoi(j.form.tag.GetActiveID()); err != nil {
//		log.Fatal("Tag cast error:", err)
//	}
//	if description, err = j.form.description.GetText(); err != nil {
//		log.Fatal("Can't get description:", err)
//	}
//
//	log.Printf("Clicked: project %d tag %d description %s", projectId, tagId, description)
//	//
//	//now := time.Now()
//	//if err = j.db.CreateNewEntry(description, projectId, tagId, now); err != nil {
//	//	log.Fatal("Can't create entry:", err)
//	//}
//}
