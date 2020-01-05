package jwtoggl

import (
	"github.com/dawidd6/go-appindicator"
	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"
	"log"
)

const (
	appID    = "com.github.tumb1er.go-jw-toggl"
	activate = "activate"
)

type jApp struct {
	app       *gtk.Application
	indicator *appindicator.Indicator
	window    *gtk.ApplicationWindow
	menu      *gtk.Menu
	showHide  *gtk.MenuItem
}

func (j *jApp) Run(args []string) int {
	return j.app.Run(args)
}

func (j *jApp) showNotification(title, body string) {
	icon := glib.NotificationNew(title)
	icon.SetBody(body)
	j.app.SendNotification(trayId, icon)

}

func (j *jApp) onToggle() {
	log.Println("onToggle")
	visible := j.window.GetVisible()
	j.window.SetVisible(!visible)
	if visible {
		j.showHide.SetLabel("Show")
	} else {
		j.showHide.SetLabel("Hide")
	}
}

func (j *jApp) onExit() {
	log.Println("onExit")
	j.app.Quit()
}

func (j *jApp) onActivate() {
	builder := NewBuilder()
	j.menu = builder.GetMenu()
	j.showHide = builder.GetShowMenuItem()
	j.window = builder.GetEntryListWindow()
	j.indicator = NewIndicator(j.menu, j.showHide)
	j.initMenu(builder)
	j.app.AddWindow(j.window)
	j.showNotification("Hello", "JW Toggl started")
}

// NewApp подготавливает основное окно приложения с иконкой в трее и меню.
func NewApp() *jApp {
	app, err := gtk.ApplicationNew(appID, glib.APPLICATION_FLAGS_NONE)
	if err != nil {
		log.Fatal("Could not create GTK application:", err)
	}
	j := &jApp{
		app: app,
	}
	_, err = app.Connect(activate, j.onActivate)
	if err != nil {
		log.Fatal("Could not connect activate handle:", err)
	}
	return j
}
