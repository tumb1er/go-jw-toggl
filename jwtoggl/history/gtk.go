package history

// #cgo pkg-config: gdk-3.0 gio-2.0 glib-2.0 gobject-2.0 gtk+-3.0
// #include <gtk/gtk.h>
// #include "gtk.go.h"
import "C"
import (
	"errors"
	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"
	"unsafe"
)

var nilPtrErr = errors.New("XXX cgo returned unexpected nil pointer")

/*
 * GtkCellRendererText
 */

// CellRendererText is a representation of GTK's GtkCellRendererText.
type CellRendererText struct {
	gtk.CellRenderer
}

// native returns a pointer to the underlying GtkCellRendererText.
func (v *CellRendererText) native() *C.GtkCellRendererText {
	if v == nil || v.GObject == nil {
		return nil
	}
	p := unsafe.Pointer(v.GObject)
	return C.toGtkCellRendererText(p)
}

func marshalCellRendererText(p uintptr) (interface{}, error) {
	c := C.g_value_get_object((*C.GValue)(unsafe.Pointer(p)))
	obj := glib.Take(unsafe.Pointer(c))
	return wrapCellRendererText(obj), nil
}

func wrapCellRendererText(obj *glib.Object) *CellRendererText {
	return &CellRendererText{
		gtk.CellRenderer{
			InitiallyUnowned: glib.InitiallyUnowned{Object: obj}}}
}

// CellRendererTextNew is a wrapper around gtk_cell_renderer_text_new().
func CellRendererTextNew() (*CellRendererText, error) {
	c := C.gtk_cell_renderer_text_new()
	if c == nil {
		return nil, nilPtrErr
	}
	obj := glib.Take(unsafe.Pointer(c))
	return wrapCellRendererText(obj), nil
}
