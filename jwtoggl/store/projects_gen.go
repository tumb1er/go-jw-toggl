package store

// gtproject.tpl предоставляет методы для работы с моделями объектов Toggl

import (
	"github.com/jinzhu/gorm"
	"go-jw-toggl/jwtoggl/dto"
	"go-jw-toggl/jwtoggl/gtoggl/gtproject"
	"go-jw-toggl/jwtoggl/tools"
	"time"
)

func (m *ProjectManager) Sync(items gtproject.Projects, wid uint64) (bool, error) {
	filtered := make(gtproject.Projects, 0, len(items))
	for _, item := range items {
		if item.WId == wid {
			filtered = append(filtered, item)
		}
	}
	qs := NewProjects(filtered)
	return qs.Sync(m.gorm())
}

func (m *ProjectManager) gorm() *gorm.DB {
	return (*gorm.DB)(m).Set("gorm:auto_preload", true)
}

func NewProjects(items ...gtproject.Projects) *Projects {
	if len(items) == 0 {
		result := make(Projects, 0)
		return &result
	}
	result := make(Projects, len(items[0]))
	for i, item := range items[0] {
		result[i] = Project{Project: item}
		result[i].ComputeUpdated()
	}
	return &result
}

func (qs Projects) DTO() []dto.Project {
	result := make([]dto.Project, len(qs))
	for i, c := range qs {
		result[i] = c
	}
	return result
}

func (qs Projects) Sync(db *gorm.DB) (changed bool, err error) {
	if err = qs.Preload(db); err != nil {
		return
	}
	err = db.Transaction(func(tx *gorm.DB) error {
		ts := time.Now().Local()
		existing := NewProjects()
		Must(tx.Find(existing))
		extIdMap := make(map[uint64]Project)
		for _, obj := range *existing {
			extIdMap[obj.Id] = obj
		}

		for _, obj := range qs {
			obj.Fetched = &ts
			if prev, ok := extIdMap[obj.Id]; ok {
				if !obj.After(prev.Modified) {
					// log.Printf("Skip saving Project %d %s\n\n", obj.Id, obj.Name)
					obj.Project = prev.Project
					obj.Modified = prev.Modified
					obj.Deleted = prev.Deleted
				} else {
					obj.Modified = obj.Updated
				}
				obj.PK = prev.PK
				if !tools.ProjectsEqual(&obj.Project, &prev.Project) {
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
				Delete(Project{}))
		changed = deleted || changed
		return nil
	})
	return changed, err
}

func listProjects(scope *gorm.DB) Projects {
	qs := NewProjects()
	Must(scope.Find(qs))
	return *qs
}

func (obj Project) GetName() string {
	return obj.Name
}

func (obj Project) GetPK() uint {
	return obj.PK
}

func (obj *Project) ComputeUpdated() {
	if !obj.Project.At.IsZero() {
		obj.Updated = &obj.Project.At
	}
}
