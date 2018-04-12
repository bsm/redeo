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
	resp.ResponseParser

	// MarkFailed marks the connection as failed which
	// will force it to be closed instead of being returned to the pool
	MarkFailed()

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

	// UnreadBytes returns the number of unread bytes.
	UnreadBytes() int
	// UnflushedBytes returns the number of pending/unflushed bytes.
	UnflushedBytes() int

	// Close (force) closes the connection.
	Close() error

	madeByRedeo()
}

// Wrap wraps a single network connection.
func Wrap(cn net.Conn) Conn {
	return &conn{
		Conn: cn,

		RequestWriter:  resp.NewRequestWriter(cn),
		ResponseReader: resp.NewResponseReader(cn),
	}
}

type conn struct {
	net.Conn

	*resp.RequestWriter
	resp.ResponseReader

	failed bool
}

// MarkFailed implements Conn interface.
func (c *conn) MarkFailed() { c.failed = true }

// UnreadBytes implements Conn interface.
func (c *conn) UnreadBytes() int { return c.ResponseReader.Buffered() }

// UnflushedBytes implements Conn interface.
func (c *conn) UnflushedBytes() int { return c.RequestWriter.Buffered() }

// Close implements Conn interface.
func (c *conn) Close() error { c.failed = true; return c.Conn.Close() }

func (c *conn) madeByRedeo() {}
