// Package resp implements low-level primitives for dealing
// with RESP (REdis Serialization Protocol). It provides client and
// server side readers and writers.
package resp

import (
	"fmt"
)

// ResponseType represents the reply type
type ResponseType uint8

// String returns the response type description
func (t ResponseType) String() string {
	switch t {
	case TypeArray:
		return "Array"
	case TypeBulk:
		return "Bulk"
	case TypeInline:
		return "Inline"
	case TypeError:
		return "Error"
	case TypeInt:
		return "Int"
	case TypeNil:
		return "Nil"
	}
	return "Unknown"
}

// response type iota
const (
	TypeUnknown ResponseType = iota
	TypeArray
	TypeBulk
	TypeInline
	TypeError
	TypeInt
	TypeNil
)

// --------------------------------------------------------------------

// Scannable interfaces may implement custom Scan behaviour
type Scannable interface {
	// ScanResponse scans theresponse from the reader
	ScanResponse(t ResponseType, r ResponseReader) error
}

// NullString is a scannable that can deal with nil values
type NullString struct {
	Value string
	Valid bool
}

// ScanResponse implements Scannable
func (s *NullString) ScanResponse(t ResponseType, r ResponseReader) error {
	if t == TypeNil {
		return r.ReadNil()
	}

	if err := r.Scan(&s.Value); err != nil {
		return err
	}
	s.Valid = true
	return nil
}

// --------------------------------------------------------------------

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
	errNotANumber             = protoError("Protocol error: expected a number")
	errNotANilMessage         = protoError("Protocol error: expected a nil")
	errBadResponseType        = protoError("Protocol error: bad response type")
)

var (
	binCRLF = []byte("\r\n")
	binOK   = []byte("+OK\r\n")
	binZERO = []byte(":0\r\n")
	binONE  = []byte(":1\r\n")
	binNIL  = []byte("$-1\r\n")
)

// MaxBufferSize is the max request/response buffer size
const MaxBufferSize = 64 * 1024

func mkStdBuffer() []byte { return make([]byte, MaxBufferSize) }
