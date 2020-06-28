package store

// foo.tpl предоставляет методы для работы с моделями объектов Toggl

import (
	"github.com/jinzhu/gorm"
	"go-jw-toggl/jwtoggl/dto"
	"go-jw-toggl/jwtoggl/gtoggl/foo"
	"go-jw-toggl/jwtoggl/tools"
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
	err = db.Transaction(func(tx *gorm.DB) error {
		existing := NewFoos()
		Must(tx.Find(existing))
		extIdMap := make(map[uint64]Foo)
		for _, obj := range *existing {
			extIdMap[obj.Id] = obj
		}

		for _, obj := range qs {
			if prev, ok := extIdMap[obj.Id]; ok {
				if prev.After(obj.Updated) {
					// log.Printf("Skip saving Foo %d %s\n\n", obj.Id, obj.Name)
					obj.Foo = prev.Foo
				}
				obj.PK = prev.PK
				if !tools.FoosEqual(&obj.Foo, &prev.Foo) {
					changed = true
				}
				Must(tx.Save(&obj))
			} else {
				changed = true
				Must(tx.Create(&obj))
			}
		}
		changed = MustOrNotFound(tx.Where("modified < updated").Delete(Foo{})) || changed
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
    obj.Updated = &obj.Foo.At
}