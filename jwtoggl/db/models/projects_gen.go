package models

import (
	"github.com/jinzhu/gorm"
	"go-jw-toggl/jwtoggl/gtoggl/gtproject"
	"go-jw-toggl/jwtoggl/tools"
	"time"
)

func (obj *Project) BeforeCreate() error {
	return obj.dump(obj.Project)
}
func (obj *Project) BeforeSave() error {
	return obj.dump(obj.Project)
}

func (obj *Project) AfterFind() error {
	return obj.load(&obj.Project)
}

func NewProjects(items ...gtproject.Projects) *Projects {
	if len(items) == 0 {
		result := make(Projects, 0)
		return &result
	}
	result := make(Projects, len(items[0]))
	for i, c := range items[0] {
		result[i] = Project{Project: c}
	}
	return &result
}

func (qs Projects) Sync(db *gorm.DB, ts time.Time) (changed bool, err error) {
	err = db.Transaction(func(tx *gorm.DB) error {
		existing := NewProjects()
		if res := tx.Find(existing); res.Error != nil {
			return res.Error
		}
		extIdMap := make(map[uint64]Project)
		for _, obj := range *existing {
			extIdMap[obj.Id] = obj
		}

		for _, obj := range qs {
			obj.SavedAt = ts
			if prev, ok := extIdMap[obj.Id]; ok {
				obj.PK = prev.PK
				if !tools.ProjectsEqual(&obj.Project, &prev.Project) {
					changed = true
				}
				tx.Save(&obj)
			} else {
				changed = true
				tx.Create(&obj)
			}
		}
		res := tx.Where("saved_at < ?", ts).Delete(Project{})
		if res.Error == nil && res.RowsAffected > 0 {
			changed = true
		}
		return res.Error
	})
	return changed, err
}
