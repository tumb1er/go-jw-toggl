package models

import (
	"github.com/jinzhu/gorm"
	"go-jw-toggl/jwtoggl/gtoggl/gttimeentry"
	"go-jw-toggl/jwtoggl/tools"
	"time"
)

func (obj *TimeEntry) BeforeCreate() error {
	return obj.dump(obj.TimeEntry)
}
func (obj *TimeEntry) BeforeSave() error {
	return obj.dump(obj.TimeEntry)
}

func (obj *TimeEntry) AfterFind() error {
	return obj.load(&obj.TimeEntry)
}

func NewTimeEntries(items ...gttimeentry.TimeEntries) *TimeEntries {
	if len(items) == 0 {
		result := make(TimeEntries, 0)
		return &result
	}
	result := make(TimeEntries, len(items[0]))
	for i, c := range items[0] {
		result[i] = TimeEntry{TimeEntry: c}
	}
	return &result
}

func (qs TimeEntries) Sync(db *gorm.DB, ts time.Time) (changed bool, err error) {
	err = db.Transaction(func(tx *gorm.DB) error {
		existing := NewTimeEntries()
		if res := tx.Find(existing); res.Error != nil {
			return res.Error
		}
		extIdMap := make(map[uint64]TimeEntry)
		for _, obj := range *existing {
			extIdMap[obj.Id] = obj
		}

		for _, obj := range qs {
			obj.SavedAt = ts
			if prev, ok := extIdMap[obj.Id]; ok {
				obj.PK = prev.PK
				if !tools.TimeEntriesEqual(&obj.TimeEntry, &prev.TimeEntry) {
					changed = true
				}
				tx.Save(&obj)
			} else {
				changed = true
				tx.Create(&obj)
			}
		}
		res := tx.Where("saved_at < ?", ts).Delete(TimeEntry{})
		if res.Error == nil && res.RowsAffected > 0 {
			changed = true
		}
		return res.Error
	})
	return changed, err
}
