package entry_form

import "C"
import (
	"go-jw-toggl/jwtoggl/builder"
	"go-jw-toggl/jwtoggl/dto"
	"go-jw-toggl/jwtoggl/entry_form/gtk_entry_form"
	"go-jw-toggl/jwtoggl/gtoggl"
	"log"
	"time"
)

type EntryForm struct {
	g                 *gtk_entry_form.GtkEntryForm
	timeEntry         dto.TimeEntrySaver
	onNewEntryCreated EntryCallback
	onEntryStopped    EntryCallback
	onEntryChanged    EntryCallback
	waiter            *time.Timer
	uiEntry           entry
}

type EntryCallback func(saver dto.TimeEntrySaver)

func NewEntryForm(b *builder.Builder) (*EntryForm, error) {
	f := &EntryForm{
		uiEntry:   entry{Start: &time.Time{}},
		timeEntry: entry{Start: &time.Time{}},
	}
	g, err := gtk_entry_form.New(b, f)
	if err != nil {
		return nil, err
	}
	f.g = g

	g.ConnectSignals()
	return f, nil
}

func (f *EntryForm) OnEditCompleted() {
	log.Println("OnEditCompleted")
	if !f.uiEntry.Validate() {
		log.Println("uiEntry validate failed")
		return
	}
	if f.timeEntry.Nil() {
		log.Println("state nil")
		return
	}
	changed := false
	if f.timeEntry.GetName() != f.uiEntry.Name {
		log.Printf("name changed: %s -> %s\n", f.timeEntry.GetName(), f.uiEntry.Name)
		changed = true
	}
	if f.timeEntry.GetProjectPK() != f.uiEntry.ProjectPK {
		log.Printf("project changed: %d -> %d\n", f.timeEntry.GetProjectPK(), f.uiEntry.ProjectPK)
		changed = true
	}
	if !gtoggl.TagSlicesEqual(f.timeEntry.GetTags(), f.uiEntry.GetTags()) {
		log.Printf("tag changed %+v %+v\n", f.timeEntry.GetTags(), f.uiEntry.GetTags())
		changed = true
	}
	if f.timeEntry.GetStart() != f.uiEntry.GetStart() {
		log.Printf("start changed: %s -> %s\n", f.timeEntry.GetStart(), f.uiEntry.GetStart())
		changed = true
	}
	if !changed {
		log.Println("Nothing changed")
		return
	}

	log.Println("UPDATE!!!")
	f.timeEntry = f.uiEntry // копирование структуры
	if f.onEntryChanged != nil {
		f.onEntryChanged(f.timeEntry)
	}
}

func (f *EntryForm) OnButtonClicked() {
	if !f.timeEntry.Nil() {
		f.handleStop()
	} else {
		if f.uiEntry.GetStart().IsZero() {
			start := time.Now().Local()
			f.uiEntry.Start = &start
		}
		f.handleStart()
	}
}

func (f *EntryForm) OnStartTimeChanged(start *time.Time) bool {
	defer f.validate()
	log.Printf("OnStartTimeChanged: %s\n", start)
	f.uiEntry.Start = start
	return f.uiEntry.ValidateStart()
}

func (f *EntryForm) OnProjectChanged(pk uint) bool {
	defer f.validate()
	log.Printf("OnProjectChanged: %d\n", pk)
	f.uiEntry.ProjectPK = pk
	return f.uiEntry.ValidateProject()
}

func (f *EntryForm) OnTagChanged(tag string) bool {
	defer f.validate()
	log.Printf("OnTagChanged: %s\n", tag)
	f.uiEntry.Tag = tag
	return f.uiEntry.ValidateTag()
}

func (f *EntryForm) OnDescriptionChanged(text string) bool {
	defer f.validate()
	log.Printf("OnDescriptionChanged: %s\n", text)
	f.uiEntry.Name = text
	return f.uiEntry.ValidateDescription()
}

func (f *EntryForm) validate() {
	valid := f.uiEntry.Validate()
	log.Println("validate:", valid)
	f.g.SetButtonSensitive(valid)
}

func (f *EntryForm) OnNewEntryCreated(cb EntryCallback) {
	f.onNewEntryCreated = cb
}
func (f *EntryForm) OnEntryStopped(cb EntryCallback) {
	f.onEntryStopped = cb
}
func (f *EntryForm) OnEntryChanged(cb EntryCallback) {
	f.onEntryChanged = cb
}

func (f *EntryForm) handleStart() {
	log.Println("handleStart")
	if f.onNewEntryCreated == nil {
		return
	}
	if f.uiEntry.Validate() {
		f.timeEntry = f.uiEntry // копирование структуры
		f.onNewEntryCreated(f.timeEntry)
	}
}

func (f *EntryForm) handleStop() {
	log.Println("handleStop")
	if f.onEntryStopped != nil {
		f.onEntryStopped(f.timeEntry)
		f.SetTimeEntry(entry{})
	}
}

func (f *EntryForm) SetTags(tags []dto.Tag) {
	f.g.SetTags(tags)
}

func (f *EntryForm) SetProjects(projects []dto.Project) {
	f.g.SetProjects(projects)
}

func (f *EntryForm) SetTimeEntry(entry dto.TimeEntrySaver) {
	f.timeEntry = entry
	if !entry.Nil() {
		f.uiEntry.PK = entry.GetPK()
		start := entry.GetStart()
		f.uiEntry.Start = &start
		f.uiEntry.Name = entry.GetName()
		f.uiEntry.ProjectPK = entry.GetProjectPK()
		f.uiEntry.Tag = gtoggl.Tag(entry.GetTags())
	} else {
		f.uiEntry.PK = 0
		f.uiEntry.Start = &time.Time{}
		f.uiEntry.Name = ""
		f.uiEntry.ProjectPK = 0
		f.uiEntry.Tag = ""

	}
	f.g.SetTimeEntry(f.uiEntry)
}

func (f *EntryForm) UpdateDuration() {
	f.g.UpdateDuration()
}
