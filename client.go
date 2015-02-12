package redeo

import (
	"fmt"
	"net"
	"sync"
	"sync/atomic"
	"time"
)

type clientSlice []Client

func (p clientSlice) Len() int           { return len(p) }
func (p clientSlice) Less(i, j int) bool { return p[i].ID < p[j].ID }
func (p clientSlice) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }

var clientInc = uint64(0)

// A client is the origin of a request
type Client struct {
	ID         uint64      `json:"id,omitempty"`
	RemoteAddr string      `json:"remote_addr,omitempty"`
	Ctx        interface{} `json:"ctx,omitempty"`
	conn       RemoteAddrCloser

	closed      bool
	firstAccess time.Time
	lastAccess  time.Time
	lastCommand string
	mutex       sync.Mutex
}

type RemoteAddrCloser interface {
	RemoteAddr() net.Addr
	Close() error
}

// NewClient creates a new client info container
func NewClient(conn RemoteAddrCloser) *Client {
	now := time.Now()
	return &Client{
		firstAccess: now,
		lastAccess:  now,
		ID:          atomic.AddUint64(&clientInc, 1),
		RemoteAddr:  conn.RemoteAddr().String(),
		conn:        conn,
	}
}

// Close will close client
func (i *Client) Close() {
	i.mutex.Lock()
	defer i.mutex.Unlock()
	i.conn.Close()
}

// OnCommand callback to track user command
func (i *Client) OnCommand(cmd string) {
	i.mutex.Lock()
	defer i.mutex.Unlock()

	i.lastAccess = time.Now()
	i.lastCommand = cmd
}

// String generates an info string
func (i *Client) String() string {
	i.mutex.Lock()
	cmd := i.lastCommand
	atime := i.lastAccess
	i.mutex.Unlock()

	now := time.Now()
	age := now.Sub(i.firstAccess) / time.Second
	idle := now.Sub(atime) / time.Second

	return fmt.Sprintf("id=%d addr=%s age=%d idle=%d cmd=%s", i.ID, i.RemoteAddr, age, idle, cmd)
}
