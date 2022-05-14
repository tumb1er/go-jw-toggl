package store

// gttimeentry.tpl предоставляет методы для работы с моделями объектов Toggl

import (
	"go-jw-toggl/jwtoggl/dto"
	"go-jw-toggl/jwtoggl/gtoggl"
	"go-jw-toggl/jwtoggl/gtoggl/gttimeentry"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"time"
)

func (m *TimeEntryManager) Sync(items gttimeentry.TimeEntries, wid uint64) (bool, error) {
	filtered := make(gttimeentry.TimeEntries, 0, len(items))
	for _, item := range items {
		if item.WId == wid {
			filtered = append(filtered, item)
		}
	}
	qs := NewTimeEntries(filtered)
	return qs.Sync(m.gorm())
}

func (m *TimeEntryManager) gorm() *gorm.DB {
	return (*gorm.DB)(m).Preload(clause.Associations).Session(&gorm.Session{})
}

func NewTimeEntries(items ...gttimeentry.TimeEntries) *TimeEntries {
	if len(items) == 0 {
		result := make(TimeEntries, 0)
		return &result
	}
	result := make(TimeEntries, len(items[0]))
	for i, item := range items[0] {
		result[i] = TimeEntry{TimeEntry: item}
		result[i].ComputeUpdated()
	}
	return &result
}

func (qs TimeEntries) DTO() []dto.TimeEntry {
	result := make([]dto.TimeEntry, len(qs))
	for i, c := range qs {
		result[i] = c
	}
	return result
}

func (qs TimeEntries) Sync(db *gorm.DB) (changed bool, err error) {
	if err = qs.Preload(db); err != nil {
		return
	}
	err = db.Transaction(func(tx *gorm.DB) error {
		ts := time.Now().Local()
		existing := NewTimeEntries()
		Must(tx.Find(existing))
		extIdMap := make(map[uint64]TimeEntry)
		for _, obj := range *existing {
			extIdMap[obj.Id] = obj
		}

		for _, obj := range qs {
			obj.Fetched = &ts
			if prev, ok := extIdMap[obj.Id]; ok {
				if !obj.After(prev.Modified) {
					// log.Printf("Skip saving TimeEntry %d %s\n\n", obj.Id, obj.Name)
					obj.TimeEntry = prev.TimeEntry
					obj.Modified = prev.Modified
					obj.Deleted = prev.Deleted
				} else {
					obj.Modified = obj.Updated
				}
				obj.PK = prev.PK
				if !gtoggl.TimeEntriesEqual(&obj.TimeEntry, &prev.TimeEntry) {
					changed = true
				}
				Must(tx.Save(&obj))
			} else {
				changed = true
				obj.Modified = obj.Updated
				Must(tx.Create(&obj))
			}
		}
		deleted := MustOrNotFound(
			tx.
				Where("modified <= updated AND fetched < ?", ts).
				Delete(TimeEntry{}))
		changed = deleted || changed
		return nil
	})
	return changed, err
}

func listTimeEntries(scope *gorm.DB) TimeEntries {
	qs := NewTimeEntries()
	Must(scope.Find(qs))
	return *qs
}

func (obj TimeEntry) GetName() string {
	return obj.Name
}

func (obj TimeEntry) GetPK() uint {
	return obj.PK
}

func (obj *TimeEntry) ComputeUpdated() {
	if !obj.TimeEntry.At.IsZero() {
		obj.Updated = &obj.TimeEntry.At
	}
}
