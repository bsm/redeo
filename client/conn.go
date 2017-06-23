package client

import (
	"io"
	"net"
	"time"

	"github.com/bsm/redeo/resp"
)

// Conn wraps a single network connection and exposes
// common read/write methods.
type Conn interface {
	// MarkFailed marks the connection as failed which
	// will force it to be closed instead of being returned to the pool
	MarkFailed()

	// PeekType returns the type of the next response block
	PeekType() (resp.ResponseType, error)
	// ReadNil reads a nil value
	ReadNil() error
	// ReadBulk reads a bulk value (optionally appending it to a passed p buffer)
	ReadBulk(p []byte) ([]byte, error)
	// ReadBulkString reads a bulk value as string
	ReadBulkString() (string, error)
	// ReadInt reads an int value
	ReadInt() (int64, error)
	// ReadArrayLen reads the array length
	ReadArrayLen() (int, error)
	// ReadError reads an error string
	ReadError() (string, error)
	// ReadInline reads an inline status string
	ReadInlineString() (string, error)
	// StreamBulk returns a bulk-reader
	StreamBulk() (io.Reader, error)

	// WriteCmd writes a full command as part of a pipeline. To execute the pipeline,
	// you must call Flush.
	WriteCmd(cmd string, args ...[]byte)
	// WriteCmdString writes a full command as part of a pipeline. To execute the pipeline,
	// you must call Flush.
	WriteCmdString(cmd string, args ...string)
	// WriteMultiBulkSize is a low-level method to write a multibulk size.
	// For normal operation, use WriteCmd or WriteCmdString.
	WriteMultiBulkSize(n int) error
	// WriteBulk is a low-level method to write a bulk.
	// For normal operation, use WriteCmd or WriteCmdString.
	WriteBulk(b []byte)
	// WriteBulkString is a low-level method to write a bulk.
	// For normal operation, use WriteCmd or WriteCmdString.
	WriteBulkString(s string)
	// CopyBulk is a low-level method to copy a large bulk of data directly to the writer.
	// For normal operation, use WriteCmd or WriteCmdString.
	CopyBulk(src io.Reader, n int64) error
	// Flush flushes the output buffer. Call this after you have completed your pipeline
	Flush() error

	// SetDeadline sets the read and write deadlines associated
	// with the connection. It is equivalent to calling both
	// SetReadDeadline and SetWriteDeadline.
	SetDeadline(time.Time) error
	// SetReadDeadline sets the deadline for future Read calls
	// and any currently-blocked Read call.
	// A zero value for t means Read will not time out.
	SetReadDeadline(time.Time) error
	// SetWriteDeadline sets the deadline for future Write calls
	// and any currently-blocked Write call.
	// Even if write times out, it may return n > 0, indicating that
	// some of the data was successfully written.
	// A zero value for t means Write will not time out.
	SetWriteDeadline(time.Time) error

	madeByRedeo()
}

type conn struct {
	net.Conn

	*resp.RequestWriter
	resp.ResponseReader

	failed bool
}

// MarkFailed implements Conn interface.
func (c *conn) MarkFailed()  { c.failed = true }
func (c *conn) madeByRedeo() {}
