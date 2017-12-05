package redeotest

import (
	"bytes"
	"fmt"
	"io"
	"strconv"

	"github.com/bsm/redeo/resp"
)

// ErrorResponse is used to wrap error strings
type ErrorResponse string

// Error implements error interface
func (e ErrorResponse) Error() string { return string(e) }

// ResponseRecorder is an implementation of resp.ResponseWriter that
// is helpful in tests.
type ResponseRecorder struct {
	resp.ResponseWriter
	b *bytes.Buffer
}

// NewRecorder inits a new recorder
func NewRecorder() *ResponseRecorder {
	b := new(bytes.Buffer)
	return &ResponseRecorder{
		b:              b,
		ResponseWriter: resp.NewResponseWriter(b),
	}
}

// Len returns the raw byte length
func (r *ResponseRecorder) Len() int {
	_ = r.ResponseWriter.Flush()
	return r.b.Len()
}

// String returns the raw string
func (r *ResponseRecorder) String() string {
	_ = r.ResponseWriter.Flush()
	return r.b.String()
}

// Quoted returns the quoted string
func (r *ResponseRecorder) Quoted() string {
	return strconv.Quote(r.String())
}

// Response returns the first response
func (r *ResponseRecorder) Response() (interface{}, error) {
	vv, err := r.Responses()
	if err != nil || len(vv) == 0 {
		return nil, err
	}
	return vv[len(vv)-1], nil
}

// Responses returns all responses
func (r *ResponseRecorder) Responses() ([]interface{}, error) {
	_ = r.ResponseWriter.Flush()

	vv := make([]interface{}, 0)
	rr := resp.NewResponseReader(bytes.NewReader(r.b.Bytes()))

	for {
		v, err := parseResult(rr)
		if err == io.EOF {
			break
		}
		vv = append(vv, v)
	}
	return vv, nil
}

func parseResult(rr resp.ResponseReader) (interface{}, error) {
	typ, err := rr.PeekType()
	if err != nil {
		return nil, err
	}

	switch typ {
	case resp.TypeBulk:
		return rr.ReadBulkString()
	case resp.TypeInline:
		return rr.ReadInlineString()
	case resp.TypeInt:
		return rr.ReadInt()
	case resp.TypeError:
		s, err := rr.ReadError()
		if err != nil {
			return nil, err
		}
		return ErrorResponse(s), nil
	case resp.TypeNil:
		return nil, rr.ReadNil()
	case resp.TypeArray:
		sz, err := rr.ReadArrayLen()
		if err != nil {
			return nil, err
		}

		vv := make([]interface{}, sz)
		for i := 0; i < int(sz); i++ {
			if vv[i], err = parseResult(rr); err != nil {
				return nil, err
			}
		}
		return vv, nil
	default:
		return nil, fmt.Errorf("unexpected response %v", typ)
	}
}
