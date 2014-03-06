package redeo

import (
	"bufio"
	"net"
	"strings"
	"sync"
)

// Server configuration
type Server struct {
	config   *Config
	commands map[string]Handler
	mutex    *sync.Mutex
}

// NewServer creates a new server instance
func NewServer(config *Config) *Server {
	if config == nil {
		config = DefaultConfig
	}

	return &Server{
		config:   config,
		commands: make(map[string]Handler),
		mutex:    new(sync.Mutex),
	}
}

// Proto returns the server protocol
func (srv *Server) Proto() string {
	return srv.config.Proto
}

// Addr returns the server address
func (srv *Server) Addr() string {
	return srv.config.Addr
}

// ListenAndServe starts the server
func (srv *Server) ListenAndServe() error {
	lis, err := net.Listen(srv.config.Proto, srv.config.Addr)
	if err != nil {
		return err
	}
	return srv.Serve(lis)
}

// Serve accepts incoming connections on the Listener lis, creating a
// new service goroutine for each.
func (srv *Server) Serve(lis net.Listener) error {
	defer lis.Close()
	for {
		conn, err := lis.Accept()
		if err != nil {
			return err
		}
		go srv.ServeClient(conn)
	}
}

// Handle registers a handler for a command
func (srv *Server) Handle(name string, handler Handler) {
	srv.mutex.Lock()
	defer srv.mutex.Unlock()

	srv.commands[strings.ToLower(name)] = handler
}

// HandleFunc registers a handler callback for a command
func (srv *Server) HandleFunc(name string, callback HandlerFunc) {
	srv.Handle(name, Handler(callback))
}

// Apply applies a request
func (srv *Server) Apply(req *Request) (*Responder, error) {
	cmd, ok := srv.commands[req.Name]
	if !ok {
		return nil, ErrUnknownCommand
	}
	res := NewResponder()
	err := cmd.ServeClient(res, req)
	return res, err
}

// Serve starts a new session, using `conn` as a transport.
func (srv *Server) ServeClient(conn net.Conn) {
	defer conn.Close()

	rd := bufio.NewReader(conn)
	for {
		req, err := ParseRequest(rd)
		if err != nil {
			srv.writeError(conn, err)
			return
		}
		req.RemoteAddr = conn.RemoteAddr()

		res, err := srv.Apply(req)
		if err != nil {
			srv.writeError(conn, err)
			return
		}

		if _, err = res.WriteTo(conn); err != nil {
			return
		}
	}
}

// Serve starts a new session, using `conn` as a transport.
func (srv *Server) writeError(conn net.Conn, err error) {
	res := NewResponder()
	res.WriteError(err)
	res.WriteTo(conn)
}
