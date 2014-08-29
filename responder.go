package redeo

import (
	"bytes"
	"io"
	"strconv"
	"strings"
)

const (
	binPLUS     = '+'
	binMINUS    = '-'
	binCOLON    = ':'
	binDOLLAR   = '$'
	binASTERISK = '*'
	binERR      = "ERR "
)

var (
	binCRLF = []byte("\r\n")
	binOK   = []byte("+OK\r\n")
	binZERO = []byte(":0\r\n")
	binONE  = []byte(":1\r\n")
	binNIL  = []byte("$-1\r\n")
)

// Responder genereates client responses
type Responder struct {
	b *bytes.Buffer
}

// NewResponder creates a new responder instance
func NewResponder() *Responder {
	return &Responder{b: new(bytes.Buffer)}
}

// WriteBulkLen writes a bulk length
func (res Responder) WriteBulkLen(n int) int {
	return res.writeInline(binASTERISK, strconv.Itoa(n))
}

// WriteString writes a bulk string
func (res Responder) WriteString(s string) int {
	n := res.writeInline(binDOLLAR, strconv.Itoa(len(s)))
	m, _ := res.b.WriteString(s)
	res.b.Write(binCRLF)
	return n + m + 2
}

// WriteBytes writes a bulk string
func (res Responder) WriteBytes(b []byte) int {
	n := res.writeInline(binDOLLAR, strconv.Itoa(len(b)))
	m, _ := res.b.Write(b)
	res.b.Write(binCRLF)
	return n + m + 2
}

// WriteString writes an inline string
func (res Responder) WriteInlineString(s string) int {
	return res.writeInline(binPLUS, s)
}

// WriteNil writes a nil value
func (res Responder) WriteNil() int {
	n, _ := res.b.Write(binNIL)
	return n
}

// WriteOK writes OK
func (res Responder) WriteOK() int {
	n, _ := res.b.Write(binOK)
	return n
}

// WriteInt writes an inline integer
func (res Responder) WriteInt(n int) int {
	return res.writeInline(binCOLON, strconv.Itoa(n))
}

// WriteZero writes a 0 integer
func (res Responder) WriteZero() int {
	n, _ := res.b.Write(binZERO)
	return n
}

// WriteOne writes a 1 integer
func (res Responder) WriteOne() int {
	n, _ := res.b.Write(binONE)
	return n
}

// WriteErrorString writes an error string
func (res Responder) WriteErrorString(s string) int {
	return res.writeInline(binMINUS, s)
}

// WriteError writes an error
func (res Responder) WriteError(err error) int {
	s := err.Error()
	i := strings.LastIndex(s, ": ")
	if i < 0 {
		i = -2
	}
	return res.WriteErrorString(binERR + s[i+2:])
}

// String returns the buffered string
func (res Responder) String() string {
	return res.b.String()
}

// Len returns the buffered length
func (res Responder) Len() int {
	return res.b.Len()
}

// Truncate truncates the buffer
func (res Responder) Truncate(n int) {
	res.b.Truncate(n)
}

// Write writes raw data to the buffer (implements io.Writer interface)
func (res Responder) Write(p []byte) (int, error) {
	return res.b.Write(p)
}

// WriteTo writes the buffer to a writer (implements io.WriterTo interface)
func (res Responder) WriteTo(w io.Writer) (int64, error) {
	return res.b.WriteTo(w)
}

func (res Responder) writeInline(prefix byte, s string) int {
	res.b.WriteByte(prefix) // Never returns an error
	n, _ := res.b.WriteString(s)
	res.b.Write(binCRLF)
	return n + 3
}
