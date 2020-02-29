package jwtoggl

import (
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
	return f, nil
}

func (f *EntryForm) SetProjects(projects gtproject.Projects) error {
	ls := f.projectsListStore
	ls.Clear()
	for _, project := range projects {
		it := ls.Append()
		if err := ls.Set(it, []int{0, 1}, []interface{}{project.Id, project.Name}); err != nil {
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
		if err := ls.Set(it, []int{0, 1}, []interface{}{tag.Id, tag.Name}); err != nil {
			return err
		}
	}
	return nil
}

func (j *jApp) onStartStopClicked() {
	it, err := j.form.project.GetActiveIter()
	if err != nil {
		log.Fatal(err)
	}
	v, err := j.form.projectsListStore.GetValue(it, 0)
	if err != nil {
		log.Fatal(err)
	}
	projectId, err := v.GoValue()
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Clicked: %+v", projectId)
}
