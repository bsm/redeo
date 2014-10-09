package redeo

import (
	"fmt"
	"net"
	"os"
	"sort"
	"sync"
	"sync/atomic"
	"time"
)

type baseInfo struct {
	StartTime time.Time `json:"-"`
}

// UpTime returns the duration since server start
func (i *baseInfo) UpTime() time.Duration {
	return time.Now().Sub(i.StartTime)
}

type ServerInfo struct {
	baseInfo
	Port      string // tcp_port
	Socket    string // socket
	ProcessID int    // process_id

	clients     map[uint64]*Client
	connections uint64
	processed   uint64

	mutex sync.Mutex
}

// NewServerInfo creates a new server info container
func NewServerInfo(config *Config) *ServerInfo {
	_, port, _ := net.SplitHostPort(config.Addr)

	return &ServerInfo{
		Port:      port,
		Socket:    config.Socket,
		ProcessID: os.Getpid(),
		clients:   make(map[uint64]*Client),
		baseInfo:  baseInfo{StartTime: time.Now()},
	}
}

// OnConnect callback to register client connection
func (i *ServerInfo) OnConnect(client *Client) {
	atomic.AddUint64(&i.connections, 1)

	i.mutex.Lock()
	defer i.mutex.Unlock()

	i.clients[client.ID] = client
}

// OnDisconnect callback to de-register client
func (i *ServerInfo) OnDisconnect(client *Client) {
	i.mutex.Lock()
	defer i.mutex.Unlock()

	delete(i.clients, client.ID)
}

// OnCommand callback to track processed command
func (i *ServerInfo) OnCommand(client *Client, cmd string) {
	atomic.AddUint64(&i.processed, 1)
	if client != nil {
		client.OnCommand(cmd)
	}
}

// TotalConnections returns the number of total connections since start
func (i *ServerInfo) TotalConnections() uint64 {
	return atomic.LoadUint64(&i.connections)
}

// TotalProcessed returns the number of processed commands since start
func (i *ServerInfo) TotalProcessed() uint64 {
	return atomic.LoadUint64(&i.processed)
}

// Clients returns connected clients
func (i *ServerInfo) Clients() []Client {
	i.mutex.Lock()
	clients := make(clientSlice, 0, len(i.clients))
	for _, c := range i.clients {
		clients = append(clients, *c)
	}
	i.mutex.Unlock()

	sort.Sort(clients)
	return []Client(clients)
}

// ClientsLen returns the number of connected clients
func (i *ServerInfo) ClientsLen() int {
	i.mutex.Lock()
	defer i.mutex.Unlock()

	return len(i.clients)
}

// ClientsString generates a client list
func (i *ServerInfo) ClientsString() string {
	str := ""
	for _, client := range i.Clients() {
		str += client.String() + "\n"
	}
	return str
}

// String generates an info string
func (i *ServerInfo) String() string {
	uptime := i.UpTime()

	return fmt.Sprintf(
		"# Server\nprocess_id:%d\ntcp_port:%s\nunix_socket:%s\nuptime_in_seconds:%d\nuptime_in_days:%d\n\n"+
			"# Clients\nconnected_clients:%d\n\n"+
			"# Stats\ntotal_connections_received:%d\ntotal_commands_processed:%d\n",
		i.ProcessID, i.Port, i.Socket, uptime/time.Second, uptime/time.Hour/24,
		i.ClientsLen(),
		i.TotalConnections(), i.TotalProcessed(),
	)
}
