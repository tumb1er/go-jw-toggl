package entry_form

import "C"
import (
	"go-jw-toggl/jwtoggl/builder"
	"go-jw-toggl/jwtoggl/dto"
	"go-jw-toggl/jwtoggl/entry_form/gtk_entry_form"
	"log"
	"time"
)

type EntryForm struct {
	g                 *gtk_entry_form.GtkEntryForm
	timeEntry         dto.TimeEntry
	onNewEntryCreated EntryCallback
	onEntryStopped    EntryCallback
	onEntryChanged    EntryCallback
	waiter            *time.Timer
	uiEntry           entry
}

type EntryCallback func(dto.TimeEntry)

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

	err = g.ConnectSignals()
	if err != nil {
		return nil, err
	}
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
	if f.timeEntry.GetName() != f.uiEntry.GetName() {
		log.Println("name changed")
		changed = true
	}
	if f.timeEntry.GetProject().GetTogglID() != f.uiEntry.GetProject().GetTogglID() {
		log.Println("project changed")
		changed = true
	}
	if f.timeEntry.GetTag() != f.uiEntry.GetTag() {
		log.Println("tag changed")
		changed = true
	}
	if f.timeEntry.GetStart() != f.uiEntry.GetStart() {
		log.Println("start changed")
		changed = true
	}
	if !changed {
		log.Println("Nothing changed")
		return
	}

	log.Println("UPDATE!!!", f.uiEntry, f.timeEntry)
	f.timeEntry = f.uiEntry // копирование структуры
	if f.onEntryChanged != nil {
		f.onEntryChanged(f.timeEntry)
	}
}

func (f *EntryForm) OnButtonClicked() {
	if !f.timeEntry.Nil() {
		f.handleStop()
	} else {
		if f.uiEntry.Start == nil || f.uiEntry.Start.IsZero() {
			start := time.Now()
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

func (f *EntryForm) OnProjectChanged(pid uint64) bool {
	defer f.validate()
	log.Printf("OnProjectChanged: %d\n", pid)
	f.uiEntry.PId = pid
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
	f.uiEntry.Description = text
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

func (f *EntryForm) SetTimeEntry(entry dto.TimeEntry) {
	f.timeEntry = entry
	if !entry.Nil() {
		f.uiEntry.Id = entry.GetTogglID()
		start := entry.GetStart()
		f.uiEntry.Start = &start
	} else {
		f.uiEntry.Id = 0
		f.uiEntry.Start = &time.Time{}
	}
	f.g.SetTimeEntry(entry)
}

func (f *EntryForm) UpdateDuration() {
	f.g.UpdateDuration()
}
