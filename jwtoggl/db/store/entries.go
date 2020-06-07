package store

//go:generate bash -c "sed 's/Foos/TimeEntries/g;s/Foo/TimeEntry/g;s/foo/gttimeentry/g' foo.tpl > entries_gen.go"

import (
	"github.com/jinzhu/gorm"
	"go-jw-toggl/jwtoggl/gtoggl/gttimeentry"
	"log"
	"time"
)

type TimeEntry struct {
	Synced
	gttimeentry.TimeEntry `gorm:"-"`
	Current               *bool
	Start                 time.Time
}

func (obj *TimeEntry) dump(out interface{}) error {
	if err := obj.Synced.dump(out); err != nil {
		return err
	}
	dto := out.(gttimeentry.TimeEntry)
	current := dto.Stop.IsZero()
	obj.Current = &current
	obj.Start = dto.Start
	return nil
}

type TimeEntries []TimeEntry

type TimeEntryManager gorm.DB

func (m *TimeEntryManager) List() gttimeentry.TimeEntries {
	return listTimeEntries(m.gorm().Where("not current").Order("start DESC"))
}

func (m *TimeEntryManager) Delete(extId uint64) bool {
	return MustOrNotFound(m.gorm().Where("ext_id = ?", extId).Delete(&TimeEntry{}))
}

func (m *TimeEntryManager) Current() *gttimeentry.TimeEntry {
	var entry TimeEntry
	if MustOrNotFound(m.gorm().Where("current").First(&entry)) {
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
	if err := entry.dump(entry.TimeEntry); err != nil {
		log.Fatal(err)
	}
	// Используем map т.к. нулевое значение entry.Current не проставляется.
	Must(m.gorm().Debug().Model(entry).Where("ext_id = ?", timeEntry.Id).
		Updates(entry))
}

func (m *TimeEntryManager) Get(extId uint64) *gttimeentry.TimeEntry {
	var entry TimeEntry
	if MustOrNotFound(m.gorm().Where("ext_id = ?", extId).First(&entry)) {
		return &entry.TimeEntry
	}
	return nil
}
