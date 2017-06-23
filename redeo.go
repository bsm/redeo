package redeo

import (
	"github.com/bsm/redeo/resp"
)

// UnknownCommand returns an unknown command error string
func UnknownCommand(cmd string) string {
	return "ERR unknown command '" + cmd + "'"
}

// WrongNumberOfArgs returns an unknown command error string
func WrongNumberOfArgs(cmd string) string {
	return "ERR wrong number of arguments for '" + cmd + "' command"
}

// Ping returns a ping handler
func Ping() Handler {
	return HandlerFunc(func(w resp.ResponseWriter, c *resp.Command) {
		switch c.ArgN() {
		case 0:
			w.AppendBulkString("PONG")
		case 1:
			w.AppendBulk(c.Arg(0))
		default:
			w.AppendError(WrongNumberOfArgs(c.Name))
		}
	})
}

// Info returns an info handler
func Info(s *Server) Handler {
	return HandlerFunc(func(w resp.ResponseWriter, c *resp.Command) {
		w.AppendBulkString(s.Info().String())
	})
}

// --------------------------------------------------------------------

// Handler is an abstract handler interface for handling commands
type Handler interface {
	// ServeRedeo serves a request.
	ServeRedeo(w resp.ResponseWriter, c *resp.Command)
}

// HandlerFunc is a callback function, implementing Handler.
type HandlerFunc func(w resp.ResponseWriter, c *resp.Command)

// ServeRedeo calls f(w, c).
func (f HandlerFunc) ServeRedeo(w resp.ResponseWriter, c *resp.Command) { f(w, c) }

// --------------------------------------------------------------------

// StreamHandler is an  interface for handling streaming commands
type StreamHandler interface {
	// ServeRedeoStream serves a streaming request.
	ServeRedeoStream(w resp.ResponseWriter, c *resp.CommandStream)
}

// StreamHandlerFunc is a callback function, implementing Handler.
type StreamHandlerFunc func(w resp.ResponseWriter, c *resp.CommandStream)

// ServeRedeoStream calls f(w, c).
func (f StreamHandlerFunc) ServeRedeoStream(w resp.ResponseWriter, c *resp.CommandStream) { f(w, c) }
