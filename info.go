package redeo

import (
	"fmt"
	"net"
	"os"
	"sort"
	"sync"
	"time"
)

type baseInfo struct {
	StartTime time.Time `json:"-"`
}

// Uptime returns the duration since server start
func (i *baseInfo) Uptime() time.Duration {
	return time.Now().Sub(i.StartTime)
}

type ServerInfo struct {
	baseInfo
	Port      string // tcp_port
	Socket    string // socket
	ProcessID int    // process_id

	clients     map[string]*Client
	connections int64
	processed   int64

	mutex sync.Mutex
}

// NewServerInfo creates a new server info container
func NewServerInfo(config *Config) *ServerInfo {
	_, port, _ := net.SplitHostPort(config.Addr)

	return &ServerInfo{
		Port:      port,
		Socket:    config.Socket,
		ProcessID: os.Getpid(),
		clients:   make(map[string]*Client),
		baseInfo:  baseInfo{StartTime: time.Now()},
	}
}

// Connections returns the number of total connections since start
func (i *ServerInfo) Connections() int64 {
	i.mutex.Lock()
	defer i.mutex.Unlock()

	return i.connections
}

// Processed returns the number of processed commands since start
func (i *ServerInfo) Processed() int64 {
	i.mutex.Lock()
	defer i.mutex.Unlock()

	return i.processed
}

// Connected callback to increment clients & connections
func (i *ServerInfo) Connected(client *Client) {
	i.mutex.Lock()
	defer i.mutex.Unlock()

	i.connections++
	i.clients[client.RemoteAddr] = client
	client.ID = len(i.clients)
}

// Disconnected callback to decerement client count
func (i *ServerInfo) Disconnected(client *Client) {
	i.mutex.Lock()
	defer i.mutex.Unlock()

	delete(i.clients, client.RemoteAddr)
}

// Called callback to increment processed command count
func (i *ServerInfo) Called(client *Client, cmd string) {
	i.mutex.Lock()
	defer i.mutex.Unlock()

	i.processed++
	client.Called(cmd)
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

// ClientLen returns the number of connected clients
func (i *ServerInfo) ClientLen() int {
	i.mutex.Lock()
	defer i.mutex.Unlock()

	return len(i.clients)
}

// String generates an info string
func (i *ServerInfo) String() string {
	uptime := i.Uptime()

	i.mutex.Lock()
	clients := len(i.clients)
	connections := i.connections
	processed := i.processed
	i.mutex.Unlock()

	return fmt.Sprintf(
		"# Server\nprocess_id:%d\ntcp_port:%s\nunix_socket:%s\nuptime_in_seconds:%d\nuptime_in_days:%d\n\n"+
			"# Clients\nconnected_clients:%d\n\n"+
			"# Stats\ntotal_connections_received:%d\ntotal_commands_processed:%d\n",
		i.ProcessID, i.Port, i.Socket, uptime/time.Second, uptime/time.Hour/24,
		clients,
		connections, processed,
	)
}

// ClientsString generates a client list
func (i *ServerInfo) ClientsString() string {
	str := ""
	for _, client := range i.Clients() {
		str += client.String() + "\n"
	}
	return str
}
