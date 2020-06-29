package store

import (
	"github.com/jinzhu/gorm"
	"go-jw-toggl/jwtoggl/gtoggl/gttimeentry"
	"log"
	"time"
)

type TimeEntryManager gorm.DB

func (m *TimeEntryManager) FromTimeEntry(entry gttimeentry.TimeEntry) *TimeEntry {
	timeEntry := m.Get(entry.Id)
	if timeEntry == nil {
		timeEntry = &TimeEntry{}
	}
	timeEntry.TimeEntry = entry
	if timeEntry.Project.Id != entry.Pid {
		timeEntry.Project.PK = 0
		Must(m.gorm().Where("id = ?", entry.Pid).Find(&timeEntry.Project))
		timeEntry.ProjectPK = timeEntry.Project.PK
	}
	return timeEntry
}

func (m *TimeEntryManager) List() TimeEntries {
	return listTimeEntries(
		m.gorm().
			Where("stop <> ? AND deleted IS NULL", time.Time{}).
			Order("start DESC"))
}

func (m *TimeEntryManager) DeletedEntries() TimeEntries {
	return listTimeEntries(
		m.gorm().
			Where("id != 0 AND deleted IS NOT NULL"))
}

func (m *TimeEntryManager) NewEntries() TimeEntries {
	return listTimeEntries(
		m.gorm().
			Where("id == 0 AND deleted IS NULL"))
}

func (m *TimeEntryManager) ChangedEntries() TimeEntries {
	return listTimeEntries(
		m.gorm().
			Where("id <> 0 AND modified > updated AND deleted IS NULL"))
}

func (m *TimeEntryManager) Delete(pk uint) bool {
	return MustOrNotFound(m.gorm().Where("pk = ?", pk).Delete(&TimeEntry{}))
}

func (m *TimeEntryManager) MarkDeleted(pk uint) bool {
	up := &TimeEntry{}
	up.MarkDeleted()
	return MustOrNotFound(m.gorm().Model(up).Where("pk = ?", pk).Update(up))
}

func (m *TimeEntryManager) Current() TimeEntry {
	var entry TimeEntry
	MustOrNotFound(m.gorm().
		Where("stop = ?", time.Time{}).
		First(&entry))
	return entry
}

func (m *TimeEntryManager) Create(timeEntry *TimeEntry) {
	Must(m.gorm().Create(timeEntry))
}

func (m *TimeEntryManager) Get(extId uint64) *TimeEntry {
	var entry TimeEntry
	if MustOrNotFound(m.gorm().
		Where("id = ?", extId).First(&entry)) {
		return &entry
	}
	return nil
}

func (m *TimeEntryManager) GetByPK(pk uint) *TimeEntry {
	var entry TimeEntry
	if MustOrNotFound(m.gorm().
		Where("pk = ?", pk).First(&entry)) {
		return &entry
	}
	return nil
}

func (m *TimeEntryManager) Update(timeEntry *TimeEntry) {
	Must(m.gorm().Model(timeEntry).
		Where("pk = ?", timeEntry.PK).
		Updates(timeEntry))
	if len(timeEntry.TagSet) > 0 {
		res := m.gorm().
			Model(timeEntry).
			Association("TagSet").
			Replace(timeEntry.TagSet)
		if res.Error != nil {
			log.Fatal(res.Error)
		}
	}
}
