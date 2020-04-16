package models

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

func ListProjects(db *gorm.DB) gtproject.Projects {
	qs := NewProjects()
	length := Must(db.Order("name ASC").Find(qs))
	result := make(gtproject.Projects, length)
	for i, c := range *qs {
		result[i] = c.Project
	}
	return result
}
