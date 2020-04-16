package models

//go:generate bash -c "sed 's/Foo/Tag/g;s/foo/gttag/g' foo.tpl > tags_gen.go"

import (
	"github.com/jinzhu/gorm"
	"go-jw-toggl/jwtoggl/gtoggl/gttag"
)

type Tag struct {
	Named
	gttag.Tag `gorm:"-"`
}

type Tags []Tag

func ListTags(db *gorm.DB) gttag.Tags {
	qs := NewTags()
	length := Must(db.Order("name ASC").Find(qs))
	result := make(gttag.Tags, length)
	for i, c := range *qs {
		result[i] = c.Tag
	}
	return result
}
