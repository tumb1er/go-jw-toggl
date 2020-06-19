package store

//#go:generate bash -c "sed 's/Foo/Tag/g;s/foo/gttag/g' foo.tpl > tags_gen.go"

import (
	"github.com/jinzhu/gorm"
	"go-jw-toggl/jwtoggl/dto"
	"go-jw-toggl/jwtoggl/gtoggl/gttag"
)

type Tag struct {
	Synced
	gttag.Tag
}

type Tags []Tag

func (qs Tags) DTO() []dto.Tag {
	result := make([]dto.Tag, len(qs))
	for i, t := range qs {
		result[i] = t
	}
	return result
}

type TagManager gorm.DB

func (t Tag) GetName() string {
	return t.Name
}

func (m *TagManager) List() Tags {
	return listTags(m.gorm().Order("name ASC"))
}
