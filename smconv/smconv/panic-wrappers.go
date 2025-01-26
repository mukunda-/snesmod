// SNESMOD
// (C) 2025 Mukunda Johnson (mukunda.com)
// Licensed under MIT

package smconv

import (
	"encoding/binary"
	"io"
)

// Wrappers around common IO operations to simplify code and avoid nil checks everywhere.
// `defer pguard(&r)` is used to catch the panics.

// Panic guard, catches error panics and returns them. All exported functions should have
// this at the top. DON'T panic past the library boundary.
func pguard(err *error) {
	if r := recover(); r != nil {
		switch r.(type) {
		case error:
			*err = r.(error)
		default:
			panic(err)
		}
	}
}

// Binary write
func bwrite(w io.Writer, data any) {
	err := binary.Write(w, binary.LittleEndian, data)
	if err != nil {
		panic(err)
	}
}

// Shortcut for seek(0)
func ptell(w io.Seeker) int64 {
	pos, err := w.Seek(0, io.SeekCurrent)
	if err != nil {
		panic(err)
	}
	return pos
}

func pseek(w io.Seeker, offset int64, whence int) {
	_, err := w.Seek(offset, whence)
	if err != nil {
		panic(err)
	}
}

// Catch an error return and panic with it.
func pcatch(err error) {
	if err != nil {
		panic(err)
	}
}
