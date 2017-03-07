package redeo

import (
	"fmt"
	"os"
	"sort"
	"strconv"
	"sync"
	"time"

	"github.com/bsm/redeo/info"
)

// ClientInfo contains client stats
type ClientInfo struct {
	// ID is the internal client ID
	ID uint64

	// RemoteAddr is the remote address string
	RemoteAddr string

	// LastCmd is the last command called by this client
	LastCmd string

	// CreateTime returns the time at which the client has
	// connected to the server
	CreateTime time.Time

	// AccessTime returns the time of the last access
	AccessTime time.Time
}

func newClientInfo(c *Client, now time.Time) *ClientInfo {
	return &ClientInfo{
		ID:         c.id,
		RemoteAddr: c.RemoteAddr().String(),
		CreateTime: now,
		AccessTime: now,
	}
}

// String generates an info string
func (i *ClientInfo) String() string {
	now := time.Now()
	return fmt.Sprintf("id=%d addr=%s age=%d idle=%d cmd=%s",
		i.ID,
		i.RemoteAddr,
		now.Sub(i.CreateTime)/time.Second,
		now.Sub(i.AccessTime)/time.Second,
		i.LastCmd,
	)
}

// --------------------------------------------------------------------

// ServerInfo contains server stats
type ServerInfo struct {
	registry *info.Registry

	startTime time.Time
	port      string
	socket    string
	pid       int

	clients     clientStats
	connections *info.Counter
	commands    *info.Counter
}

// newServerInfo creates a new server info container
func newServerInfo() *ServerInfo {
	info := &ServerInfo{
		registry:    info.New(),
		startTime:   time.Now(),
		connections: info.NewCounter(),
		commands:    info.NewCounter(),
		clients:     clientStats{stats: make(map[uint64]*ClientInfo)},
	}
	return info.withDefaults()
}

// ------------------------------------------------------------------------

// Section finds-or-creates an info section
func (i *ServerInfo) Section(name string) *info.Section { return i.registry.Section(name) }

// String generates an info string
func (i *ServerInfo) String() string { return i.registry.String() }

// NumClients returns the number of connected clients
func (i *ServerInfo) NumClients() int { return i.clients.Len() }

// ClientInfo returns details about connected clients
func (i *ServerInfo) ClientInfo() []ClientInfo { return i.clients.All() }

// TotalConnections returns the total number of connections made since the
// start of the server.
func (i *ServerInfo) TotalConnections() int64 { return i.connections.Value() }

// TotalCommands returns the total number of commands executed since the start
// of the server.
func (i *ServerInfo) TotalCommands() int64 { return i.commands.Value() }

// Apply default info
func (i *ServerInfo) withDefaults() *ServerInfo {
	server := i.Section("Server")
	server.Register("process_id", info.IntValue(os.Getpid()))
	server.Register("uptime_in_seconds", info.Callback(func() string {
		d := time.Since(i.startTime) / time.Second
		return strconv.FormatInt(int64(d), 10)
	}))
	server.Register("uptime_in_days", info.Callback(func() string {
		d := time.Since(i.startTime) / time.Hour / 24
		return strconv.FormatInt(int64(d), 10)
	}))

	clients := i.Section("Clients")
	clients.Register("connected_clients", info.Callback(func() string {
		return strconv.Itoa(i.NumClients())
	}))

	stats := i.Section("Stats")
	stats.Register("total_connections_received", i.connections)
	stats.Register("total_commands_processed", i.commands)

	return i
}

func (i *ServerInfo) register(c *Client) {
	i.clients.Add(c)
	i.connections.Inc(1)
}

func (i *ServerInfo) deregister(clientID uint64) {
	i.clients.Del(clientID)
}

func (i *ServerInfo) command(clientID uint64, cmd string) {
	i.clients.Cmd(clientID, cmd)
	i.commands.Inc(1)
}

// --------------------------------------------------------------------

type clientStats struct {
	stats map[uint64]*ClientInfo
	mu    sync.RWMutex
}

func (s *clientStats) Add(c *Client) {
	info := newClientInfo(c, time.Now())
	s.mu.Lock()
	s.stats[c.id] = info
	s.mu.Unlock()
}

func (s *clientStats) Cmd(clientID uint64, cmd string) {
	s.mu.Lock()
	if info, ok := s.stats[clientID]; ok {
		info.AccessTime = time.Now()
		info.LastCmd = cmd
	}
	s.mu.Unlock()
}

func (s *clientStats) Del(clientID uint64) {
	s.mu.Lock()
	delete(s.stats, clientID)
	s.mu.Unlock()
}

func (s *clientStats) Len() int {
	s.mu.RLock()
	n := len(s.stats)
	s.mu.RUnlock()
	return n
}

func (s *clientStats) All() []ClientInfo {
	s.mu.RLock()
	defer s.mu.RUnlock()

	res := make(clientInfoSlice, 0, len(s.stats))
	for _, info := range s.stats {
		res = append(res, *info)
	}
	sort.Sort(res)
	return res
}

type clientInfoSlice []ClientInfo

func (p clientInfoSlice) Len() int           { return len(p) }
func (p clientInfoSlice) Less(i, j int) bool { return p[i].ID < p[j].ID }
func (p clientInfoSlice) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }
