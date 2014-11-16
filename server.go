package redeo

import (
	"bufio"
	"io"
	"net"
	"os"
	"strings"
	"sync"
	"time"
)

// Server configuration
type Server struct {
	config   *Config
	info     *ServerInfo
	commands map[string]Handler
	mutex    sync.Mutex
}

// NewServer creates a new server instance
func NewServer(config *Config) *Server {
	if config == nil {
		config = DefaultConfig
	}

	return &Server{
		config:   config,
		info:     newServerInfo(config),
		commands: make(map[string]Handler),
	}
}

// Addr returns the server TCP address
func (srv *Server) Addr() string {
	return srv.config.Addr
}

// Addr returns the server Socket address
func (srv *Server) Socket() string {
	return srv.config.Socket
}

// Info returns the server info registry
func (srv *Server) Info() *ServerInfo {
	return srv.info
}

// ListenAndServe starts the server
func (srv *Server) ListenAndServe() error {
	errs := make(chan error)

	if srv.Addr() != "" {
		lis, err := net.Listen("tcp", srv.Addr())
		if err != nil {
			return err
		}
		go srv.Serve(errs, lis)
	}

	if srv.Socket() != "" {
		lis, err := srv.listenUnix()
		if err != nil {
			return err
		}
		go srv.Serve(errs, lis)
	}

	return <-errs
}

// Serve accepts incoming connections on the Listener lis, creating a
// new service goroutine for each.
func (srv *Server) Serve(errs chan error, lis net.Listener) {
	defer lis.Close()

	for {
		conn, err := lis.Accept()
		if err != nil {
			errs <- err
			return
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
		return nil, UnknownCommand(req.Name)
	}

	srv.info.OnCommand(req.client, req.Name)
	res := NewResponder()
	err := cmd.ServeClient(res, req)
	return res, err
}

// Serve starts a new session, using `conn` as a transport.
func (srv *Server) ServeClient(conn net.Conn) {
	defer conn.Close()

	if alive := srv.config.TCPKeepAlive; alive > 0 {
		if tcpconn, ok := conn.(*net.TCPConn); ok {
			tcpconn.SetKeepAlive(true)
			tcpconn.SetKeepAlivePeriod(alive)
		}
	}

	buffer := bufio.NewReader(conn)
	client := NewClient(conn.RemoteAddr().String())
	srv.info.OnConnect(client)
	defer srv.info.OnDisconnect(client)

	for {
		if timeout := srv.config.Timeout; timeout > 0 {
			conn.SetReadDeadline(time.Now().Add(timeout))
		}

		req, err := ParseRequest(buffer)
		if err != nil {
			srv.writeError(conn, err)
			return
		}
		req.client = client

		res, err := srv.Apply(req)
		if err != nil {
			srv.writeError(conn, err)
			// Don't disconnect clients on simple command errors to allow pipelining
			if _, ok := err.(ClientError); ok {
				continue
			}
			return
		}

		if _, err = res.WriteTo(conn); err != nil {
			return
		}

		if client.closed {
			return
		}
	}
}

// Serve starts a new session, using `conn` as a transport.
func (srv *Server) writeError(conn net.Conn, err error) {
	// Don't try to respond on EOFs
	if err == io.EOF {
		return
	}
	res := NewResponder()
	res.WriteError(err)
	res.WriteTo(conn)
}

// listenUnix starts the unix listener on socket path
func (srv *Server) listenUnix() (net.Listener, error) {
	if stat, err := os.Stat(srv.Socket()); !os.IsNotExist(err) && !stat.IsDir() {
		if err = os.RemoveAll(srv.Socket()); err != nil {
			return nil, err
		}
	}
	return net.Listen("unix", srv.Socket())
}
