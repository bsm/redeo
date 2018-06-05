package replication

import (
	"net"
	"sync"
	"time"

	"github.com/bsm/redeo/client"
)

const (
	replStateNone        uint32 = iota // no replication
	replStateMustConnect               // must connect to master
	replStateConnecting                // establishing a connection
	replStateConnected                 // connected
)

type masterLink struct {
	conn    client.Conn // active connection
	active  string      // address of the active connection
	target  string      // address of the target master
	state   uint32
	timeout time.Duration

	mu sync.RWMutex
}

// SetAddr sets a new master address.
func (l *masterLink) SetAddr(addr string) {
	l.mu.Lock()
	l.target = addr
	l.mu.Unlock()
}

// Addr returns the master address.
func (l *masterLink) Addr() string {
	l.mu.RLock()
	addr := l.target
	l.mu.RUnlock()

	return addr
}

// MaintainConn ensures connection is established to the right master.
func (l *masterLink) MaintainConn() error {
	l.mu.RLock()
	target, active := l.target, l.active
	l.mu.RUnlock()

	if target == active {
		return nil
	}

	// if target is not set, try discannect
	if target == "" {
		return l.safeUpdateConn("", nil, replStateNone)
	}

	// set new target
	cn, err := net.DialTimeout("tcp", target, l.timeout)
	if err != nil {
		return err
	}
	return l.safeUpdateConn(target, client.Wrap(cn), replStateMustConnect)
}

// Close closes the master link.
func (l *masterLink) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.active = ""
	if l.conn != nil {
		return l.conn.Close()
	}
	return nil
}

func (l *masterLink) safeUpdateConn(target string, cn client.Conn, state uint32) error {
	var closable client.Conn

	l.mu.Lock()
	if target == l.target {
		closable, l.conn = l.conn, cn
		l.active = target
		l.state = state
	} else {
		closable = cn
	}
	l.mu.Unlock()

	if closable != nil {
		return closable.Close()
	}
	return nil
}
