package builder

import (
	"errors"
	"fmt"
	"github.com/gotk3/gotk3/gtk"
	"log"
	"os"
	"path"
	"path/filepath"
)

const (
	mainWindow   = "main_window"
	statusBar    = "status_bar"
	entryHistory = "entry_history"
	historyView  = "history_view"
)

type Builder struct {
	*gtk.Builder
}

func (b *Builder) GetMenu(id string) (*gtk.Menu, error) {
	obj, err := b.GetObject(id)
	if err != nil {
		return nil, err
	}
	if ret, ok := obj.(*gtk.Menu); !ok {
		return nil, fmt.Errorf("menu cast error: %T", obj)
	} else {
		return ret, nil
	}
}

func (b *Builder) GetMainWindow() *gtk.ApplicationWindow {
	obj, err := b.GetObject(mainWindow)
	if err != nil {
		log.Fatal("Could not get entry list window:", err)
	}
	if ret, ok := obj.(*gtk.ApplicationWindow); !ok {
		log.Fatalf("entry list window cast error: %T", obj)
		return nil
	} else {
		return ret
	}
}

func (b *Builder) GetMenuItem(id string) (*gtk.MenuItem, error) {
	obj, err := b.GetObject(id)
	if err != nil {
		return nil, err
	}
	if ret, ok := obj.(*gtk.MenuItem); !ok {
		return nil, fmt.Errorf("MenuItem cast error: %T", obj)
	} else {
		return ret, nil
	}
}

func (b *Builder) GetEntry(id string) (*gtk.Entry, error) {
	obj, err := b.GetObject(id)
	if err != nil {
		log.Fatal("Could not get entry:", err)
	}
	if ret, ok := obj.(*gtk.Entry); !ok {
		return nil, errors.New("entry cast error")
	} else {
		return ret, nil
	}
}

func (b *Builder) GetComboBox(id string) (*gtk.ComboBox, error) {
	obj, err := b.GetObject(id)
	if err != nil {
		log.Fatal("Could not get ComboBox:", err)
	}
	if ret, ok := obj.(*gtk.ComboBox); !ok {
		return nil, errors.New("combobox cast error")
	} else {
		return ret, nil
	}
}

func (b *Builder) GetButton(id string) (*gtk.Button, error) {
	obj, err := b.GetObject(id)
	if err != nil {
		log.Fatal("Could not get Button:", err)
	}
	if ret, ok := obj.(*gtk.Button); !ok {
		return nil, errors.New("button cast error")
	} else {
		return ret, nil
	}
}

func (b *Builder) GetImage(id string) (*gtk.Image, error) {
	obj, err := b.GetObject(id)
	if err != nil {
		log.Fatal("Could not get Image:", err)
	}
	if ret, ok := obj.(*gtk.Image); !ok {
		return nil, errors.New("image cast error")
	} else {
		return ret, nil
	}
}

func (b *Builder) GetListStore(id string) (*gtk.ListStore, error) {
	obj, err := b.GetObject(id)
	if err != nil {
		log.Fatal("Could not get ListStore:", err)
	}
	if ret, ok := obj.(*gtk.ListStore); !ok {
		return nil, errors.New("list store cast error")
	} else {
		return ret, nil
	}
}

func (b *Builder) GetHistoryTreeStore() *gtk.TreeStore {
	obj, err := b.GetObject(entryHistory)
	if err != nil {
		log.Fatal("Could not get TreeStore:", err)
	}
	if ret, ok := obj.(*gtk.TreeStore); !ok {
		log.Fatalf("Tree store cast error: %T\n", obj)
		return nil
	} else {
		return ret
	}
}

func (b *Builder) GetHistoryView() *gtk.TreeView {
	obj, err := b.GetObject(historyView)
	if err != nil {
		log.Fatal("Could not get TreeView:", err)
	}
	if ret, ok := obj.(*gtk.TreeView); !ok {
		log.Fatalf("Tree view cast error: %T\n", obj)
		return nil
	} else {
		return ret
	}
}

func (b *Builder) GetStatusbar() *gtk.Statusbar {
	obj, err := b.GetObject(statusBar)
	if err != nil {
		log.Fatal("Could not get Statusbar:", err)
	}
	if ret, ok := obj.(*gtk.Statusbar); !ok {
		log.Fatalf("statusbar cast error: %T\n", obj)
		return nil
	} else {
		return ret
	}
}

func NewBuilder() *Builder {
	b, err := gtk.BuilderNew()
	if err != nil {
		log.Fatal("Could not create new builder:", err)
	}
	cwd, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		log.Fatal("Could not get cwd:", err)
	}
	glade := path.Join(cwd, "glade/jwtoggl.glade")
	if _, err := os.Stat(glade); err != nil {
		log.Fatal("glade file check error:", err)
	}
	err = b.AddFromFile(glade)
	if err != nil {
		log.Fatal("Could not open glade file:", err)
	}
	return &Builder{b}
}
