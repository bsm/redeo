package resp

import (
	"bytes"
	"context"
	"errors"
	"io"
	"strconv"
)

// CommandArgument is an argument of a command
type CommandArgument []byte

// Bytes returns the argument as bytes
func (c CommandArgument) Bytes() []byte { return c }

// String returns the argument converted to a string
func (c CommandArgument) String() string { return string(c) }

// Float returns the argument as a float64.
func (c CommandArgument) Float() (float64, error) {
	return strconv.ParseFloat(string(c), 64)
}

// Int returns the argument as an int64.
func (c CommandArgument) Int() (int64, error) {
	return strconv.ParseInt(string(c), 10, 64)
}

// --------------------------------------------------------------------

// Command instances are parsed by a RequestReader
type Command struct {
	// Name refers to the command name
	Name string

	// Args returns arguments
	Args []CommandArgument

	ctx context.Context
}

// NewCommand returns a new command instance;
// useful for tests
func NewCommand(name string, args ...CommandArgument) *Command {
	return &Command{Name: name, Args: args}
}

// Arg returns the Nth argument
func (c *Command) Arg(n int) CommandArgument {
	if n > -1 && n < len(c.Args) {
		return c.Args[n]
	}
	return nil
}

// ArgN returns the number of arguments
func (c *Command) ArgN() int {
	return len(c.Args)
}

// Reset discards all data and resets all state
func (c *Command) Reset() {
	args := c.Args
	for i, v := range args {
		args[i] = v[:0]
	}
	*c = Command{Args: args[:0]}
}

// Context returns the context
func (c *Command) Context() context.Context {
	if c.ctx != nil {
		return c.ctx
	}
	return context.Background()
}

// SetContext sets the request context.
func (c *Command) SetContext(ctx context.Context) {
	if ctx != nil {
		c.ctx = ctx
	}
}

func (c *Command) grow(n int) {
	if d := n - cap(c.Args); d > 0 {
		c.Args = c.Args[:cap(c.Args)]
		c.Args = append(c.Args, make([]CommandArgument, d)...)
	} else {
		c.Args = c.Args[:n]
	}
}

func (c *Command) readMultiBulk(r *bufioR, name string, nargs int) error {
	c.Name = name
	c.grow(nargs)

	var err error
	for i := 0; i < nargs; i++ {
		c.Args[i], err = r.ReadBulk(c.Args[i])
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *Command) readInline(r *bufioR) (bool, error) {
	line, err := r.ReadLine()
	if err != nil {
		return false, err
	}

	data := line.Trim()

	var name []byte
	var n int

	name, n = appendArgument(name, data)
	data = data[n:]
	if len(name) == 0 {
		return false, nil
	}

	for pos := 0; len(data) != 0; pos++ {
		c.grow(pos + 1)
		c.Args[pos], n = appendArgument(c.Args[pos], data)
		data = data[n:]
	}

	c.Name = string(name)
	return true, nil
}

// --------------------------------------------------------------------

func readCommand(c interface {
	readInline(*bufioR) (bool, error)
	readMultiBulk(*bufioR, string, int) error
}, r *bufioR) error {
	x, err := r.PeekByte()
	if err != nil {
		return err
	}

	if x == '*' {
		sz, err := r.ReadArrayLen()
		if err != nil {
			return err
		} else if sz < 1 {
			return readCommand(c, r)
		}

		name, err := r.ReadBulkString()
		if err != nil {
			return err
		}

		return c.readMultiBulk(r, name, sz-1)
	}

	if ok, err := c.readInline(r); err != nil {
		return err
	} else if !ok {
		return readCommand(c, r)
	}

	return nil
}

// --------------------------------------------------------------------

var errNoMoreArgs = errors.New("resp: no more arguments")

// CommandStream instances are created by a RequestReader
type CommandStream struct {
	// Name refers to the command name
	Name string

	ctx context.Context

	inline   Command
	isInline bool

	nargs int
	pos   int
	arg   io.ReadCloser

	rd *bufioR
}

// Reset discards all data and resets all state
func (c *CommandStream) Reset() {
	c.inline.Reset()
	*c = CommandStream{inline: c.inline}
}

// Discard discards the (remaining) arguments
func (c *CommandStream) Discard() error {
	if c.isInline {
		if c.pos < len(c.inline.Args) {
			c.pos = len(c.inline.Args)
			return nil
		}
	}

	var err error
	if c.arg != nil {
		if e := c.arg.Close(); e != nil {
			err = e
		}
		c.arg = nil
	}

	if c.rd != nil {
		for ; c.pos < c.nargs; c.pos++ {
			if e := c.rd.SkipBulk(); e != nil {
				err = e
			}
		}
	}
	return err
}

// ArgN returns the number of arguments
func (c *CommandStream) ArgN() int {
	if c.isInline {
		return c.inline.ArgN()
	}
	return c.nargs
}

// More returns true if there are unread arguments
func (c *CommandStream) More() bool {
	return c.pos < c.ArgN()
}

// Next returns the next argument as an io.Reader
func (c *CommandStream) Next() (io.Reader, error) {
	if c.ctx != nil {
		if err := c.ctx.Err(); err != nil {
			return nil, err
		}
	}
	if !c.More() {
		return nil, errNoMoreArgs
	}

	if c.isInline {
		arg := bytes.NewReader(c.inline.Args[c.pos])
		c.pos++
		return arg, nil
	}

	var err error
	c.arg, err = c.rd.StreamBulk()
	c.pos++
	return c.arg, err
}

// Context returns the context
func (c *CommandStream) Context() context.Context {
	if c.ctx != nil {
		return c.ctx
	}
	return context.Background()
}

// SetContext sets the request context.
func (c *CommandStream) SetContext(ctx context.Context) {
	if ctx != nil {
		c.ctx = ctx
	}
}

func (c *CommandStream) readMultiBulk(r *bufioR, name string, nargs int) error {
	c.Name = name
	c.nargs = nargs
	c.rd = r
	return nil
}

func (c *CommandStream) readInline(r *bufioR) (bool, error) {
	c.isInline = true

	if ok, err := c.inline.readInline(r); err != nil || !ok {
		return ok, err
	}
	c.Name = c.inline.Name
	return true, nil
}
