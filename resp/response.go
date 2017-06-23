package resp

import "io"

// ResponseWriter is used by servers to wrap a client connection and send
// protocol-compatible responses in buffered pipelines.
type ResponseWriter interface {
	// AppendArrayLen appends an array header to the output buffer.
	AppendArrayLen(n int)
	// AppendBulk appends bulk bytes to the output buffer.
	AppendBulk(p []byte)
	// AppendBulkString appends a bulk string to the output buffer.
	AppendBulkString(s string)
	// AppendInline appends inline bytes to the output buffer.
	AppendInline(p []byte)
	// AppendInlineString appends an inline string to the output buffer.
	AppendInlineString(s string)
	// AppendError appends an error message to the output buffer.
	AppendError(msg string)
	// AppendErrorf appends an error message to the output buffer.
	AppendErrorf(pattern string, args ...interface{})
	// AppendInt appends a numeric response to the output buffer.
	AppendInt(n int64)
	// AppendNil appends a nil-value to the output buffer.
	AppendNil()
	// AppendOK appends "OK" to the output buffer.
	AppendOK()
	// CopyBulk copies n bytes from a reader.
	// This call may flush pending buffer to prevent overflows.
	CopyBulk(src io.Reader, n int64) error
	// Buffered returns the number of pending bytes.
	Buffered() int
	// Flush flushes pending buffer.
	Flush() error
	// Reset resets the writer to a new writer and recycles internal buffers.
	Reset(w io.Writer)
}

// NewResponseWriter wraps any writer interface, but
// normally a net.Conn.
func NewResponseWriter(wr io.Writer) ResponseWriter {
	w := new(bufioW)
	w.reset(mkStdBuffer(), wr)
	return w
}

// --------------------------------------------------------------------

// ResponseType represents the reply type
type ResponseType uint8

const (
	TypeUnknown ResponseType = iota
	TypeArray
	TypeBulk
	TypeInline
	TypeError
	TypeInt
	TypeNil
)

// ResponseReader is used by clients to wrap a server connection and
// parse responses.
type ResponseReader interface {
	// PeekType returns the type of the next response block
	PeekType() (ResponseType, error)
	// ReadNil reads a nil value
	ReadNil() error
	// ReadBulkString reads a bulk and returns a string
	ReadBulkString() (string, error)
	// ReadBulk reads a bulk and returns bytes (optionally appending to a passed p buffer)
	ReadBulk(p []byte) ([]byte, error)
	// StreamBulk parses a bulk responses and returns a streaming reader object
	StreamBulk() (io.Reader, error)
	// ReadInt reads an int value
	ReadInt() (int64, error)
	// ReadArrayLen reads the array length
	ReadArrayLen() (int, error)
	// ReadError reads an error string
	ReadError() (string, error)
	// ReadInlineString reads a status string
	ReadInlineString() (string, error)
	// Reset resets the reader to a new reader and recycles internal buffers.
	Reset(r io.Reader)
}

// ResponseReader wraps any reader interface, but
// normally a net.Conn.
func NewResponseReader(rd io.Reader) ResponseReader {
	r := new(bufioR)
	r.reset(mkStdBuffer(), rd)
	return r
}
