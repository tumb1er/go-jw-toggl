package jwtoggl

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

type builder struct {
	builder *gtk.Builder
}

func (b *builder) GetMenu() *gtk.Menu {
	obj, err := b.builder.GetObject(trayMenuId)
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

func (b *builder) GetMainWindow() *gtk.ApplicationWindow {
	obj, err := b.builder.GetObject(mainWindow)
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

func (b *builder) GetMenuItemNew() *gtk.MenuItem {
	obj, err := b.builder.GetObject(menuNew)
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

func (b *builder) GetEntry(id string) (*gtk.Entry, error) {
	obj, err := b.builder.GetObject(id)
	if err != nil {
		log.Fatal("Could not get entry:", err)
	}
	if ret, ok := obj.(*gtk.Entry); !ok {
		return nil, errors.New("entry cast error")
	} else {
		return ret, nil
	}
}

func (b *builder) GetComboBox(id string) (*gtk.ComboBox, error) {
	obj, err := b.builder.GetObject(id)
	if err != nil {
		log.Fatal("Could not get ComboBox:", err)
	}
	if ret, ok := obj.(*gtk.ComboBox); !ok {
		return nil, errors.New("combobox cast error")
	} else {
		return ret, nil
	}
}

func (b *builder) GetButton(id string) (*gtk.Button, error) {
	obj, err := b.builder.GetObject(id)
	if err != nil {
		log.Fatal("Could not get Button:", err)
	}
	if ret, ok := obj.(*gtk.Button); !ok {
		return nil, errors.New("button cast error")
	} else {
		return ret, nil
	}
}

func (b *builder) GetImage(id string) (*gtk.Image, error) {
	obj, err := b.builder.GetObject(id)
	if err != nil {
		log.Fatal("Could not get Image:", err)
	}
	if ret, ok := obj.(*gtk.Image); !ok {
		return nil, errors.New("image cast error")
	} else {
		return ret, nil
	}
}
func (b *builder) GetListStore(id string) (*gtk.ListStore, error) {
	obj, err := b.builder.GetObject(id)
	if err != nil {
		log.Fatal("Could not get ListStore:", err)
	}
	if ret, ok := obj.(*gtk.ListStore); !ok {
		return nil, errors.New("list store cast error")
	} else {
		return ret, nil
	}
}

func NewBuilder() *builder {
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
	return &builder{builder: b}
}
