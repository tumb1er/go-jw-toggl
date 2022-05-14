package window

import (
	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/gtk"
	"go-jw-toggl/jwtoggl/builder"
	"log"
)

type MainWindow struct {
	window                    *gtk.ApplicationWindow
	xPos, yPos, width, height int
}

func New(b *builder.Builder) (*MainWindow, error) {
	w := &MainWindow{
		window: b.GetMainWindow(),
	}
	return w, nil
}

func (w *MainWindow) ConnectSignals(
	onWindowDelete func() bool,
	onWindowLeave func() bool,
) {
	w.window.Connect("delete-event", onWindowDelete)
	w.window.Connect("focus-out-event", onWindowLeave)
}

func (w *MainWindow) AddToApp(app *gtk.Application) {
	app.AddWindow(w.window)
}
func (w *MainWindow) Hide() {
	w.width = w.window.GetAllocatedWidth()
	w.height = w.window.GetAllocatedHeight()
	w.xPos, w.yPos = w.window.GetPosition()
	w.window.Hide()
}

func (w *MainWindow) Show() {
	if w.xPos*w.yPos*w.width*w.height == 0 {
		display, err := gdk.DisplayGetDefault()
		if err != nil {
			log.Fatal("DefaultDisplay error: ", err)
		}
		window, err := w.window.GetWindow()
		var monitor *gdk.Monitor
		if err != nil {
			monitor, err = display.GetPrimaryMonitor()
		} else {
			monitor, err = display.GetMonitorAtWindow(window)
		}
		if err == nil {
			g := monitor.GetGeometry()
			w.width = g.GetWidth() / 4
			w.height = g.GetHeight()
			w.xPos = g.GetX() + g.GetWidth() - w.width
			w.yPos = g.GetY()
		}

	}

	if !w.window.GetVisible() {
		log.Printf("Resize window to %dx%d\n", w.width, w.height)
		w.window.Resize(w.width, w.height)
		log.Printf("Move window to %dx%d\n", w.xPos, w.yPos)
		w.window.Move(w.xPos, w.yPos)
		w.window.SetVisible(true)
	} else {
		w.window.Present()
	}
}
