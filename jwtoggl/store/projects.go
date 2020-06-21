package store

//go:generate bash -c "sed 's/Foo/Project/g;s/foo/gtproject/g' foo.tpl > projects_gen.go"

import (
	"github.com/jinzhu/gorm"
	"go-jw-toggl/jwtoggl/gtoggl/gtproject"
)

type Project struct {
	Synced
	gtproject.Project
}

func (p Project) GetClientTogglID() uint64 {
	return p.CId
}

func (p Project) GetTogglID() uint64 {
	return p.Id
}

func (p Project) GetName() string {
	return p.Name
}

func (p Project) GetHexColor() string {
	return p.HexColor
}

type Projects []Project

type ProjectManager gorm.DB

func (m *ProjectManager) List() Projects {
	return listProjects(m.gorm().Order("name ASC"))
}
