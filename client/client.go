// Package client implements a minimalist client
// for working with redis servers.
package client

import (
	"net"
	"sync"

	"github.com/bsm/pool"
	"github.com/bsm/redeo/resp"
)

// Pool is a minimalist redis client connection pool
type Pool struct {
	conns   *pool.Pool
	readers sync.Pool
	writers sync.Pool
}

// New initializes a new pool with a custom dialer
func New(opt *pool.Options, dialer func() (net.Conn, error)) (*Pool, error) {
	if dialer == nil {
		dialer = func() (net.Conn, error) {
			return net.Dial("tcp", "127.0.0.1:6379")
		}
	}

	conns, err := pool.New(opt, dialer)
	if err != nil {
		return nil, err
	}

	return &Pool{
		conns: conns,
	}, nil
}

// Get returns a connection
func (p *Pool) Get() (Conn, error) {
	cn, err := p.conns.Get()
	if err != nil {
		return nil, err
	}

	return &conn{
		Conn: cn,

		RequestWriter:  p.newRequestWriter(cn),
		ResponseReader: p.newResponseReader(cn),
	}, nil
}

// Put allows to return a connection back to the pool.
// Call this method after every call/pipeline.
// Do not use the connection again after this method
// is triggered.
func (p *Pool) Put(cn Conn) {
	cs, ok := cn.(*conn)
	if !ok {
		return
	} else if cs.failed {
		_ = cs.Close()
		return
	}

	p.writers.Put(cs.RequestWriter)
	p.readers.Put(cs.ResponseReader)
	p.conns.Put(cs.Conn)
}

// Close closes the client and all underlying connections
func (p *Pool) Close() error {
	return p.conns.Close()
}

func (p *Pool) newRequestWriter(cn net.Conn) *resp.RequestWriter {
	if v := p.writers.Get(); v != nil {
		w := v.(*resp.RequestWriter)
		w.Reset(cn)
		return w
	}
	return resp.NewRequestWriter(cn)
}

func (p *Pool) newResponseReader(cn net.Conn) resp.ResponseReader {
	if v := p.readers.Get(); v != nil {
		r := v.(resp.ResponseReader)
		r.Reset(cn)
		return r
	}
	return resp.NewResponseReader(cn)
}
