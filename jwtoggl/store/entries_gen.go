package store

// gttimeentry.tpl предоставляет методы для работы с моделями объектов Toggl

import (
	"github.com/jinzhu/gorm"
	"go-jw-toggl/jwtoggl/dto"
	"go-jw-toggl/jwtoggl/gtoggl/gttimeentry"
	"go-jw-toggl/jwtoggl/tools"
	"time"
)

func (m *TimeEntryManager) Sync(items gttimeentry.TimeEntries, ts time.Time, wid uint64) (bool, error) {
	filtered := make(gttimeentry.TimeEntries, 0, len(items))
	for _, item := range items {
		if item.WId == wid {
			filtered = append(filtered, item)
		}
	}
	qs := NewTimeEntries(filtered)
	return qs.Sync(m.gorm(), ts)
}

func (m *TimeEntryManager) gorm() *gorm.DB {
	return (*gorm.DB)(m).Set("gorm:auto_preload", true)
}

func NewTimeEntries(items ...gttimeentry.TimeEntries) *TimeEntries {
	if len(items) == 0 {
		result := make(TimeEntries, 0)
		return &result
	}
	result := make(TimeEntries, len(items[0]))
	for i, item := range items[0] {
		result[i] = TimeEntry{TimeEntry: item}
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

func (qs TimeEntries) Sync(db *gorm.DB, ts time.Time) (changed bool, err error) {
	err = db.Transaction(func(tx *gorm.DB) error {
		existing := NewTimeEntries()
		Must(tx.Find(existing))
		extIdMap := make(map[uint64]TimeEntry)
		for _, obj := range *existing {
			extIdMap[obj.Id] = obj
		}

		for _, obj := range qs {
			if prev, ok := extIdMap[obj.Id]; ok {
				if prev.SavedAt.After(obj.UpdatedAt) {
					// log.Printf("Skip saving TimeEntry %d %s\n\n", obj.Id, obj.Name)
					obj.TimeEntry = prev.TimeEntry
				}
				obj.PK = prev.PK
				if !tools.TimeEntriesEqual(&obj.TimeEntry, &prev.TimeEntry) {
					changed = true
				}
				Must(tx.Save(&obj))
			} else {
				changed = true
				Must(tx.Create(&obj))
			}
		}
		changed = MustOrNotFound(tx.Where("saved_at < ?", ts).Delete(TimeEntry{})) || changed
		return nil
	})
	return changed, err
}

func listTimeEntries(scope *gorm.DB) TimeEntries {
	qs := NewTimeEntries()
	Must(scope.Find(qs))
	return *qs
}

func (obj TimeEntry) GetTogglID() uint64 {
	return obj.Id
}

func (obj TimeEntry) GetName() string {
	return obj.Name
}
