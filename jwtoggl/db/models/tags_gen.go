package models

import (
	"github.com/jinzhu/gorm"
	"go-jw-toggl/jwtoggl/gtoggl/gttag"
	"go-jw-toggl/jwtoggl/tools"
	"time"
)

func (obj *Tag) BeforeCreate() error {
	return obj.dump(obj.Tag)
}

func (obj *Tag) BeforeSave() error {
	return obj.dump(obj.Tag)
}

func (obj *Tag) AfterFind() error {
	return obj.load(&obj.Tag)
}

func NewTags(items ...gttag.Tags) *Tags {
	if len(items) == 0 {
		result := make(Tags, 0)
		return &result
	}
	result := make(Tags, len(items[0]))
	for i, c := range items[0] {
		result[i] = Tag{Tag: c}
	}
	return &result
}

func (qs Tags) Sync(db *gorm.DB, ts time.Time) (changed bool, err error) {
	err = db.Transaction(func(tx *gorm.DB) error {
		existing := NewTags()
		if res := tx.Find(existing); res.Error != nil {
			return res.Error
		}
		extIdMap := make(map[uint64]Tag)
		for _, obj := range *existing {
			extIdMap[obj.Id] = obj
		}

		for _, obj := range qs {
			obj.SavedAt = ts
			if prev, ok := extIdMap[obj.Id]; ok {
				obj.PK = prev.PK
				if !tools.TagsEqual(&obj.Tag, &prev.Tag) {
					changed = true
				}
				tx.Save(&obj)
			} else {
				changed = true
				tx.Create(&obj)
			}
		}
		res := tx.Where("saved_at < ?", ts).Delete(Tag{})
		if res.Error == nil && res.RowsAffected > 0 {
			changed = true
		}
		return res.Error
	})
	return changed, err
}
