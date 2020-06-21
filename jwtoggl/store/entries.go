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

func (obj TimeEntry) GetTags() []string {
	if obj.Tag == "" {
		return []string{}
	}
	return []string{obj.Tag}
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

func (obj TimeEntry) Nil() bool {
	return obj.PK == 0
}

func (obj TimeEntry) IsCurrent() bool {
	return obj.Stop.IsZero()
}

func (obj TimeEntry) GetDuration() time.Duration {
	return obj.Stop.Sub(obj.Start)
}

func (obj TimeEntry) GetStart() time.Time {
	return obj.Start
}

func (obj TimeEntry) TaskKey() string {
	key := obj.Name
	if strings.HasPrefix(obj.Name, "#") {
		key = strings.SplitN(obj.Name, " ", 2)[0]
	}
	if len(obj.Tags) > 0 {
		key += obj.Tags[0]
	}
	return key
}

func (obj TimeEntry) GetProjectTogglID() uint64 {
	return obj.Pid
}

func (obj TimeEntry) GetTag() string {
	if len(obj.Tags) > 0 {
		return obj.Tags[0]
	}
	return ""
}

func (obj TimeEntry) GetStop() time.Time {
	return obj.Stop
}

func (obj *TimeEntry) BeforeCreate() error {
	obj.loadTagFromTags()
	return nil
}

func (obj *TimeEntry) loadTagFromTags() {
	obj.Tag = ""
	if len(obj.Tags) > 0 {
		obj.Tag = obj.Tags[0]
	}
}

func (obj *TimeEntry) BeforeSave() error {
	obj.loadTagFromTags()
	return nil
}

func (obj *TimeEntry) AfterFind() error {
	if obj.Tag != "" {
		obj.Tags = []string{obj.Tag}
	} else {
		obj.Tags = []string{}
	}
	return nil
}

func (obj TimeEntry) Equals(other dto.TimeEntry) bool {
	if obj.Nil() != other.Nil() {
		return false
	}
	if obj.Nil() && other.Nil() {
		return true
	}
	return obj.GetName() == other.GetName() &&
		obj.GetProjectTogglID() == other.GetProjectTogglID() &&
		obj.GetTag() == other.GetTag() &&
		obj.GetStart() == other.GetStart() &&
		obj.GetStop() == other.GetStop()
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
