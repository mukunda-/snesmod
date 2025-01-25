// SNESMOD
// (C) 2025 Mukunda Johnson (mukunda.com)
// Licensed under MIT

package smconv

import (
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSeekingWriteBuffer(t *testing.T) {
	buf := SeekingByteBuffer{}

	{
		// Writing bytes at the end extends the buffer with the bytes.
		n, err := buf.Write([]byte{1, 2, 3})
		assert.NoError(t, err)
		assert.EqualValues(t, 3, n)
	}

	{
		// Seek allows us to see the current cursor position.
		offset, err := buf.Seek(0, io.SeekCurrent)
		assert.NoError(t, err)
		assert.EqualValues(t, 3, offset)
	}

	{
		// Seeking moves the cursor, where data in the buffer will be overwritten by the
		// next write operation.
		buf.Seek(-2, io.SeekCurrent)
		n, err := buf.Write([]byte{4, 5, 6})
		assert.NoError(t, err)
		assert.EqualValues(t, 3, n)
		assert.Equal(t, []byte{1, 4, 5, 6}, buf.Bytes())
	}

	{
		// Seeking past the beginning or end of the data is clamped to the boundary.
		buf.Seek(-200, io.SeekCurrent)
		assert.EqualValues(t, 0, buf.Tell())
		buf.Seek(200, io.SeekCurrent)
		assert.EqualValues(t, 4, buf.Tell())
	}

	{
		buf.Seek(2, io.SeekStart)
		buf.Write([]byte{9})
		assert.Equal(t, []byte{1, 4, 9, 6}, buf.Bytes())
		buf.Write([]byte{10})
		assert.Equal(t, []byte{1, 4, 9, 10}, buf.Bytes())
		buf.Write([]byte{11})
		assert.Equal(t, []byte{1, 4, 9, 10, 11}, buf.Bytes())
		buf.Write([]byte{12})
		assert.Equal(t, []byte{1, 4, 9, 10, 11, 12}, buf.Bytes())
	}
}
