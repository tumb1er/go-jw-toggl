package store

//go:generate bash -c "sed 's/Foos/TimeEntries/g;s/Foo/TimeEntry/g;s/foo/gttimeentry/g' foo.tpl > entries_gen.go"

import (
	"github.com/jinzhu/gorm"
	"go-jw-toggl/jwtoggl/gtoggl/gttimeentry"
	"time"
)

type TimeEntry struct {
	Synced
	gttimeentry.TimeEntry
	Tag string
}

type TimeEntries []TimeEntry

type TimeEntryManager gorm.DB

func (e *TimeEntry) BeforeCreate() error {
	e.loadTagFromTags()
	return nil
}

func (e *TimeEntry) loadTagFromTags() {
	e.Tag = ""
	if len(e.Tags) > 0 {
		e.Tag = e.Tags[0]
	}
}

func (e *TimeEntry) BeforeSave() error {
	e.loadTagFromTags()
	return nil
}

func (e *TimeEntry) AfterFind() error {
	if e.Tag != "" {
		e.Tags = []string{e.Tag}
	} else {
		e.Tags = []string{}
	}
	return nil
}

func (m *TimeEntryManager) List() gttimeentry.TimeEntries {
	return listTimeEntries(m.gorm().Where("stop <> ?", time.Time{}).Order("start DESC"))
}

func (m *TimeEntryManager) Delete(extId uint64) bool {
	return MustOrNotFound(m.gorm().Where("id = ?", extId).Delete(&TimeEntry{}))
}

func (m *TimeEntryManager) Current() *gttimeentry.TimeEntry {
	var entry TimeEntry
	if MustOrNotFound(m.gorm().Where("stop = ?", time.Time{}).First(&entry)) {
		return &entry.TimeEntry
	}
	return nil
}

func (m *TimeEntryManager) Create(timeEntry gttimeentry.TimeEntry) {
	entry := &TimeEntry{TimeEntry: timeEntry}
	Must(m.gorm().Create(entry))
}

func (m *TimeEntryManager) Update(timeEntry gttimeentry.TimeEntry) {
	entry := &TimeEntry{TimeEntry: timeEntry}
	Must(m.gorm().Model(entry).Where("id = ?", timeEntry.Id).Updates(entry))
}

func (m *TimeEntryManager) Get(extId uint64) *gttimeentry.TimeEntry {
	var entry TimeEntry
	if MustOrNotFound(m.gorm().Where("id = ?", extId).First(&entry)) {
		return &entry.TimeEntry
	}
	return nil
}
