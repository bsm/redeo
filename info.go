package redeo

import (
	cryptorand "crypto/rand"
	"encoding/hex"
	"fmt"
	mathrand "math/rand"
	"os"
	"sort"
	"strconv"
	"sync"
	"time"

	"github.com/bsm/redeo/v2/info"
)

// CommandDescription describes supported commands
type CommandDescription struct {
	// Name is the command name, returned as a lowercase string.
	Name string

	// Arity is the command arity specification.
	// https://redis.io/commands/command#command-arity.
	// It follows a simple pattern:
	//   positive if command has fixed number of required arguments.
	//   negative if command has minimum number of required arguments, but may have more.
	Arity int64

	// Flags is an enumeration of command flags.
	// https://redis.io/commands/command#flags.
	Flags []string

	// FirstKey is the position of first key in argument list.
	// https://redis.io/commands/command#first-key-in-argument-list
	FirstKey int64

	// LastKey is the position of last key in argument list.
	// https://redis.io/commands/command#last-key-in-argument-list
	LastKey int64

	// KeyStepCount is the step count for locating repeating keys.
	// https://redis.io/commands/command#step-count
	KeyStepCount int64
}

// --------------------------------------------------------------------

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

	startTime   time.Time
	clients     clientStats
	connections *info.IntValue
	commands    *info.IntValue
}

// newServerInfo creates a new server info container
func newServerInfo() *ServerInfo {
	info := &ServerInfo{
		registry:    info.New(),
		startTime:   time.Now(),
		connections: info.NewIntValue(0),
		commands:    info.NewIntValue(0),
		clients:     clientStats{stats: make(map[uint64]*ClientInfo)},
	}
	info.initDefaults()
	return info
}

// ------------------------------------------------------------------------

// Fetch finds or creates an info section. This method is not thread-safe.
func (i *ServerInfo) Fetch(name string) *info.Section { return i.registry.FetchSection(name) }

// Find finds an info section by name.
func (i *ServerInfo) Find(name string) *info.Section { return i.registry.FindSection(name) }

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
func (i *ServerInfo) initDefaults() {
	runID := make([]byte, 20)
	if _, err := cryptorand.Read(runID); err != nil {
		_, _ = mathrand.Read(runID)
	}

	server := i.Fetch("Server")
	server.Register("process_id", info.StaticInt(int64(os.Getpid())))
	server.Register("run_id", info.StaticString(hex.EncodeToString(runID)))
	server.Register("uptime_in_seconds", info.Callback(func() string {
		d := time.Since(i.startTime) / time.Second
		return strconv.FormatInt(int64(d), 10)
	}))
	server.Register("uptime_in_days", info.Callback(func() string {
		d := time.Since(i.startTime) / time.Hour / 24
		return strconv.FormatInt(int64(d), 10)
	}))

	clients := i.Fetch("Clients")
	clients.Register("connected_clients", info.Callback(func() string {
		return strconv.Itoa(i.NumClients())
	}))

	stats := i.Fetch("Stats")
	stats.Register("total_connections_received", i.connections)
	stats.Register("total_commands_processed", i.commands)
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
