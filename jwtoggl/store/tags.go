package store

//go:generate bash -c "sed 's/Foo/Tag/g;s/foo/gttag/g' foo.tpl > tags_gen.go"

import (
	"github.com/jinzhu/gorm"
	"go-jw-toggl/jwtoggl/gtoggl/gttag"
)

type Tag struct {
	Synced
	gttag.Tag
}

type Tags []Tag

type TagManager gorm.DB

func (m *TagManager) List() gttag.Tags {
	return listTags(m.gorm().Order("name ASC"))
}
