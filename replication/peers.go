package replication

import (
	"net"
	"time"

	"github.com/bsm/redeo/client"
)

type masterConn struct {
	client.Conn
	Addr string
}

func newMasterConn(addr string, timeout time.Duration) (*masterConn, error) {
	cn, err := net.DialTimeout("tcp", addr, timeout)
	if err != nil {
		return nil, err
	}
	return &masterConn{
		Conn: client.Wrap(cn),
		Addr: addr,
	}, nil
}
