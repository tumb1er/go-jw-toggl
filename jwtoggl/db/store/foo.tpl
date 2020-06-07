package store

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

func (m *FooManager) Sync(items foo.Foos, ts time.Time) (bool, error) {
	qs := NewFoos(items)
	return qs.Sync(m.gorm(), ts)
}

func (m *FooManager) gorm() *gorm.DB {
	return (*gorm.DB)(m)
}

func NewFoos(items ...foo.Foos) *Foos {
	if len(items) == 0 {
		result := make(Foos, 0)
		return &result
	}
	result := make(Foos, len(items[0]))
	for i, item := range items[0] {
		result[i] = Foo{Foo: item}
	}
	return &result
}

func (qs Foos) Sync(db *gorm.DB, ts time.Time) (changed bool, err error) {
	err = db.Transaction(func(tx *gorm.DB) error {
		existing := NewFoos()
		Must(tx.Find(existing))
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
				Must(tx.Save(&obj))
			} else {
				changed = true
				Must(tx.Create(&obj))
			}
		}
		changed = MustOrNotFound(tx.Where("saved_at < ?", ts).Delete(Foo{})) || changed
		return nil
	})
	return changed, err
}

func listFoos(scope *gorm.DB) foo.Foos {
	qs := NewFoos()
	length := Must(scope.Find(qs))
	result := make(foo.Foos, length)
	for i, obj := range *qs {
		result[i] = obj.Foo
	}
	return result
}
