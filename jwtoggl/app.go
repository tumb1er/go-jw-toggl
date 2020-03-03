package jwtoggl

import (
	"context"
	"github.com/dawidd6/go-appindicator"
	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"
	"github.com/tumb1er/go-notifier/notifier"
	"go-jw-toggl/jwtoggl/builder"
	"go-jw-toggl/jwtoggl/db"
	"go-jw-toggl/jwtoggl/entry_form"
	"log"
	"os"
)

const (
	appID    = "com.github.tumb1er.go-jw-toggl"
	activate = "activate"
)

type jApp struct {
	app       *gtk.Application
	form      *entry_form.EntryForm
	indicator *appindicator.Indicator
	window    *gtk.ApplicationWindow
	menu      *gtk.Menu
	menuNew   *gtk.MenuItem
	db        DB
	ctx       context.Context
	cancel    context.CancelFunc
	popup     *notifier.Notifier
}

func (j *jApp) Run(args []string) int {
	return j.app.Run(args)
}

func (j *jApp) onActivate() {
	b := builder.NewBuilder()
	var err error
	if j.form, err = entry_form.NewEntryForm(b); err != nil {
		log.Fatal("Form build error:", err)
	}
	j.menu = b.GetMenu()
	j.menuNew = b.GetMenuItemNew()
	j.window = b.GetMainWindow()
	j.indicator = NewIndicator(j.menu, j.menuNew)
	j.initMenu(b)
	j.app.AddWindow(j.window)
	//j.showNotification("Hello", "JW Toggl started")
	go j.StartSync(j.ctx)
}

// NewApp подготавливает основное окно приложения с иконкой в трее и меню.
func NewApp() *jApp {
	app, err := gtk.ApplicationNew(appID, glib.APPLICATION_FLAGS_NONE)
	if err != nil {
		log.Fatal("Could not create GTK application:", err)
	}
	togglDb, err := db.NewDB(os.Getenv("TOGGL_KEY"))
	if err != nil {
		log.Fatal("Could not create DB:", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	popup, err := notifier.NewNotifier(getIconPath(TogglIcon))
	if err != nil {
		log.Fatal("Could not create popup:", err)
	}
	j := &jApp{
		app:    app,
		db:     togglDb,
		ctx:    ctx,
		cancel: cancel,
		popup:  popup,
	}
	_, err = app.Connect(activate, j.onActivate)
	if err != nil {
		log.Fatal("Could not connect activate handle:", err)
	}
	return j
}
