// Package resp implements low-level primitives for dealing
// with RESP (REdis Serialization Protocol). It provides client and
// server side readers and writers.
package resp

import "fmt"

type protoError string

func (p protoError) Error() string { return string(p) }

func protoErrorf(m string, args ...interface{}) error {
	return protoError(fmt.Sprintf(m, args...))
}

// IsProtocolError returns true if the error is a protocol error
func IsProtocolError(err error) bool {
	_, ok := err.(protoError)
	return ok
}

const (
	errInvalidMultiBulkLength = protoError("Protocol error: invalid multibulk length")
	errInvalidBulkLength      = protoError("Protocol error: invalid bulk length")
	errBlankBulkLength        = protoError("Protocol error: expected '$', got ' '")
	errInlineRequestTooLong   = protoError("Protocol error: too big inline request")
	errNotAnInteger           = protoError("Protocol error: expected an integer")
	errNotANilMessage         = protoError("Protocol error: expected a nil")
)

var (
	binCRLF = []byte("\r\n")
	binOK   = []byte("+OK\r\n")
	binZERO = []byte(":0\r\n")
	binONE  = []byte(":1\r\n")
	binNIL  = []byte("$-1\r\n")
)

const MaxBufferSize = 64 * 1024

func mkStdBuffer() []byte { return make([]byte, MaxBufferSize) }
