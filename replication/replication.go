package replication

import (
	"net"
	"strings"
	"sync"
	"time"

	"github.com/bsm/redeo"
	"github.com/bsm/redeo/internal/util"
	"github.com/bsm/redeo/resp"
)

// Replication handles replication between a master and a slave.
type Replication struct {
	id      string
	addr    string
	backlog *backlog
	config  Config

	dataStore   DataStore
	stableStore StableStore

	master *masterLink // master connection link

	closeOnce sync.Once
	closing   chan struct{}
	closed    chan struct{}
}

// NewReplication inits a replication handler with a local announcement address, a data and a stable store.
func NewReplication(addr string, stableStore StableStore, dataStore DataStore, conf *Config) *Replication {
	// copy and normalize config
	var config Config
	if conf != nil {
		config = *conf
	}
	config.norm()

	repl := &Replication{
		id:          util.GenerateID(20),
		addr:        addr,
		stableStore: stableStore,
		dataStore:   dataStore,
		backlog:     newBacklog(-7, config.BacklogSize),
		master:      &masterLink{timeout: config.DialTimeout},
		config:      config,

		closing: make(chan struct{}),
		closed:  make(chan struct{}),
	}

	// start background loop
	go repl.loop()

	return repl
}

// SlaveOf returns a SlaveOf handler.
// https://redis.io/commands/slaveof
func (r *Replication) SlaveOf() redeo.Handler {
	return redeo.HandlerFunc(func(w resp.ResponseWriter, c *resp.Command) {
		if c.ArgN() != 2 {
			w.AppendError(redeo.WrongNumberOfArgs(c.Name))
			return
		}

		addr := net.JoinHostPort(c.Args[0].String(), c.Args[1].String())
		if strings.ToLower(addr) == "no:one" {
			r.master.SetAddr("")
		} else {
			r.master.SetAddr(addr)
		}
		w.AppendOK()
	})
}

// Close closes the replication and frees resources. This method should be called
// once the server has been closed.
func (r *Replication) Close() error {
	var err error
	r.closeOnce.Do(func() {
		close(r.closing)
		<-r.closed

		err = r.master.Close()
	})
	return err
}

// background loop
func (r *Replication) loop() {
	defer close(r.closed)

	tick := time.NewTicker(time.Second)
	defer tick.Stop()

	for {
		select {
		case <-r.closing:
			return
		case <-tick.C:
			if err := r.cron(); err != nil {
				r.config.Logger.Printf("[ERR] %v", err)
			}
		}
	}
}

// periodic cron
func (r *Replication) cron() error {
	if err := r.master.ManageConnection(); err != nil {
		return err
	}
	return nil
}

/*

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
*/
