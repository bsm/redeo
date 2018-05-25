package replication

import (
	"errors"
	"sync"

	"github.com/bsm/redeo/resp"
)

const minBacklogSize = 1 << 17 // 128KiB

var errOffsetOutOfBounds = errors.New("requested offset is out of bounds")

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

// Write implements io.Writer interface.
func (b *backlog) Write(data []byte) (n int, _ error) {
	b.Lock()
	defer b.Unlock()

	bsize := len(b.buffer)
	b.offset += int64(len(data))

	for p := data; len(p) > 0; {
		written := copy(b.buffer[b.pos:], p)
		if b.pos += written; b.pos == bsize {
			b.pos = 0
		}

		b.histlen += written
		n += written
		p = p[written:]
	}

	if b.histlen > bsize {
		b.histlen = bsize
	}
	return
}

// MinOffset returns the minimum stored offset.
func (b *backlog) MinOffset() int64 {
	b.RLock()
	n := b.offset - int64(b.histlen)
	b.RUnlock()
	return n
}

// MaxOffset returns the maximum stored offset.
func (b *backlog) MaxOffset() int64 {
	b.RLock()
	n := b.offset
	b.RUnlock()
	return n
}
