package state

import (
	"go-jw-toggl/jwtoggl/dto"
)

//go:generate bash -c "sed 's/Foo/Tag/g;s/foo/gttag/g' foo.tpl > tags_gen.go"
//go:generate bash -c "sed 's/Foo/Tag/g;s/foo/gttag/g' bar.tpl >> tags_gen.go"

type OnTagsChangedCallback func(dto.Tags)

type TagsDB struct {
	common
	onTagsChanged OnTagsChangedCallback
}
