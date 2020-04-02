package tools

import (
	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"
)

func SetProperties(obj *glib.Object, values map[string]interface{}) error {
	for name, value := range values {
		if err := obj.SetProperty(name, value); err != nil {
			return err
		}
	}
	return nil
}

func AddAttributes(column *gtk.TreeViewColumn, r gtk.ICellRenderer, attributes map[string]int) {
	for name, columnId := range attributes {
		column.AddAttribute(r, name, columnId)
	}
}
