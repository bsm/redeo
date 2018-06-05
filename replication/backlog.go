package replication

import (
	"bytes"
	"errors"
	"io"
	"sync"

	"github.com/bsm/redeo/resp"
)

const minBacklogSize = 1 << 17 // 128KiB

var bufPool, reqWPool sync.Pool

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

// Feed feeds the backlog with commands to propagate.
func (b *backlog) Feed(cmd *resp.Command) {
	var wb *bytes.Buffer
	if v := bufPool.Get(); v != nil {
		wb = v.(*bytes.Buffer)
	} else {
		wb = new(bytes.Buffer)
	}

	var rw *resp.RequestWriter
	if v := reqWPool.Get(); v != nil {
		rw = v.(*resp.RequestWriter)
		rw.Reset(wb)
	} else {
		rw = resp.NewRequestWriter(wb)
	}

	rw.WriteCommand(cmd)
	_ = rw.Flush()
	_, _ = b.Write(wb.Bytes())

	reqWPool.Put(rw)
	bufPool.Put(wb)
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

// Offsets returns the stored offset range.
func (b *backlog) Offsets() (earliest, latest int64) {
	b.RLock()
	earliest = b.offset
	latest = earliest + int64(b.histlen)
	b.RUnlock()
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
