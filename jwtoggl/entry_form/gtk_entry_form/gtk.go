package gtk_entry_form

// #cgo pkg-config: gdk-3.0 gio-2.0 glib-2.0 gobject-2.0 gtk+-3.0
// #include <gtk/gtk.h>
// #include "gtk.go.h"
import "C"
import (
	"github.com/gotk3/gotk3/gtk"
	"unsafe"
)

// native() returns a pointer to the underlying GObject as a GtkCellLayout.
func native(v *gtk.CellLayout) *C.GtkCellLayout {
	if v == nil || v.GObject == nil {
		return nil
	}
	p := unsafe.Pointer(v.GObject)
	return C.toGtkCellLayout(p)
}

func CellLayoutClear(v *gtk.CellLayout) {
	C.gtk_cell_layout_clear(native(v))
}
