package store

//go:generate bash -c "sed 's/Foo/Project/g;s/foo/gtproject/g' foo.tpl > projects_gen.go"

import (
	"github.com/jinzhu/gorm"
	"go-jw-toggl/jwtoggl/gtoggl/gtproject"
)

type Project struct {
	Named
	gtproject.Project `gorm:"-"`
}

type Projects []Project

type ProjectManager gorm.DB

func (m *ProjectManager) List() gtproject.Projects {
	return listProjects(m.gorm().Order("name ASC"))
}
