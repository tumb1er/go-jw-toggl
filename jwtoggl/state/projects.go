package state

//go:generate bash -c "sed 's/Foo/Project/g;s/foo/gtproject/g' foo.tpl > projects_gen.go"
//go:generate bash -c "sed 's/Foo/Project/g;s/foo/gtproject/g' bar.tpl >> projects_gen.go"

import (
	"go-jw-toggl/jwtoggl/gtoggl/gtproject"
)

type OnProjectsChangedCallback func(gtproject.Projects)

type ProjectsDB struct {
	common
	onProjectsChanged OnProjectsChangedCallback
}
