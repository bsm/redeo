package redeo

import (
	"fmt"
	"net"
	"os"
	"sync"
	"sync/atomic"
	"time"
)

type Info struct {
	StartTime time.Time
	Port      string // tcp_port
	Socket    string // socket
	ProcessID int    // process_id

	clients     int64
	connections int64
	processed   int64

	mutex sync.Mutex
}

// NewInfo creates a new info container
func NewInfo(config *Config) *Info {
	_, port, _ := net.SplitHostPort(config.Addr)
	return &Info{
		StartTime: time.Now(),
		Port:      port,
		Socket:    config.Socket,
		ProcessID: os.Getpid(),
	}
}

// Uptime returns the duration since server start
func (i *Info) Uptime() time.Duration {
	return time.Now().Sub(i.StartTime)
}

// Clients returns the number of connected clients
func (i *Info) Clients() int64 {
	return atomic.LoadInt64(&i.clients)
}

// Connections returns the number of total connections since start
func (i *Info) Connections() int64 {
	return atomic.LoadInt64(&i.connections)
}

// Processed returns the number of processed commands since start
func (i *Info) Processed() int64 {
	return atomic.LoadInt64(&i.processed)
}

// Connected callback to increment clients & connections
func (i *Info) Connected() {
	atomic.AddInt64(&i.connections, 1)
	atomic.AddInt64(&i.clients, 1)
}

// Disconnected callback to decerement client count
func (i *Info) Disconnected() {
	atomic.AddInt64(&i.clients, -1)
}

// Called callback to increment processed command count
func (i *Info) Called() {
	atomic.AddInt64(&i.processed, 1)
}

// String generates an info string
func (i *Info) String() string {
	uptime := i.Uptime()
	return fmt.Sprintf(
		"# Server\nprocess_id:%d\ntcp_port:%s\nunix_socket:%s\nuptime_in_seconds:%d\nuptime_in_days:%d\n\n"+
			"# Clients\nconnected_clients:%d\n\n"+
			"# Stats\ntotal_connections_received:%d\ntotal_commands_processed:%d\n",
		i.ProcessID, i.Port, i.Socket, uptime/time.Second, uptime/time.Hour/24,
		i.Clients(),
		i.Connections(), i.Processed(),
	)
}
