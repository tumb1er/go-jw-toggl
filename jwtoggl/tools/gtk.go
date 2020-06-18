package tools

// #cgo pkg-config: gtk+-3.0
// #include <gtk/gtk.h>
// #include "gtk.go.h"
import "C"
import "unsafe"

func Escape(str string) string {
	cstr := C.CString(str)
	defer C.free(unsafe.Pointer(cstr))
	escaped := C.escape_markup(cstr, C.gint(len(str)))
	return C.GoString((*C.char)(escaped))
}
