package jwtoggl

import (
	"go-jw-toggl/jwtoggl/builder"
	"log"
)

func (j *jApp) initMenu(b *builder.Builder) {
	b.ConnectSignals(map[string]interface{}{
		"toggle": j.onToggle,
		"exit":   j.onExit,
	})
}

func (j *jApp) onToggle() {
	log.Println("onToggle")
	visible := j.window.GetVisible()
	if visible {
		j.form.Validate()
	}
	j.window.SetVisible(!visible)
}

func (j *jApp) onExit() {
	log.Println("onExit")
	j.cancel()
	j.app.Quit()
}
