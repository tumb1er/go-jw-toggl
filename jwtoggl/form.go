package jwtoggl

import (
	"fmt"
	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"
	"go-jw-toggl/jwtoggl/gtoggl/gtproject"
	"go-jw-toggl/jwtoggl/gtoggl/gttag"
	"log"
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
}

func NewEntryForm(b *builder) (*EntryForm, error) {
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
	c, _ := b.GetEntry("text_project_entry")
	log.Printf("Entry %T: %+v\n", c, c)
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
	cols := []int{0, 1}
	for _, tag := range tags {
		it := ls.Append()
		values := []interface{}{tag.Id, tag.Name}
		if err := ls.Set(it, cols, values); err != nil {
			return err
		}
	}
	return nil
}

func (j *jApp) onStartStopClicked() {
	var it *gtk.TreeIter
	var err error
	var glibValue *glib.Value
	var goValue interface{}
	var projectId, tagId uint64
	var description string
	if it, err = j.form.project.GetActiveIter(); err != nil {
		log.Fatal("Can't get active project iter:", err)
	}
	if glibValue, err = j.form.projectsListStore.GetValue(it, 0); err != nil {
		log.Fatal("Can't get project value from store", err)
	}
	if goValue, err = glibValue.GoValue(); err != nil {
		log.Fatal("Can't get go value from glib", err)
	}
	if v, ok := goValue.(uint64); !ok {
		log.Fatalf("Can't cast projectId: %+v\n", v)
	} else {
		projectId = v
	}

	if it, err = j.form.tag.GetActiveIter(); err != nil {
		log.Fatal("Can't get active tag iter:", err)
	}
	if glibValue, err = j.form.tagsListStore.GetValue(it, 0); err != nil {
		log.Fatal("Can't get tag value from store", err)
	}
	if goValue, err = glibValue.GoValue(); err != nil {
		log.Fatal("Can't get go value from glib", err)
	}
	if v, ok := goValue.(uint64); !ok {
		log.Fatalf("Can't cast tagId: %+v\n", v)
	} else {
		tagId = v
	}
	if description, err = j.form.description.GetText(); err != nil {
		log.Fatal("Can't get description:", err)
	}

	log.Printf("Clicked: project %d tag %d description %s", projectId, tagId, description)
	//
	//now := time.Now()
	//if err = j.db.CreateNewEntry(description, projectId, tagId, now); err != nil {
	//	log.Fatal("Can't create entry:", err)
	//}
}
