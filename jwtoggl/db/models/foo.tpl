package models

import (
	"github.com/jinzhu/gorm"
	"go-jw-toggl/jwtoggl/gtoggl/foo"
	"go-jw-toggl/jwtoggl/tools"
	"time"
)

func (obj *Foo) BeforeCreate() error {
	return obj.dump(obj.Foo)
}

func (obj *Foo) BeforeSave() error {
	return obj.dump(obj.Foo)
}

func (obj *Foo) AfterFind() error {
	return obj.load(&obj.Foo)
}

func NewFoos(items ...foo.Foos) *Foos {
	if len(items) == 0 {
		result := make(Foos, 0)
		return &result
	}
	result := make(Foos, len(items[0]))
	for i, c := range items[0] {
		result[i] = Foo{Foo: c}
	}
	return &result
}

func (qs Foos) Sync(db *gorm.DB, ts time.Time) (changed bool, err error) {
	err = db.Transaction(func(tx *gorm.DB) error {
		existing := NewFoos()
		if res := tx.Find(existing); res.Error != nil {
			return res.Error
		}
		extIdMap := make(map[uint64]Foo)
		for _, obj := range *existing {
			extIdMap[obj.Id] = obj
		}

		for _, obj := range qs {
			obj.SavedAt = ts
			if prev, ok := extIdMap[obj.Id]; ok {
				obj.PK = prev.PK
				if !tools.FoosEqual(&obj.Foo, &prev.Foo) {
					changed = true
				}
				tx.Save(&obj)
			} else {
				changed = true
				tx.Create(&obj)
			}
		}
		res := tx.Where("saved_at < ?", ts).Delete(Foo{})
		if res.Error == nil && res.RowsAffected > 0 {
			changed = true
		}
		return res.Error
	})
	return changed, err
}
