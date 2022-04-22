package store

// foo.tpl предоставляет методы для работы с моделями объектов Toggl

import (
	"github.com/jinzhu/gorm"
	"go-jw-toggl/jwtoggl/dto"
	"go-jw-toggl/jwtoggl/gtoggl"
	"go-jw-toggl/jwtoggl/gtoggl/foo"
	"time"
)

func (m *FooManager) Sync(items foo.Foos, wid uint64) (bool, error) {
	filtered := make(foo.Foos, 0, len(items))
	for _, item := range items {
		if item.WId == wid {
			filtered = append(filtered, item)
		}
	}
	qs := NewFoos(filtered)
	return qs.Sync(m.gorm())
}

func (m *FooManager) gorm() *gorm.DB {
	return (*gorm.DB)(m).Set("gorm:auto_preload", true)
}

func NewFoos(items ...foo.Foos) *Foos {
	if len(items) == 0 {
		result := make(Foos, 0)
		return &result
	}
	result := make(Foos, len(items[0]))
	for i, item := range items[0] {
		result[i] = Foo{Foo: item}
		result[i].ComputeUpdated()
	}
	return &result
}

func (qs Foos) DTO() []dto.Foo {
	result := make([]dto.Foo, len(qs))
	for i, c := range qs {
		result[i] = c
	}
	return result
}

func (qs Foos) Sync(db *gorm.DB) (changed bool, err error) {
	if err = qs.Preload(db); err != nil {
		return
	}
	err = db.Transaction(func(tx *gorm.DB) error {
		ts := time.Now().Local()
		existing := NewFoos()
		Must(tx.Find(existing))
		extIdMap := make(map[uint64]Foo)
		for _, obj := range *existing {
			extIdMap[obj.Id] = obj
		}

		for _, obj := range qs {
			obj.Fetched = &ts
			if prev, ok := extIdMap[obj.Id]; ok {
				if !obj.After(prev.Modified) {
					// log.Printf("Skip saving Foo %d %s\n\n", obj.Id, obj.Name)
					obj.Foo = prev.Foo
					obj.Modified = prev.Modified
					obj.Deleted = prev.Deleted
				} else {
					obj.Modified = obj.Updated
				}
				obj.PK = prev.PK
				if !gtoggl.FoosEqual(&obj.Foo, &prev.Foo) {
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
				Delete(Foo{}))
		changed = deleted || changed
		return nil
	})
	return changed, err
}

func listFoos(scope *gorm.DB) Foos {
	qs := NewFoos()
	Must(scope.Find(qs))
	return *qs
}

func (obj Foo) GetName() string {
	return obj.Name
}

func (obj Foo) GetPK() uint {
	return obj.PK
}

func (obj *Foo) ComputeUpdated() {
	if !obj.Foo.At.IsZero() {
		obj.Updated = &obj.Foo.At
	}
}
