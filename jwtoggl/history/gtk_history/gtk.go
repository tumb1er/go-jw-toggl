package gtk_history

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

// Private gtk.* reexport
var nilPtrErr = errors.New("cgo returned unexpected nil pointer")

func gbool(b bool) C.gboolean {
	if b {
		return C.gboolean(1)
	}
	return C.gboolean(0)
}

// CellArea is a representation of GTK's GtkCellArea.
type CellArea struct {
	glib.InitiallyUnowned
}

// ICellArea is an interface type implemented by all structs
// embedding a CellArea.  It is meant to be used as an argument type
// for wrapper functions that wrap around a C GTK function taking a
// GtkCellArea.
type ICellArea interface {
	toCellArea() *C.GtkCellArea
}

// native returns a pointer to the underlying GtkCellArea.
func (v *CellArea) native() *C.GtkCellArea {
	if v == nil || v.GObject == nil {
		return nil
	}
	p := unsafe.Pointer(v.GObject)
	return C.toGtkCellArea(p)
}

func (v *CellArea) toCellArea() *C.GtkCellArea {
	if v == nil {
		return nil
	}
	return v.native()
}

func marshalCellArea(p uintptr) (interface{}, error) {
	c := C.g_value_get_object((*C.GValue)(unsafe.Pointer(p)))
	obj := glib.Take(unsafe.Pointer(c))
	return wrapCellArea(obj), nil
}

func wrapCellArea(obj *glib.Object) *CellArea {
	return &CellArea{glib.InitiallyUnowned{obj}}
}

/*
 * GtkCellAreaBox
 */

// CellAreaBox is a representation of GTK's GtkCellAreaBox.
type CellAreaBox struct {
	CellArea
}

// native returns a pointer to the underlying GtkCellAreaBox.
func (v *CellAreaBox) native() *C.GtkCellAreaBox {
	if v == nil || v.GObject == nil {
		return nil
	}
	p := unsafe.Pointer(v.GObject)
	return C.toGtkCellAreaBox(p)
}

func marshalCellAreaBox(p uintptr) (interface{}, error) {
	c := C.g_value_get_object((*C.GValue)(unsafe.Pointer(p)))
	obj := glib.Take(unsafe.Pointer(c))
	return wrapCellAreaBox(obj), nil
}

func wrapCellAreaBox(obj *glib.Object) *CellAreaBox {
	return &CellAreaBox{CellArea{glib.InitiallyUnowned{obj}}}
}

// CellAreaBoxNew is a wrapper around gtk_cell_area_box_new().
func CellAreaBoxNew() (*CellAreaBox, error) {
	c := C.gtk_cell_area_box_new()
	if c == nil {
		return nil, nilPtrErr
	}
	obj := glib.Take(unsafe.Pointer(c))
	return wrapCellAreaBox(obj), nil
}

// PackStart() is a wrapper around gtk_cell_area_box_pack_start().
func (v *CellAreaBox) PackStart(cell ICellRenderer, expand, align, fixed bool) {
	C.gtk_cell_area_box_pack_start(v.native(), cell.toCellRenderer(), gbool(expand), gbool(align), gbool(fixed))
}

/*
 * GtkTreeViewColumn extensions
 */

// TreeViewColumnNew() is a wrapper around gtk_tree_view_column_new().
func TreeViewColumnNewWithArea(area ICellArea) (*gtk.TreeViewColumn, error) {
	c := C.gtk_tree_view_column_new_with_area(area.toCellArea())
	if c == nil {
		return nil, nilPtrErr
	}
	return wrapTreeViewColumn(glib.Take(unsafe.Pointer(c))), nil
}

func wrapTreeViewColumn(obj *glib.Object) *gtk.TreeViewColumn {
	return &gtk.TreeViewColumn{glib.InitiallyUnowned{obj}}
}

// CellRenderer is a representation of GTK's GtkCellRenderer.
type CellRenderer struct {
	glib.InitiallyUnowned
}

// ICellRenderer is an interface type implemented by all structs
// embedding a CellRenderer.  It is meant to be used as an argument type
// for wrapper functions that wrap around a C GTK function taking a
// GtkCellRenderer.
type ICellRenderer interface {
	toCellRenderer() *C.GtkCellRenderer
}

// native returns a pointer to the underlying GtkCellRenderer.
func (v *CellRenderer) native() *C.GtkCellRenderer {
	if v == nil || v.GObject == nil {
		return nil
	}
	p := unsafe.Pointer(v.GObject)
	return C.toGtkCellRenderer(p)
}

func (v *CellRenderer) toCellRenderer() *C.GtkCellRenderer {
	if v == nil {
		return nil
	}
	return v.native()
}

func marshalCellRenderer(p uintptr) (interface{}, error) {
	c := C.g_value_get_object((*C.GValue)(unsafe.Pointer(p)))
	obj := glib.Take(unsafe.Pointer(c))
	return wrapCellRenderer(obj), nil
}

func wrapCellRenderer(obj *glib.Object) *CellRenderer {
	return &CellRenderer{glib.InitiallyUnowned{obj}}
}

/*
 * GtkCellRendererText
 */

// CellRendererText is a representation of GTK's GtkCellRendererText.
type CellRendererText struct {
	CellRenderer
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
	return &CellRendererText{CellRenderer{glib.InitiallyUnowned{obj}}}
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

/*
 * Helpers
 */

func SetProperties(obj *glib.Object, values map[string]interface{}) error {
	for name, value := range values {
		if err := obj.SetProperty(name, value); err != nil {
			return err
		}
	}
	return nil
}

func AddAttributes(column *gtk.TreeViewColumn, r *CellRendererText, attributes map[string]int) {
	renderer := gtk.CellRenderer{InitiallyUnowned: r.InitiallyUnowned}
	for name, columnId := range attributes {
		column.AddAttribute(&renderer, name, columnId)
	}
}
