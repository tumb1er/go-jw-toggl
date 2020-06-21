package store

//go:generate bash -c "sed 's/Foos/TimeEntries/g;s/Foo/TimeEntry/g;s/foo/gttimeentry/g' foo.tpl > entries_gen.go"

import (
	"github.com/jinzhu/gorm"
	"go-jw-toggl/jwtoggl/dto"
	"go-jw-toggl/jwtoggl/gtoggl/gttimeentry"
	"strings"
	"time"
)

type TimeEntry struct {
	Synced
	gttimeentry.TimeEntry
	Tag string
}

func (e TimeEntry) GetTags() []string {
	if e.Tag == "" {
		return []string{}
	}
	return []string{e.Tag}
}

type TimeEntries []TimeEntry

type TimeEntryManager gorm.DB

func NewTimeEntry(entry *gttimeentry.TimeEntry) TimeEntry {
	if entry == nil {
		return TimeEntry{}
	}
	timeEntry := TimeEntry{
		TimeEntry: *entry,
		Tag:       "",
	}
	if len(entry.Tags) > 0 {
		timeEntry.Tag = entry.Tags[0]
	}
	return timeEntry
}

func (e TimeEntry) Nil() bool {
	return e.PK == 0
}

func (e TimeEntry) IsCurrent() bool {
	return e.Stop.IsZero()
}

func (e TimeEntry) GetDuration() time.Duration {
	return e.Stop.Sub(e.Start)
}

func (e TimeEntry) GetStart() time.Time {
	return e.Start
}

func (e TimeEntry) TaskKey() string {
	key := e.Description
	if strings.HasPrefix(e.Description, "#") {
		key = strings.SplitN(e.Description, " ", 2)[0]
	}
	if len(e.Tags) > 0 {
		key += e.Tags[0]
	}
	return key
}

func (e TimeEntry) GetDescription() string {
	return e.Description
}

func (e TimeEntry) GetProjectTogglID() uint64 {
	return e.Pid
}

func (e TimeEntry) GetTag() string {
	if len(e.Tags) > 0 {
		return e.Tags[0]
	}
	return ""
}

func (e TimeEntry) GetStop() time.Time {
	return e.Stop
}

func (e TimeEntry) GetTogglID() uint64 {
	return e.Id
}

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

func (e TimeEntry) Equals(other dto.TimeEntry) bool {
	if e.Nil() != other.Nil() {
		return false
	}
	if e.Nil() && other.Nil() {
		return true
	}
	return e.GetDescription() == other.GetDescription() &&
		e.GetProjectTogglID() == other.GetProjectTogglID() &&
		e.GetTag() == other.GetTag() &&
		e.GetStart() == other.GetStart() &&
		e.GetStop() == other.GetStop()
}

func (m *TimeEntryManager) List() TimeEntries {
	return listTimeEntries(m.gorm().Where("stop <> ?", time.Time{}).Order("start DESC"))
}

func (m *TimeEntryManager) Delete(extId uint64) bool {
	return MustOrNotFound(m.gorm().Where("id = ?", extId).Delete(&TimeEntry{}))
}

func (m *TimeEntryManager) Current() TimeEntry {
	var entry TimeEntry
	MustOrNotFound(m.gorm().Where("stop = ?", time.Time{}).First(&entry))
	return entry
}

func (m *TimeEntryManager) Create(timeEntry *TimeEntry) {
	Must(m.gorm().Create(timeEntry))
}

func (m *TimeEntryManager) Update(timeEntry *TimeEntry) {
	Must(m.gorm().Model(timeEntry).Where("id = ?", timeEntry.Id).Updates(timeEntry))
}

func (m *TimeEntryManager) Get(extId uint64) *TimeEntry {
	var entry TimeEntry
	if MustOrNotFound(m.gorm().Where("id = ?", extId).First(&entry)) {
		return &entry
	}
	return nil
}
