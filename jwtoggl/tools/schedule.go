package tools

import (
	"github.com/gotk3/gotk3/glib"
	"log"
)

func Schedule(f interface{}, params ...interface{}) {
	if _, err := glib.IdleAdd(f, params...); err != nil {
		log.Fatal("IdleAdd error:", err)
	}
}
