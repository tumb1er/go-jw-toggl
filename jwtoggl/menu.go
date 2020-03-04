package jwtoggl

import (
	"go-jw-toggl/jwtoggl/builder"
	"log"
)

func (j *jApp) initMenu(b *builder.Builder) {
	j.menu = b.GetMenu()
	j.menuNew = b.GetMenuItemNew()
	b.ConnectSignals(map[string]interface{}{
		"toggle": j.onToggle,
		"exit":   j.onExit,
	})
}

func (j *jApp) onToggle() {
	log.Println("onToggle")
	j.Call(func() {
		visible := j.window.GetVisible()
		j.window.SetVisible(!visible)
	})
}

func (j *jApp) onExit() {
	log.Println("onExit")
	j.cancel()
	j.Call(j.app.Quit)
}
