package tools

import (
	"github.com/gotk3/gotk3/glib"
)

func Schedule(f interface{}) {
	glib.IdleAdd(f)
}
