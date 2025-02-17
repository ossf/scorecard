package foo

import (
	"unsafe"
)

func UnsafeFoo(input string) {
	_ = unsafe.Pointer(&input)
}
