package store

// gttag.tpl предоставляет методы для работы с моделями объектов Toggl

import (
	"go-jw-toggl/jwtoggl/dto"
	"go-jw-toggl/jwtoggl/gtoggl"
	"go-jw-toggl/jwtoggl/gtoggl/gttag"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"time"
)

func (m *TagManager) Sync(items gttag.Tags, wid uint64) (bool, error) {
	filtered := make(gttag.Tags, 0, len(items))
	for _, item := range items {
		if item.WId == wid {
			filtered = append(filtered, item)
		}
	}
	qs := NewTags(filtered)
	return qs.Sync(m.gorm())
}

func (m *TagManager) gorm() *gorm.DB {
	return (*gorm.DB)(m).Preload(clause.Associations).Session(&gorm.Session{})
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

func (qs Tags) Sync(db *gorm.DB) (changed bool, err error) {
	if err = qs.Preload(db); err != nil {
		return
	}
	err = db.Transaction(func(tx *gorm.DB) error {
		ts := time.Now().Local()
		existing := NewTags()
		Must(tx.Find(existing))
		extIdMap := make(map[uint64]Tag)
		for _, obj := range *existing {
			extIdMap[obj.Id] = obj
		}

		for _, obj := range qs {
			obj.Fetched = &ts
			if prev, ok := extIdMap[obj.Id]; ok {
				if !obj.After(prev.Modified) {
					// log.Printf("Skip saving Tag %d %s\n\n", obj.Id, obj.Name)
					obj.Tag = prev.Tag
					obj.Modified = prev.Modified
					obj.Deleted = prev.Deleted
				} else {
					obj.Modified = obj.Updated
				}
				obj.PK = prev.PK
				if !gtoggl.TagsEqual(&obj.Tag, &prev.Tag) {
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
				Delete(Tag{}))
		changed = deleted || changed
		return nil
	})
	return changed, err
}

func listTags(scope *gorm.DB) Tags {
	qs := NewTags()
	Must(scope.Find(qs))
	return *qs
}

func (obj Tag) GetName() string {
	return obj.Name
}

func (obj Tag) GetPK() uint {
	return obj.PK
}

func (obj *Tag) ComputeUpdated() {
	if !obj.Tag.At.IsZero() {
		obj.Updated = &obj.Tag.At
	}
}
