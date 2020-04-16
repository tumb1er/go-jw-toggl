package models

//go:generate bash -c "sed 's/Foos/TimeEntries/g;s/Foo/TimeEntry/g;s/foo/gttimeentry/g' foo.tpl > entries_gen.go"

import (
	"github.com/jinzhu/gorm"
	"go-jw-toggl/jwtoggl/gtoggl/gttimeentry"
	"time"
)

type TimeEntry struct {
	Synced
	gttimeentry.TimeEntry `gorm:"-"`
	Current               bool
	Start                 time.Time
}

func (obj *TimeEntry) dump(out interface{}) error {
	if err := obj.Synced.dump(out); err != nil {
		return err
	}
	dto := out.(gttimeentry.TimeEntry)
	obj.Current = dto.Stop.IsZero()
	obj.Start = dto.Start
	return nil
}

type TimeEntries []TimeEntry

func ListEntries(db *gorm.DB) gttimeentry.TimeEntries {
	qs := NewTimeEntries()
	length := Must(db.Order("start DESC").Find(qs))
	result := make(gttimeentry.TimeEntries, length)
	for i, c := range *qs {
		result[i] = c.TimeEntry
	}
	return result
}

func (obj TimeEntry) Delete(db *gorm.DB) {
	MustOrNotFound(db.Delete(&obj))
}

func DeleteEntry(db *gorm.DB, extId uint64) bool {
	return MustOrNotFound(db.Where("ext_id = ?", extId).Delete(&TimeEntry{}))
}

func CurrentEntry(db *gorm.DB) *gttimeentry.TimeEntry {
	var entry TimeEntry
	if MustOrNotFound(db.Where("current").First(&entry)) {
		return &entry.TimeEntry
	}
	return nil
}

func CreateEntry(db *gorm.DB, timeEntry gttimeentry.TimeEntry) {
	entry := &TimeEntry{TimeEntry: timeEntry}
	Must(db.Create(entry))
}

func UpdateEntry(db *gorm.DB, timeEntry gttimeentry.TimeEntry) {
	entry := &TimeEntry{TimeEntry: timeEntry}
	Must(db.Where("ext_id = ?", timeEntry.Id).Updates(entry))
}

func EntryById(db *gorm.DB, extId uint64) *gttimeentry.TimeEntry {
	var entry TimeEntry
	if MustOrNotFound(db.Where("ext_id = ?", extId).First(&entry)) {
		return &entry.TimeEntry
	}
	return nil
}
