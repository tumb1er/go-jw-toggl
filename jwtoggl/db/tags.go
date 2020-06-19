package db

import (
	"go-jw-toggl/jwtoggl/gtoggl/gttag"
)

//go:generate bash -c "sed 's/Foo/Tag/g;s/foo/gttag/g' foo.tpl > tags_gen.go"
//go:generate bash -c "sed 's/Foo/Tag/g;s/foo/gttag/g' bar.tpl >> tags_gen.go"

type OnTagsChangedCallback func(gttag.Tags)

type TagsDB struct {
	common
	onTagsChanged OnTagsChangedCallback
}
