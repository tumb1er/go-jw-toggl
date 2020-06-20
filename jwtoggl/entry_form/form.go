package entry_form

import "C"
import (
	"go-jw-toggl/jwtoggl/builder"
	"go-jw-toggl/jwtoggl/dto"
	"go-jw-toggl/jwtoggl/entry_form/gtk_entry_form"
	"go-jw-toggl/jwtoggl/gtoggl/gttimeentry"
	"go-jw-toggl/jwtoggl/tools"
	"log"
	"time"
)

type EntryForm struct {
	g                 *gtk_entry_form.GtkEntryForm
	timeEntry         *gttimeentry.TimeEntry
	onNewEntryCreated EntryCallback
	onEntryStopped    EntryCallback
	onEntryChanged    EntryCallback
	waiter            *time.Timer
	uiEntry           entry
}

type EntryCallback func(entry *gttimeentry.TimeEntry)

func NewEntryForm(b *builder.Builder) (*EntryForm, error) {
	f := &EntryForm{uiEntry: entry{Start: &time.Time{}}}
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
	entry := f.uiEntry.TimeEntry()
	if entry != nil && entry.Id != 0 && !tools.TimeEntriesEqual(f.timeEntry, entry) {
		log.Println("UPDATE!!!", entry.Start)
		if f.onEntryChanged != nil {
			f.onEntryChanged(entry)
		}
		f.timeEntry = entry
	} else {
		log.Println("No update")
	}
}

func (f *EntryForm) OnButtonClicked() {
	if f.timeEntry != nil {
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
	entry := f.uiEntry.TimeEntry()
	if entry != nil {
		f.onNewEntryCreated(entry)
	}
}

func (f *EntryForm) handleStop() {
	log.Println("handleStop")
	if f.onEntryStopped != nil {
		f.onEntryStopped(f.timeEntry)
		f.SetTimeEntry(nil)
	}
}

func (f *EntryForm) SetTags(tags []dto.Tag) {
	f.g.SetTags(tags)
}

func (f *EntryForm) SetProjects(projects []dto.Project) {
	f.g.SetProjects(projects)
}

func (f *EntryForm) SetTimeEntry(entry *gttimeentry.TimeEntry) {
	f.timeEntry = entry
	f.g.SetTimeEntry(entry)
}

func (f *EntryForm) UpdateDuration() {
	f.g.UpdateDuration()
}
