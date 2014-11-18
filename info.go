package redeo

import (
	"net"
	"os"
	"sort"
	"strconv"
	"sync"
	"time"

	"github.com/bsm/redeo/info"
)

type ServerInfo struct {
	registry *info.Registry

	startTime time.Time
	port      string
	socket    string
	pid       int

	clients     map[uint64]*Client
	connections *info.Counter
	commands    *info.Counter

	mutex sync.RWMutex
}

// newServerInfo creates a new server info container
func newServerInfo(config *Config) *ServerInfo {

	info := &ServerInfo{
		registry:    info.New(),
		startTime:   time.Now(),
		connections: info.NewCounter(),
		commands:    info.NewCounter(),
		clients:     make(map[uint64]*Client),
	}
	return info.withDefaults(config)
}

// Apply default info
func (i *ServerInfo) withDefaults(config *Config) *ServerInfo {
	_, port, _ := net.SplitHostPort(config.Addr)

	server := i.Section("Server")
	server.Register("process_id", info.PlainInt(os.Getpid()))
	server.Register("tcp_port", info.PlainString(port))
	server.Register("unix_socket", info.PlainString(config.Socket))
	server.Register("uptime_in_seconds", info.Callback(func() string {
		d := time.Now().Sub(i.startTime) / time.Second
		return strconv.FormatInt(int64(d), 10)
	}))
	server.Register("uptime_in_days", info.Callback(func() string {
		d := time.Now().Sub(i.startTime) / time.Hour / 24
		return strconv.FormatInt(int64(d), 10)
	}))

	clients := i.Section("Clients")
	clients.Register("connected_clients", info.Callback(func() string {
		return strconv.Itoa(i.ClientsLen())
	}))

	stats := i.Section("Stats")
	stats.Register("total_connections_received", i.connections)
	stats.Register("total_commands_processed", i.commands)

	return i
}

// Section finds-or-creates an info section
func (i *ServerInfo) Section(name string) *info.Section { return i.registry.Section(name) }

// String generates an info string
func (i *ServerInfo) String() string { return i.registry.String() }

// OnConnect callback to register client connection
func (i *ServerInfo) OnConnect(client *Client) {
	i.connections.Inc(1)

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
	i.commands.Inc(1)
	if client != nil {
		client.OnCommand(cmd)
	}
}

// Clients returns connected clients
func (i *ServerInfo) Clients() []Client {
	i.mutex.RLock()
	clients := make(clientSlice, 0, len(i.clients))
	for _, c := range i.clients {
		clients = append(clients, *c)
	}
	i.mutex.RUnlock()

	sort.Sort(clients)
	return []Client(clients)
}

// ClientsLen returns the number of connected clients
func (i *ServerInfo) ClientsLen() int {
	i.mutex.RLock()
	defer i.mutex.RUnlock()
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

// TotalConnections returns the total number of connections made since the
// start of the server.
func (i *ServerInfo) TotalConnections() int64 {
	return i.connections.Value()
}

// TotalCommands returns the total number of commands executed since the start
// of the server.
func (i *ServerInfo) TotalCommands() int64 {
	return i.commands.Value()
}
