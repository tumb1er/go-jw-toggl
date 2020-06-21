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

func (obj Project) GetClientTogglID() uint64 {
	return obj.CId
}

func (obj Project) GetHexColor() string {
	return obj.HexColor
}

type Projects []Project

type ProjectManager gorm.DB

func (m *ProjectManager) List() Projects {
	return listProjects(m.gorm().Order("name ASC"))
}
