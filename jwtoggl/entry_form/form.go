package entry_form

import "C"
import (
	"go-jw-toggl/jwtoggl/builder"
	"go-jw-toggl/jwtoggl/entry_form/gtk_entry_form"
	"go-jw-toggl/jwtoggl/gtoggl/gtproject"
	"go-jw-toggl/jwtoggl/gtoggl/gttag"
	"go-jw-toggl/jwtoggl/gtoggl/gttimentry"
	"go-jw-toggl/jwtoggl/tools"
	"log"
	"time"
)

type EntryForm struct {
	g                 *gtk_entry_form.GtkEntryForm
	eid               uint64
	onNewEntryCreated EntryCallback
	onEntryStopped    EntryCallback
	onEntryChanged    EntryCallback
	waiter            *time.Timer
}

type EntryCallback func(entry *gttimeentry.TimeEntry)

func NewEntryForm(b *builder.Builder) (*EntryForm, error) {
	g, err := gtk_entry_form.New(b)
	if err != nil {
		return nil, err
	}
	f := &EntryForm{g: g}

	err = g.ConnectSignals(
		f.onButtonClicked,
		f.onDescriptionChanged,
		f.onActiveProjectChanged,
		f.onActiveTagChanged,
		f.onEditDone)
	if err != nil {
		return nil, err
	}
	return f, nil
}

func (f *EntryForm) Validate() {
	tools.Schedule(f.validate)
}

func (f *EntryForm) validate() {
	f.g.GetTimeEntry(func(entry *gttimeentry.TimeEntry, err error) {
		valid := entry != nil && err == nil
		log.Println("validate:", valid)
		f.g.SetButtonSensitive(valid)
	})
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

func (f *EntryForm) onActiveProjectChanged() {
	f.g.SetActiveProject()
	f.onEditDone()
}

func (f *EntryForm) onActiveTagChanged() {
	f.g.SetActiveTag()
	f.onEditDone()
}
func (f *EntryForm) onDescriptionChanged() {
	f.Validate()
}

func (f *EntryForm) onEditDone() {
	log.Println("EditDone")
	if f.waiter != nil {
		f.waiter.Stop()
	}
	f.waiter = time.AfterFunc(5*time.Second, f.onEditCompleted)
}

func (f *EntryForm) onEditCompleted() {
	log.Println("EditCompleted")
	f.waiter = nil
	// TODO compare with current entry
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
	f.g.GetTimeEntry(func(entry *gttimeentry.TimeEntry, err error) {
		if entry != nil && f.onNewEntryCreated != nil {
			f.onNewEntryCreated(entry)
		}
	})
}

func (f *EntryForm) handleStop() {
	log.Println("handleStop")
	f.g.GetTimeEntry(func(entry *gttimeentry.TimeEntry, err error) {
		if entry != nil && f.onEntryStopped != nil {
			entry.Id = f.eid
			f.onEntryStopped(entry)
			f.g.SetTimeEntry(nil)
		}
	})
}

func (f *EntryForm) SetTags(tags gttag.Tags) {
	f.g.SetTags(tags)
}

func (f *EntryForm) SetProjects(projects gtproject.Projects) {
	f.g.SetProjects(projects)
}

func (f *EntryForm) SetTimeEntry(entry *gttimeentry.TimeEntry) {
	if entry != nil {
		f.eid = entry.Id
	} else {
		f.eid = 0
	}
	f.g.SetTimeEntry(entry)
}
