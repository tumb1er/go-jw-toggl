package jwtoggl

import (
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
