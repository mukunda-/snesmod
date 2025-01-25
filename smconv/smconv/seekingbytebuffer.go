// SNESMOD
// (C) 2025 Mukunda Johnson (mukunda.com)
// Licensed under MIT

package smconv

// In memory write-seek buffer.
type SeekingByteBuffer struct {
	buf    []byte
	cursor int
}

func (m *SeekingByteBuffer) Write(out []byte) (written int, err error) {
	if m.cursor < len(m.buf) {
		written = copy(m.buf[m.cursor:], out)
		m.cursor += written
	}

	if written < len(out) {
		m.buf = append(m.buf, out[written:]...)
		m.cursor = len(m.buf)
		written = len(out)
	}
	return written, nil
}

func (m *SeekingByteBuffer) Seek(offset int64, whence int) (int64, error) {
	switch whence {
	case 0:
		m.cursor = int(offset)
	case 1:
		m.cursor += int(offset)
	case 2:
		m.cursor = len(m.buf) - int(offset)
	}

	if m.cursor < 0 {
		m.cursor = 0
	}
	if m.cursor > len(m.buf) {
		m.cursor = len(m.buf)
	}

	return int64(m.cursor), nil
}

func (m *SeekingByteBuffer) Tell() int {
	return m.cursor
}

func (m *SeekingByteBuffer) Bytes() []byte {
	return m.buf
}
