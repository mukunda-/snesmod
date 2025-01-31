// SNESMOD
// (C) 2025 Mukunda Johnson (mukunda.com)
// Licensed under MIT

package smconv

import (
	"encoding/binary"
	"io"

	"go.mukunda.com/errorcat"
)

type eC = errorcat.Context

// Binary write
func bwrite(cat eC, w io.Writer, data any) {
	err := binary.Write(w, binary.LittleEndian, data)
	cat.Catch(err)
}

// Shortcut for seek(0)
func ptell(cat eC, w io.Seeker) int64 {
	pos, err := w.Seek(0, io.SeekCurrent)
	cat.Catch(err)
	return pos
}

func pseek(cat eC, w io.Seeker, offset int64, whence int) {
	_, err := w.Seek(offset, whence)
	cat.Catch(err)
}
