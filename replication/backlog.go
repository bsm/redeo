package replication

import (
	"errors"
	"io"
	"sync"

	"github.com/bsm/redeo/resp"
)

const minBacklogSize = 1 << 17 // 128KiB

var ErrOffsetOutOfBounds = errors.New("redeo: requested offset is out of bounds")

type backlog struct {
	buffer  []byte
	offset  int64 // start offset
	histlen int   // actual data length
	pos     int   // circular buffer current position

	sync.RWMutex
}

func newBacklog(offset int64, size int) *backlog {
	if size < minBacklogSize {
		size = minBacklogSize
	}

	return &backlog{
		buffer: make([]byte, size),
		offset: offset,
	}
}

func (b *backlog) Feed(cmd *resp.Command) error {
	return nil
}

// Resize resizes the backlog by recreating the buffer
func (b *backlog) Resize(size int) {
	if size < minBacklogSize {
		size = minBacklogSize
	}

	buffer := make([]byte, size)

	b.Lock()
	defer b.Unlock()

	if len(b.buffer) == size {
		return
	}

	b.buffer = buffer
	b.offset += int64(b.histlen)
	b.histlen = 0
	b.pos = 0
}

// Write implements io.Writer interface.
func (b *backlog) Write(data []byte) (n int, _ error) {
	b.Lock()
	defer b.Unlock()

	bsize := len(b.buffer)
	for p := data; len(p) > 0; {
		written := copy(b.buffer[b.pos:], p)
		if b.pos += written; b.pos == bsize {
			b.pos = 0
		}

		b.histlen += written
		n += written
		p = p[written:]
	}

	if delta := b.histlen - bsize; delta > 0 {
		b.offset += int64(delta)
		b.histlen = bsize
	}
	return
}

// ReadAt implements io.ReaderAt interface.
func (b *backlog) ReadAt(p []byte, off int64) (n int, err error) {
	b.RLock()
	defer b.RUnlock()

	if off < b.offset {
		return 0, ErrOffsetOutOfBounds
	}
	if off > b.offset+int64(b.histlen) {
		return 0, ErrOffsetOutOfBounds
	}

	if blen := len(b.buffer); b.histlen < blen {
		n += copy(p, b.buffer[int(off-b.offset):b.pos])
	} else if b.histlen == blen && b.pos == 0 {
		n += copy(p, b.buffer[int(off-b.offset):])
	} else {
		o1, o2 := int(off-b.offset)+b.pos, 0
		if o1 < blen {
			n += copy(p, b.buffer[o1:])
		} else {
			o2 = o1 - blen
		}
		n += copy(p[n:], b.buffer[o2:b.pos])
	}

	if n < len(p) {
		err = io.EOF
	}
	return
}
