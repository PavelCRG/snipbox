package bridge

/*
#cgo CFLAGS: -I${SRCDIR}/../../../engine/include
#cgo !windows LDFLAGS: -L${SRCDIR}/../../../engine/target/release -lsnipbox_engine
#cgo windows LDFLAGS: -L${SRCDIR}/../../../engine/target/x86_64-pc-windows-gnu/release -lsnipbox_engine -lntdll
#include <stdlib.h>
#include "snipbox_engine.h"
*/
import "C"
import (
	"encoding/json"
	"errors"
	"unsafe"
)

func call(fn func(*C.char) *C.char, arg string) (string, error) {
	in := C.CString(arg)
	defer C.free(unsafe.Pointer(in))
	out := fn(in)
	if out == nil {
		return "", errors.New("engine returned no result")
	}
	defer C.snipbox_free_string(out)
	return checked(C.GoString(out))
}

func checked(value string) (string, error) {
	var response struct {
		Error string `json:"error"`
	}
	if json.Unmarshal([]byte(value), &response) == nil && response.Error != "" {
		return "", errors.New(response.Error)
	}
	return value, nil
}
func Search(commands any, query string) (string, error) {
	b, err := json.Marshal(commands)
	if err != nil {
		return "", err
	}
	input := C.CString(string(b))
	q := C.CString(query)
	defer C.free(unsafe.Pointer(input))
	defer C.free(unsafe.Pointer(q))
	out := C.snipbox_search(input, q)
	if out == nil {
		return "", errors.New("engine returned no result")
	}
	defer C.snipbox_free_string(out)
	return checked(C.GoString(out))
}
func Parse(command string) (string, error) {
	return call(func(s *C.char) *C.char { return C.snipbox_parse_template(s) }, command)
}
func Render(command string, values map[string]string) (string, error) {
	b, err := json.Marshal(values)
	if err != nil {
		return "", err
	}
	c := C.CString(command)
	v := C.CString(string(b))
	defer C.free(unsafe.Pointer(c))
	defer C.free(unsafe.Pointer(v))
	out := C.snipbox_render_template(c, v)
	if out == nil {
		return "", errors.New("engine returned no result")
	}
	defer C.snipbox_free_string(out)
	return checked(C.GoString(out))
}
