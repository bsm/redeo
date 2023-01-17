package redeo

import (
	"context"
	"net"
	"sync"
	"sync/atomic"

	"github.com/bsm/redeo/v2/resp"
)

var (
	clientInc  = uint64(0)
	readerPool sync.Pool
	writerPool sync.Pool
)

type ctxKeyClient struct{}

// Client contains information about a client connection
type Client struct {
	id uint64
	cn net.Conn

	rd *resp.RequestReader
	wr resp.ResponseWriter

	ctx    context.Context
	closed bool

	cmd  *resp.Command
	scmd *resp.CommandStream
}

func newClient(cn net.Conn) *Client {
	c := new(Client)
	c.reset(cn)
	return c
}

// GetClient retrieves the client from a the context.
// This function may return nil if a client is not set.
func GetClient(ctx context.Context) *Client {
	if ctx != nil {
		if c, ok := ctx.Value(ctxKeyClient{}).(*Client); ok {
			return c
		}
	}
	return nil
}

// ID return the unique client id
func (c *Client) ID() uint64 { return c.id }

// Context return the client context
func (c *Client) Context() context.Context {
	if c.ctx != nil {
		return c.ctx
	}
	return context.Background()
}

// SetContext sets the client's context
func (c *Client) SetContext(ctx context.Context) {
	c.ctx = ctx
}

// RemoteAddr return the remote client address
func (c *Client) RemoteAddr() net.Addr {
	return c.cn.RemoteAddr()
}

// Close will disconnect as soon as all pending replies have been written
// to the client
func (c *Client) Close() {
	c.closed = true
}

func (c *Client) readCmd(cmd *resp.Command) (*resp.Command, error) {
	var err error
	if cmd, err = c.rd.ReadCmd(cmd); err == nil {
		cmd.SetContext(context.WithValue(cmd.Context(), ctxKeyClient{}, c))
	}
	return cmd, err
}

func (c *Client) streamCmd(cmd *resp.CommandStream) (*resp.CommandStream, error) {
	var err error
	if cmd, err = c.rd.StreamCmd(cmd); err == nil {
		cmd.SetContext(context.WithValue(cmd.Context(), ctxKeyClient{}, c))
	}
	return cmd, err
}

func (c *Client) pipeline(fn func(string) error) error {
	for more := true; more; more = c.rd.Buffered() != 0 {
		name, err := c.rd.PeekCmd()
		if err != nil {
			_ = c.rd.SkipCmd()
			return err
		}
		if err := fn(name); err != nil {
			return err
		}
	}
	return nil
}

func (c *Client) release() {
	_ = c.cn.Close()
	readerPool.Put(c.rd)
	writerPool.Put(c.wr)
}

func (c *Client) reset(cn net.Conn) {
	*c = Client{
		id: atomic.AddUint64(&clientInc, 1),
		cn: cn,
	}

	if v := readerPool.Get(); v != nil {
		rd := v.(*resp.RequestReader)
		rd.Reset(cn)
		c.rd = rd
	} else {
		c.rd = resp.NewRequestReader(cn)
	}

	if v := writerPool.Get(); v != nil {
		wr := v.(resp.ResponseWriter)
		wr.Reset(cn)
		c.wr = wr
	} else {
		c.wr = resp.NewResponseWriter(cn)
	}
}
