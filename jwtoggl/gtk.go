package jwtoggl

// #cgo pkg-config: gdk-3.0 gio-2.0 glib-2.0 gobject-2.0 gtk+-3.0
// #cgo CFLAGS: -I /home/tumbler/go/pkg/github.com/gotk3/gotk3/
// #include <gtk/gtk.h>
// #include "gtk_.go.h"
import "C"
import (
	"github.com/gotk3/gotk3/gtk"
	"unsafe"
)

func toCellRenderer(v *gtk.CellRenderer) *C.GtkCellRenderer {
	// native returns a pointer to the underlying GtkCellRenderer.
	if v == nil || v.GObject == nil {
		return nil
	}
	p := unsafe.Pointer(v.GObject)
	return C.toGtkCellRenderer(p)
}

// native() returns a pointer to the underlying GObject as a GtkCellLayout.
func native(v *gtk.CellLayout) *C.GtkCellLayout {
	if v == nil || v.GObject == nil {
		return nil
	}
	p := unsafe.Pointer(v.GObject)
	return C.toGtkCellLayout(p)
}

func gbool(b bool) C.gboolean {
	if b {
		return C.gboolean(1)
	}
	return C.gboolean(0)
}

// PackStart() is a wrapper around gtk_cell_layout_pack_start().
func PackStart(v *gtk.CellLayout, cell *gtk.CellRenderer, expand bool) {
	C.gtk_cell_layout_pack_start(native(v), toCellRenderer(cell),
		gbool(expand))
}

func Clear(v *gtk.CellLayout) {
	C.gtk_cell_layout_clear(native(v))
}
