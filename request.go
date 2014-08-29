package redeo

import (
	"bufio"
	"io"
	"strconv"
	"strings"
)

// Request contains a command, arguments, and client information
type Request struct {
	Name string      `json:"name"`
	Args []string    `json:"args,omitempty"`
	Ctx  interface{} `json:"ctx,omitempty"`

	client *Client
}

// Client returns the client
func (r *Request) Client() *Client {
	return r.client
}

// ParseRequest parses a new request from a buffered connection
func ParseRequest(rd *bufio.Reader) (req *Request, err error) {
	var line []byte
	if line, _, err = rd.ReadLine(); err != nil {
		return nil, io.EOF
	} else if len(line) < 1 {
		return nil, io.EOF
	}

	switch line[0] {
	case binASTERISK:
		var argc int
		if argc, err = strconv.Atoi(string(line[1:])); err != nil {
			return nil, ErrInvalidRequest
		}

		args := make([]string, argc)
		for i := 0; i < argc; i++ {
			args[i], err = parseArgument(rd)
			if err != nil {
				return
			}
		}
		req = &Request{Name: strings.ToLower(args[0]), Args: args[1:]}
	default:
		req = &Request{Name: strings.ToLower(string(line))}
	}
	return
}

func parseArgument(rd *bufio.Reader) (part string, err error) {
	var line []byte
	if line, _, err = rd.ReadLine(); err != nil {
		return "", io.EOF
	} else if len(line) < 1 {
		return "", io.EOF
	} else if line[0] != binDOLLAR {
		return "", ErrInvalidRequest
	}

	var blen int
	if blen, err = strconv.Atoi(string(line[1:])); err != nil {
		return "", ErrInvalidRequest
	}

	buf := make([]byte, blen+2)
	if _, err = io.ReadFull(rd, buf); err != nil {
		return "", io.EOF
	}

	return string(buf[:blen]), nil
}
