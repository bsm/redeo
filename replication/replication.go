package replication

import (
	"net"
	"strings"
	"sync"
	"time"

	"github.com/bsm/redeo"
	"github.com/bsm/redeo/resp"
)

// Replication handles replication between a master and a slave.
type Replication struct {
	masterID string
	addr     string
	config   Config

	backlog *backlog
	stable  StableStore
	data    DataStore
	master  masterLink // master connection link

	closeOnce sync.Once
	closing   chan struct{}
	waitFor   sync.WaitGroup
}

// NewReplication inits a replication handler with a local announcement address, a data and a stable store.
func NewReplication(addr string, stable StableStore, data DataStore, conf *Config) (*Replication, error) {
	// copy and normalize config
	var config Config
	if conf != nil {
		config = *conf
	}
	config.norm()

	var state runtimeState
	if err := (&state).Load(stable); err != nil {
		return nil, err
	}

	repl := &Replication{
		masterID: state.ID,
		addr:     addr,
		config:   config,

		backlog: newBacklog(state.Offset, config.BacklogSize),
		stable:  stable,
		data:    data,
		master: masterLink{
			timeout: config.DialTimeout,
			target:  state.SlaveOf,
		},

		closing: make(chan struct{}),
	}

	repl.waitFor.Add(2)
	go repl.cronLoop()
	go repl.persistLoop()

	return repl, nil
}

// Propagate propagates a command to slaves.
func (r *Replication) Propagate(cmd *resp.Command) {
	r.backlog.Feed(cmd)
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
func (r *Replication) Close() (err error) {
	r.closeOnce.Do(func() {
		close(r.closing)
		r.waitFor.Wait()

		if e := r.persist(); e != nil {
			err = e
		}
		if e := r.master.Close(); e != nil {
			err = e
		}
	})
	return
}

func (r *Replication) cronLoop() {
	defer r.waitFor.Done()

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

func (r *Replication) persistLoop() {
	defer r.waitFor.Done()

	tick := time.NewTicker(time.Second)
	defer tick.Stop()

	for {
		select {
		case <-r.closing:
			return
		case <-tick.C:
			if err := r.persist(); err != nil {
				r.config.Logger.Printf("[ERR] %v", err)
			}
		}
	}
}

func (r *Replication) cron() error {
	if err := r.master.MaintainConn(); err != nil {
		return err
	}
	return nil
}

func (r *Replication) persist() error {
	_, latest := r.backlog.Offsets()
	state := &runtimeState{
		ID:      r.masterID,
		Offset:  latest,
		SlaveOf: r.master.Addr(),
	}
	if err := state.Save(r.stable); err != nil {
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
