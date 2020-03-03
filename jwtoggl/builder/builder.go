package builder

import (
	"errors"
	"github.com/gotk3/gotk3/gtk"
	"log"
	"os"
	"path"
)

const (
	trayMenuId = "tray_menu"
	menuNew    = "menu_new"
	mainWindow = "main_window"
)

type Builder struct {
	*gtk.Builder
}

func (b *Builder) GetMenu() *gtk.Menu {
	obj, err := b.GetObject(trayMenuId)
	if err != nil {
		log.Fatal("Could not get tray menu:", err)
	}
	if ret, ok := obj.(*gtk.Menu); !ok {
		log.Fatal("Tray menu cast error: %T", obj)
		return nil
	} else {
		return ret
	}
}

func (b *Builder) GetMainWindow() *gtk.ApplicationWindow {
	obj, err := b.GetObject(mainWindow)
	if err != nil {
		log.Fatal("Could not get entry list window:", err)
	}
	if ret, ok := obj.(*gtk.ApplicationWindow); !ok {
		log.Fatal("entry list window cast error: %T", obj)
		return nil
	} else {
		return ret
	}
}

func (b *Builder) GetMenuItemNew() *gtk.MenuItem {
	obj, err := b.GetObject(menuNew)
	if err != nil {
		log.Fatal("Could not get new menu item:", err)
	}
	if ret, ok := obj.(*gtk.MenuItem); !ok {
		log.Fatal("MenuItem cast error: %T", obj)
		return nil
	} else {
		return ret
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

func NewBuilder() *Builder {
	b, err := gtk.BuilderNew()
	if err != nil {
		log.Fatal("Could not create new builder:", err)
	}
	cwd, err := os.Getwd()
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
