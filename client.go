package redeo

import (
	"fmt"
	"sync"
	"time"
)

// A client is the origin of a request
type Client struct {
	baseInfo
	ID         int         `json:"-"`
	RemoteAddr string      `json:"remote_addr,omitempty"`
	Ctx        interface{} `json:"ctx,omitempty"`

	cmd   string
	mutex sync.Mutex
}

type clientSlice []Client

func (p clientSlice) Len() int           { return len(p) }
func (p clientSlice) Less(i, j int) bool { return p[i].ID < p[j].ID }
func (p clientSlice) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }

// NewClient creates a new client info container
func NewClient(addr string) *Client {
	return &Client{
		RemoteAddr: addr,
		baseInfo:   baseInfo{StartTime: time.Now()},
	}
}

// Called callback to set last user command
func (i *Client) Called(cmd string) {
	i.mutex.Lock()
	defer i.mutex.Unlock()

	i.cmd = cmd
}

// Command returns the last user command
func (i *Client) LastCommand() string {
	i.mutex.Lock()
	defer i.mutex.Unlock()

	return i.cmd
}

// String generates an info string
func (i *Client) String() string {
	return fmt.Sprintf("id=%d addr=%s age=%d cmd=%s", i.ID, i.RemoteAddr, i.Uptime()/time.Second, i.LastCommand())
}
