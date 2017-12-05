package resp

import (
	"io"
)

// CustomResponse values implement custom serialization and can be passed
// to ResponseWriter.Append.
type CustomResponse interface {
	// AppendTo must be implemented by custom response types
	AppendTo(w ResponseWriter)
}

// ResponseWriter is used by servers to wrap a client connection and send
// protocol-compatible responses in buffered pipelines.
type ResponseWriter interface {
	io.Writer

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
	// Append automatically serialized given values and appends them to the output buffer.
	// Supported values include:
	//   * nil
	//   * error
	//   * string
	//   * []byte
	//   * bool
	//   * float32, float64
	//   * int, int8, int16, int32, int64
	//   * uint, uint8, uint16, uint32, uint64
	//   * CustomResponse instances
	//   * slices and maps of any of the above
	Append(v interface{}) error
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

// ResponseParser is a basic response parser
type ResponseParser interface {
	// PeekType returns the type of the next response block
	PeekType() (ResponseType, error)
	// ReadNil reads a nil value
	ReadNil() error
	// ReadBulkString reads a bulk and returns a string
	ReadBulkString() (string, error)
	// ReadBulk reads a bulk and returns bytes (optionally appending to a passed p buffer)
	ReadBulk(p []byte) ([]byte, error)
	// StreamBulk parses a bulk responses and returns a streaming reader object
	// Returned responses must be closed.
	StreamBulk() (io.ReadCloser, error)
	// ReadInt reads an int value
	ReadInt() (int64, error)
	// ReadArrayLen reads the array length
	ReadArrayLen() (int, error)
	// ReadError reads an error string
	ReadError() (string, error)
	// ReadInlineString reads a status string
	ReadInlineString() (string, error)
	// Scan scans results into the given values.
	Scan(vv ...interface{}) error
}

// ResponseReader is used by clients to wrap a server connection and
// parse responses.
type ResponseReader interface {
	ResponseParser

	// Buffered returns the number of buffered (unread) bytes.
	Buffered() int
	// Reset resets the reader to a new reader and recycles internal buffers.
	Reset(r io.Reader)
}

// NewResponseReader returns ResponseReader, which wraps any reader interface, but
// normally a net.Conn.
func NewResponseReader(rd io.Reader) ResponseReader {
	r := new(bufioR)
	r.reset(mkStdBuffer(), rd)
	return r
}
