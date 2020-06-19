package store

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

func (m *ProjectManager) Sync(items gtproject.Projects, ts time.Time, wid uint64) (bool, error) {
	filtered := make(gtproject.Projects, 0, len(items))
	for _, item := range items {
		if item.WId == wid {
			filtered = append(filtered, item)
		}
	}
	qs := NewProjects(filtered)
	return qs.Sync(m.gorm(), ts)
}

func (m *ProjectManager) gorm() *gorm.DB {
	return (*gorm.DB)(m)
}

func NewProjects(items ...gtproject.Projects) *Projects {
	if len(items) == 0 {
		result := make(Projects, 0)
		return &result
	}
	result := make(Projects, len(items[0]))
	for i, item := range items[0] {
		result[i] = Project{Project: item}
	}
	return &result
}

func (qs Projects) Sync(db *gorm.DB, ts time.Time) (changed bool, err error) {
	err = db.Transaction(func(tx *gorm.DB) error {
		existing := NewProjects()
		Must(tx.Find(existing))
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
				Must(tx.Save(&obj))
			} else {
				changed = true
				Must(tx.Create(&obj))
			}
		}
		changed = MustOrNotFound(tx.Where("saved_at < ?", ts).Delete(Project{})) || changed
		return nil
	})
	return changed, err
}

func listProjects(scope *gorm.DB) gtproject.Projects {
	qs := NewProjects()
	length := Must(scope.Find(qs))
	result := make(gtproject.Projects, length)
	for i, obj := range *qs {
		result[i] = obj.Project
	}
	return result
}
