package replication

import (
	"sync/atomic"

	"github.com/bsm/redeo"
)

const (
	roleMaster = iota
	roleSlave
)

// Replication handles replication between a master and a slave.
type Replication struct {
	addr string
	role uint32

	info *redeo.ServerInfo
}

// NewReplication inits a replication handler with a local announcement address
// and installs itself on a server.
func NewReplication(addr string, server *redeo.Server) *Replication {
	repl := &Replication{
		addr: addr,
		role: roleMaster,
		info: server.Info(),
	}
	repl.updateInfo()
	return repl
}

// Returns "master" or "slave"
func (r *Replication) roleName() string {
	switch atomic.LoadUint32(&r.role) {
	case roleSlave:
		return "slave"
	default:
		return "master"
	}
}

func (r *Replication) updateInfo() {
	// r.info.Fetch("Replication").Replace(func(s *info.Section) {
	// 	s.Register("role", info.Callback(r.roleName))
	// 	s.Register("connected_slaves", info.StaticString("0"))
	// 	// s.Register("master_repl_offset", repl.masterOffset)
	// })
}
