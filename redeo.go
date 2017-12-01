package redeo

import (
	"errors"
	"strings"

	"github.com/bsm/redeo/resp"
)

// UnknownCommand returns an unknown command error string
func UnknownCommand(cmd string) string {
	return "ERR unknown command '" + cmd + "'"
}

// ErrUnknownCommand returns an unknown command error
func ErrUnknownCommand(cmd string) error {
	return errors.New(UnknownCommand(cmd))
}

// WrongNumberOfArgs returns an unknown command error string
func WrongNumberOfArgs(cmd string) string {
	return "ERR wrong number of arguments for '" + cmd + "' command"
}

// ErrWrongNumberOfArgs returns an unknown command error
func ErrWrongNumberOfArgs(cmd string) error {
	return errors.New(WrongNumberOfArgs(cmd))
}

// Ping returns a ping handler.
// https://redis.io/commands/ping
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

// Echo returns an echo handler.
// https://redis.io/commands/echo
func Echo() Handler {
	return HandlerFunc(func(w resp.ResponseWriter, c *resp.Command) {
		switch c.ArgN() {
		case 1:
			w.AppendBulk(c.Arg(0))
		default:
			w.AppendError(WrongNumberOfArgs(c.Name))
		}
	})
}

// Info returns an info handler.
// https://redis.io/commands/info
func Info(s *Server) Handler {
	return HandlerFunc(func(w resp.ResponseWriter, c *resp.Command) {
		w.AppendBulkString(s.Info().String())
	})
}

// Commands returns a command handler.
// https://redis.io/commands/command
func Commands(cmds []CommandDetails) Handler {
	return HandlerFunc(func(w resp.ResponseWriter, c *resp.Command) {
		w.AppendArrayLen(len(cmds))

		for _, cmd := range cmds {
			w.AppendArrayLen(6)
			w.AppendBulkString(strings.ToLower(cmd.Name))
			w.AppendInt(cmd.Arity)
			w.AppendArrayLen(len(cmd.Flags))
			for _, flag := range cmd.Flags {
				w.AppendBulkString(flag)
			}
			w.AppendInt(cmd.FirstKey)
			w.AppendInt(cmd.LastKey)
			w.AppendInt(cmd.KeyStepCount)
		}
	})
}

// SubCommands returns a handler that is parsing sub-commands
type SubCommands map[string]Handler

func (s SubCommands) ServeRedeo(w resp.ResponseWriter, c *resp.Command) {

	// First, check if we have a subcommand
	if c.ArgN() == 0 {
		w.AppendError(WrongNumberOfArgs(c.Name))
		return
	}

	firstArg := c.Arg(0).String()
	if h, ok := s[strings.ToLower(firstArg)]; ok {
		cmd := resp.NewCommand(c.Name+" "+firstArg, c.Args[1:]...)
		cmd.SetContext(c.Context())
		h.ServeRedeo(w, cmd)
		return
	}

	w.AppendError("ERR Unknown " + strings.ToLower(c.Name) + " subcommand '" + firstArg + "'")

}

// --------------------------------------------------------------------

// Handler is an abstract handler interface for responding to commands
type Handler interface {
	// ServeRedeo serves a request.
	ServeRedeo(w resp.ResponseWriter, c *resp.Command)
}

// HandlerFunc is a callback function, implementing Handler.
type HandlerFunc func(w resp.ResponseWriter, c *resp.Command)

// ServeRedeo calls f(w, c).
func (f HandlerFunc) ServeRedeo(w resp.ResponseWriter, c *resp.Command) { f(w, c) }

// --------------------------------------------------------------------

// StreamHandler is an  interface for responding to streaming commands
type StreamHandler interface {
	// ServeRedeoStream serves a streaming request.
	ServeRedeoStream(w resp.ResponseWriter, c *resp.CommandStream)
}

// StreamHandlerFunc is a callback function, implementing Handler.
type StreamHandlerFunc func(w resp.ResponseWriter, c *resp.CommandStream)

// ServeRedeoStream calls f(w, c).
func (f StreamHandlerFunc) ServeRedeoStream(w resp.ResponseWriter, c *resp.CommandStream) { f(w, c) }

// --------------------------------------------------------------------

// CommandHandler is an simplified command handler
type CommandHandler interface {
	// ServeRedeoCommand accepts a command and must return one of:
	//   * nil
	//   * error
	//   * string
	//   * []byte
	//   * bool
	//   * float32, float64
	//   * int, int8, int16, int32, int64
	//   * uint, uint8, uint16, uint32, uint64
	//   * resp.CustomResponse instances
	//   * slices of any of the above typs
	//   * maps containing keys and values of any of the above types
	ServeRedeoCommand(c *resp.Command) interface{}
}

// CommandHandlerFunc is a callback function, implementing Handler.
type CommandHandlerFunc func(*resp.Command) interface{}

// ServeRedeo calls f(w, c).
func (f CommandHandlerFunc) ServeRedeoCommand(c *resp.Command) interface{} { return f(c) }
