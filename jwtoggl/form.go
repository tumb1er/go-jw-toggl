package jwtoggl

import (
	"fmt"
	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/gtk"
	"go-jw-toggl/jwtoggl/gtoggl/gtproject"
	"go-jw-toggl/jwtoggl/gtoggl/gttag"
	"image/color"
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
	cols := []int{0, 1, 2}
	var c color.RGBA
	for _, project := range projects {
		it := ls.Append()
		_, err := fmt.Sscanf(project.HexColor, "#%02x%02x%02x", &c.R, &c.G, &c.B)
		if err != nil {
			return err
		}
		html := fmt.Sprintf("<span foreground=\"%s\">%s</span>", project.HexColor, project.Name)
		values := []interface{}{project.Id, html, *gdk.NewRGBA(
			float64(c.R), float64(c.G), float64(c.B), float64(c.A))}
		if err := ls.Set(it, cols, values); err != nil {
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
