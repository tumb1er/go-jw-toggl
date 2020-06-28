package store

// gttag.tpl предоставляет методы для работы с моделями объектов Toggl

import (
	"github.com/jinzhu/gorm"
	"go-jw-toggl/jwtoggl/dto"
	"go-jw-toggl/jwtoggl/gtoggl/gttag"
	"go-jw-toggl/jwtoggl/tools"
	"time"
)

func (m *TagManager) Sync(items gttag.Tags, ts time.Time, wid uint64) (bool, error) {
	filtered := make(gttag.Tags, 0, len(items))
	for _, item := range items {
		if item.WId == wid {
			filtered = append(filtered, item)
		}
	}
	qs := NewTags(filtered)
	return qs.Sync(m.gorm(), ts)
}

func (m *TagManager) gorm() *gorm.DB {
	return (*gorm.DB)(m).Set("gorm:auto_preload", true)
}

func NewTags(items ...gttag.Tags) *Tags {
	if len(items) == 0 {
		result := make(Tags, 0)
		return &result
	}
	result := make(Tags, len(items[0]))
	for i, item := range items[0] {
		result[i] = Tag{Tag: item}
		result[i].ComputeUpdated()
	}
	return &result
}

func (qs Tags) DTO() []dto.Tag {
	result := make([]dto.Tag, len(qs))
	for i, c := range qs {
		result[i] = c
	}
	return result
}

func (qs Tags) Sync(db *gorm.DB, ts time.Time) (changed bool, err error) {
	err = db.Transaction(func(tx *gorm.DB) error {
		existing := NewTags()
		Must(tx.Find(existing))
		extIdMap := make(map[uint64]Tag)
		for _, obj := range *existing {
			extIdMap[obj.Id] = obj
		}

		for _, obj := range qs {
			if prev, ok := extIdMap[obj.Id]; ok {
				if prev.After(obj.Updated) {
					// log.Printf("Skip saving Tag %d %s\n\n", obj.Id, obj.Name)
					obj.Tag = prev.Tag
				}
				obj.PK = prev.PK
				if !tools.TagsEqual(&obj.Tag, &prev.Tag) {
					changed = true
				}
				Must(tx.Debug().Save(&obj))
			} else {
				changed = true
				Must(tx.Debug().Create(&obj))
			}
		}
		changed = MustOrNotFound(tx.Debug().Where("modified < updated").Delete(Tag{})) || changed
		return nil
	})
	return changed, err
}

func listTags(scope *gorm.DB) Tags {
	qs := NewTags()
	Must(scope.Debug().Find(qs))
	return *qs
}

func (obj Tag) GetName() string {
	return obj.Name
}

func (obj Tag) GetPK() uint {
	return obj.PK
}

func (obj *Tag) ComputeUpdated() {
	obj.Updated = &obj.Tag.At
}
