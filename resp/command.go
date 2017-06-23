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

	baseCmd
}

// Arg returns the command argument at position i
func (c *Command) Arg(i int) CommandArgument {
	if i > -1 && i < c.argc {
		return c.Args()[i]
	}
	return nil
}

// Args returns all command argument values
func (c *Command) Args() []CommandArgument { return c.argv }

func (c *Command) updateName() {
	c.Name = string(c.name)
}

func (c *Command) reset() {
	c.baseCmd.reset()
	*c = Command{baseCmd: c.baseCmd}
}

func (c *Command) parseMultiBulk(r *bufioR) (bool, error) {
	ok, err := c.baseCmd.parseMultiBulk(r)
	if err != nil || !ok {
		return ok, err
	}

	c.grow(c.argc)
	for i := 0; i < c.argc; i++ {
		c.argv[i], err = r.ReadBulk(c.argv[i])
		if err != nil {
			return false, err
		}
	}
	return true, nil
}

// --------------------------------------------------------------------

var errNoMoreArgs = errors.New("resp: no more arguments")

// CommandStream instances are created by a RequestReader
type CommandStream struct {
	// Name refers to the command name
	Name string

	baseCmd

	p int
	r *bufioR
}

// Discard discards the (remaining) arguments
func (c *CommandStream) Discard() error {
	if c.p < len(c.argv) {
		c.p = len(c.argv)
		return nil
	}

	var err error
	if c.r == nil {
		for ; c.p < c.argc; c.p++ {
			if e := c.r.SkipBulk(); e != nil {
				err = e
			}
		}
	}
	return err
}

// NextArg returns the next argument as an io.Reader
func (c *CommandStream) NextArg() (io.Reader, error) {
	if c.p < len(c.argv) {
		rd := bytes.NewReader(c.argv[c.p])
		c.p++
		return rd, nil
	} else if c.p < c.argc && c.r != nil {
		rd, err := c.r.StreamBulk()
		c.p++
		return rd, err
	}
	return nil, errNoMoreArgs
}

func (c *CommandStream) updateName() {
	c.Name = string(c.name)
}

func (c *CommandStream) reset() {
	c.baseCmd.reset()
	*c = CommandStream{baseCmd: c.baseCmd}
}

func (c *CommandStream) parseMultiBulk(r *bufioR) (bool, error) {
	ok, err := c.baseCmd.parseMultiBulk(r)
	if err != nil || !ok {
		return ok, err
	}

	if c.argc > 0 {
		c.r = r
	}
	return true, nil
}

// --------------------------------------------------------------------

type anyCmd interface {
	parseMultiBulk(*bufioR) (bool, error)
	parseInline(*bufioR) (bool, error)
	updateName()
}

func parseCommand(c anyCmd, r *bufioR) error {
	x, err := r.PeekByte()
	if err != nil {
		return err
	}

	if x == '*' {
		if ok, err := c.parseMultiBulk(r); err != nil {
			return err
		} else if !ok {
			return parseCommand(c, r)
		}
		c.updateName()
		return nil
	}

	if ok, err := c.parseInline(r); err != nil {
		return err
	} else if !ok {
		return parseCommand(c, r)
	}
	c.updateName()
	return nil
}

// --------------------------------------------------------------------

type baseCmd struct {
	argc int
	argv []CommandArgument
	name []byte

	ctx context.Context
}

// ArgN returns the number of command arguments
func (c *baseCmd) ArgN() int {
	return c.argc
}

// Context returns the context
func (c *baseCmd) Context() context.Context {
	if c.ctx != nil {
		return c.ctx
	}
	return context.Background()
}

// SetContext sets the request context.
func (c *baseCmd) SetContext(ctx context.Context) {
	if ctx != nil {
		c.ctx = ctx
	}
}

func (c *baseCmd) parseMultiBulk(r *bufioR) (bool, error) {
	n, err := r.ReadArrayLen()
	if err != nil || n < 1 {
		return false, err
	}

	c.argc = n - 1
	c.name, err = r.ReadBulk(c.name)
	if err != nil {
		return false, err
	}

	return true, nil
}

func (c *baseCmd) parseInline(r *bufioR) (bool, error) {
	line, err := r.ReadLine()
	if err != nil {
		return false, err
	}

	hasName := false
	inWord := false
	for _, x := range line.Trim() {
		switch x {
		case ' ', '\t':
			inWord = false
		default:
			if !inWord && hasName {
				c.argc++
				c.grow(c.argc)
			}
			if pos := c.argc - 1; pos > -1 {
				c.argv[pos] = append(c.argv[pos], x)
			} else {
				hasName = true
				c.name = append(c.name, x)
			}
			inWord = true
		}
	}
	return hasName, nil
}

func (c *baseCmd) grow(n int) {
	if d := n - cap(c.argv); d > 0 {
		c.argv = c.argv[:cap(c.argv)]
		c.argv = append(c.argv, make([]CommandArgument, d)...)
	} else {
		c.argv = c.argv[:n]
	}
}

func (c *baseCmd) reset() {
	argv := c.argv
	for i, v := range argv {
		argv[i] = v[:0]
	}
	*c = baseCmd{
		argv: argv[:0],
		name: c.name[:0],
	}
}
