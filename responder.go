package redeo

import (
	"bytes"
	"io"
	"strconv"
	"strings"
)

const (
	CodeInline  = '+'
	CodeError   = '-'
	CodeFixnum  = ':'
	CodeStrLen  = '$'
	CodeBulkLen = '*'
)

var (
	binCRLF = []byte("\r\n")
	binOK   = []byte("+OK\r\n")
	binZERO = []byte(":0\r\n")
	binONE  = []byte(":1\r\n")
	binNIL  = []byte("$-1\r\n")
)

// Responder generates client responses
type Responder struct {
	w io.Writer
	b *bytes.Buffer
}

// NewResponder creates a new responder instance
func NewResponder(w io.Writer) *Responder {
	return &Responder{w: w, b: new(bytes.Buffer)}
}

// WriteBulkLen writes a bulk length
func (res *Responder) WriteBulkLen(n int) int {
	return res.writeInline(CodeBulkLen, strconv.Itoa(n))
}

// WriteString writes a bulk string
func (res *Responder) WriteString(s string) int {
	n := res.writeInline(CodeStrLen, strconv.Itoa(len(s)))
	m, _ := res.b.WriteString(s)
	res.b.Write(binCRLF)
	return n + m + 2
}

// WriteBytes writes a bulk string
func (res *Responder) WriteBytes(b []byte) int {
	n := res.writeInline(CodeStrLen, strconv.Itoa(len(b)))
	m, _ := res.b.Write(b)
	res.b.Write(binCRLF)
	return n + m + 2
}

// WriteString writes an inline string
func (res *Responder) WriteInlineString(s string) int {
	return res.writeInline(CodeInline, s)
}

// WriteNil writes a nil value
func (res *Responder) WriteNil() int {
	n, _ := res.b.Write(binNIL)
	return n
}

// WriteOK writes OK
func (res *Responder) WriteOK() int {
	n, _ := res.b.Write(binOK)
	return n
}

// WriteInt writes an inline integer
func (res *Responder) WriteInt(n int) int {
	return res.writeInline(CodeFixnum, strconv.Itoa(n))
}

// WriteZero writes a 0 integer
func (res *Responder) WriteZero() int {
	n, _ := res.b.Write(binZERO)
	return n
}

// WriteOne writes a 1 integer
func (res *Responder) WriteOne() int {
	n, _ := res.b.Write(binONE)
	return n
}

// WriteErrorString writes an error string
func (res *Responder) WriteErrorString(s string) int {
	return res.writeInline(CodeError, s)
}

// WriteError writes an error using the standard "ERR message" format
func (res *Responder) WriteError(err error) int {
	s := err.Error()
	i := strings.LastIndex(s, ": ")
	if i < 0 {
		i = -2
	}
	return res.WriteErrorString("ERR " + s[i+2:])
}

// String returns the buffered string
func (res *Responder) String() string {
	return res.b.String()
}

// Len returns the buffered length
func (res *Responder) Len() int {
	return res.b.Len()
}

// Truncate truncates the buffer
func (res *Responder) Truncate(n int) {
	res.b.Truncate(n)
}

// StreamN streams data from a reader
func (res *Responder) StreamN(rd io.Reader, n int64) (int64, error) {
	o := strconv.FormatInt(n, 10)
	b := append([]byte{CodeStrLen}, []byte(o)...)

	_, err := res.Write(append(b, binCRLF...))
	if err != nil {
		return 0, err
	}
	m, err := io.CopyN(res.w, rd, n)
	if err != nil {
		return 0, err
	}

	_, err = res.Write(binCRLF)
	return int64(len(o)) + m + 5, err
}

// Write allows servers to write raw data straight to the socket without buffering.
// This is useful e.g. for streaming responses but may break the redis protocol.
// Be careful with this!
func (res *Responder) Write(p []byte) (int, error) {
	return res.w.Write(p)
}

// Flushes the buffered data to output
func (res *Responder) flush() error {
	_, err := res.b.WriteTo(res.w)
	return err
}

func (res *Responder) writeInline(prefix byte, s string) int {
	res.b.WriteByte(prefix) // Never returns an error
	n, _ := res.b.WriteString(s)
	res.b.Write(binCRLF)
	return n + 3
}
