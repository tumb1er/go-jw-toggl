package jwtoggl

import (
	"context"
	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"
	"go-jw-toggl/jwtoggl/builder"
	"go-jw-toggl/jwtoggl/db"
	"go-jw-toggl/jwtoggl/entry_form"
	"go-jw-toggl/jwtoggl/gtoggl/gttimentry"
	"go-jw-toggl/jwtoggl/history"
	"go-jw-toggl/jwtoggl/status"
	"go-jw-toggl/jwtoggl/tray"
	"go-jw-toggl/jwtoggl/window"
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
	entries   *history.History
	tray      *tray.Tray
	window    *window.MainWindow
	statusBar *status.Status
	db        DB
	ctx       context.Context
	cancel    context.CancelFunc
}

func (j *jApp) Run(args []string) int {
	return j.app.Run(args)
}

func (j *jApp) onActivate() {
	b := builder.NewBuilder()
	j.initForm(b)
	j.initTray(b)
	j.initWindow(b)
	j.initHistory(b)
	j.initStatusBar(b)
	go j.StartSync(j.ctx)
}

func (j *jApp) initWindow(b *builder.Builder) {
	var err error
	if j.window, err = window.New(b); err != nil {
		log.Fatal(err)
	}
	if err = j.window.ConnectSignals(j.onWindowDelete); err != nil {
		log.Fatal(err)
	}
	j.window.AddToApp(j.app)

}

func (j *jApp) initForm(b *builder.Builder) {
	var err error
	if j.form, err = entry_form.NewEntryForm(b); err != nil {
		log.Fatal("Form build error:", err)
	}
	j.form.OnNewEntryCreated(j.onNewEntryCreated)
	j.form.OnEntryStopped(j.onEntryStopped)
	j.form.OnEntryChanged(j.onEntryChanged)
}

func (j *jApp) onNewEntryCreated(entry *gttimeentry.TimeEntry) {
	j.handleError(j.db.CreateNewEntry(entry), "Entry creation error")
}

func (j *jApp) onEntryStopped(entry *gttimeentry.TimeEntry) {
	j.handleError(j.db.StopEntry(entry), "Entry stop error")
}

func (j *jApp) onEntryChanged(entry *gttimeentry.TimeEntry) {
	j.handleError(j.db.UpdateCurrentEntry(entry), "Entry save error")
}

func (j *jApp) onWindowDelete() bool {
	j.window.Hide()
	return true
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
	j := &jApp{
		app:    app,
		db:     togglDb,
		ctx:    ctx,
		cancel: cancel,
	}
	_, err = app.Connect(activate, j.onActivate)
	if err != nil {
		log.Fatal("Could not connect activate handle:", err)
	}
	return j
}
