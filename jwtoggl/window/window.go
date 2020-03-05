package window

import (
	"github.com/gotk3/gotk3/gtk"
	"go-jw-toggl/jwtoggl/builder"
)

type MainWindow struct {
	window *gtk.ApplicationWindow
}

func New(b *builder.Builder) (*MainWindow, error) {
	w := &MainWindow{
		window: b.GetMainWindow(),
	}
	return w, nil
}

func (w *MainWindow) ConnectSignals(
	onWindowDelete func() bool,
) error {
	if _, err := w.window.Connect("delete-event", onWindowDelete); err != nil {
		return err
	}
	return nil
}

func (w *MainWindow) AddToApp(app *gtk.Application) {
	app.AddWindow(w.window)
}
func (w *MainWindow) Hide() {
	w.window.Hide()
}

func (w *MainWindow) Show() {
	if !w.window.GetVisible() {
		w.window.SetVisible(true)
	} else {
		w.window.Present()
	}
}
