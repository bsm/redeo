package replication

import (
	"context"
	"fmt"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/bsm/redeo"
	"github.com/bsm/redeo/internal/util"
	"github.com/bsm/redeo/resp"
)

const (
	replStateNone        uint32 = iota // no replication
	replStateMustConnect               // must connect to master
	replStateConnecting                // establishing a connection
	replStateConnected                 // connected
)

// Replication handles replication between a master and a slave.
type Replication struct {
	id      string
	addr    string
	server  *redeo.Server
	backlog *backlog
	config  Config

	// master connection
	replState   uint32
	masterAddr  string
	masterConn  *masterConn
	transitions interface{}

	closeOnce sync.Once
	closing   chan struct{}
	closed    chan struct{}
}

// NewReplication inits a replication handler with a local announcement address
// and installs itself on a server.
func NewReplication(addr string, server *redeo.Server, conf *Config) *Replication {
	// copy and normalize config
	var config Config
	if conf != nil {
		config = *conf
	}
	config.norm()

	repl := &Replication{
		id:      util.GenerateID(20),
		addr:    addr,
		server:  server,
		backlog: newBacklog(-7, config.BacklogSize),
		config:  config,

		transitions: make(interface{}, 1024),
		closing:     make(chan struct{}),
		closed:      make(chan struct{}),
	}

	// start background loop
	go repl.loop()

	return repl
}

// SlaveOf returns a SlaveOf handler.
// https://redis.io/commands/slaveof
func (r *Replication) SlaveOf() Handler {
	return HandlerFunc(func(w resp.ResponseWriter, c *resp.Command) {
		if c.ArgN() != 2 {
			w.AppendError(WrongNumberOfArgs(c.Name))
			return
		}

		addr := net.JoinHostPort(c.Args[0].String(), c.Args[1].String())
		if strings.ToLower(addr) == "no:one" {
			r.transitions <- &transitionToMaster{}
		} else {
			r.transitions <- &transitionToSlave{MasterAddr: addr}
		}
		w.AppendOK()
	})
}

// Close closes the replication and frees resources. This method should be called
// once the server has been closed.
func (r *Replication) Close() error {
	r.closeOnce.Do(func() {
		close(r.closing)
		<-r.closed
	})
	return nil
}

// background loop
func (r *Replication) loop() {
	defer close(r.closed)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	tick := time.NewTicker(time.Second)
	defer tick.Stop()

	for {
		select {
		case <-r.closing:
			return
		case t := <-r.transitions:
			switch tt := t.(type) {
			case *transitionToMaster:
				r.becomeMaster()
			case *transitionToSlave:
				r.becomeSlaveOf(tt.MasterAddr)
			}
		case <-tick.C:
			if err := r.cron(ctx); err != nil {
				r.config.Logger.Printf("[ERR] %v", err)
			}
		}
	}
}

// periodic cron
func (r *Replication) cron() error {
	switch r.replState {
	case replStateMustConnect:
		cn, err := newMasterConn(r.masterAddr, r.config.DialTimeout)
		if err != nil {
			return fmt.Errorf("unable to connect to master: %v", err)
		}

		r.masterConn = cn
		r.replState = replStateConnecting
	case replStateConnecting:

	}
}

func (r *Replication) becomeMaster() {
	if r.masterConn != nil {
		_ = r.masterConn.Close()
		r.masterConn = nil
	}
	if r.masterAddr != "" {
		r.masterAddr = ""
	}
	r.replState = replStateNone
}

func (r *Replication) becomeSlave(masterAddr string) {
	// don't do anything if already connected/connecting to the master
	if r.masterAddr == masterAddr {
		return
	}

	// disconnect from previous master
	if r.masterAddr != "" {
		r.becomeMaster()
	}

	// set master address and state
	r.masterAddr = masterAddr
	r.replState = replStateMustConnect
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
