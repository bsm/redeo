package redeo_test

import (
	"net"
	"sync"

	"github.com/bsm/redeo/v2"
	"github.com/bsm/redeo/v2/resp"
)

func ExampleServer() {
	srv := redeo.NewServer(nil)

	// Define handlers
	srv.HandleFunc("ping", func(w resp.ResponseWriter, _ *resp.Command) {
		w.AppendInlineString("PONG")
	})
	srv.HandleFunc("info", func(w resp.ResponseWriter, _ *resp.Command) {
		w.AppendBulkString(srv.Info().String())
	})

	// More handlers; demo usage of redeo.WrapperFunc
	srv.Handle("echo", redeo.WrapperFunc(func(c *resp.Command) interface{} {
		if c.ArgN() != 1 {
			return redeo.ErrWrongNumberOfArgs(c.Name)
		}
		return c.Arg(0)
	}))

	// Open a new listener
	lis, err := net.Listen("tcp", ":9736")
	if err != nil {
		panic(err)
	}
	defer lis.Close()

	// Start serving (blocking)
	if err := srv.Serve(lis); err != nil {
		panic(err)
	}
}

func ExampleClient() {
	srv := redeo.NewServer(nil)
	srv.HandleFunc("myip", func(w resp.ResponseWriter, cmd *resp.Command) {
		client := redeo.GetClient(cmd.Context())
		if client == nil {
			w.AppendNil()
			return
		}
		w.AppendInlineString(client.RemoteAddr().String())
	})
}

func ExamplePing() {
	srv := redeo.NewServer(nil)
	srv.Handle("ping", redeo.Ping())
}

func ExampleInfo() {
	srv := redeo.NewServer(nil)
	srv.Handle("info", redeo.Info(srv))
}

func ExampleCommandDescriptions() {
	srv := redeo.NewServer(nil)
	srv.Handle("command", redeo.CommandDescriptions{
		{Name: "get", Arity: 2, Flags: []string{"readonly", "fast"}, FirstKey: 1, LastKey: 1, KeyStepCount: 1},
		{Name: "randomkey", Arity: 1, Flags: []string{"readonly", "random"}},
		{Name: "mset", Arity: -3, Flags: []string{"write", "denyoom"}, FirstKey: 1, LastKey: -1, KeyStepCount: 2},
		{Name: "quit", Arity: 1},
	})
}

func ExampleSubCommands() {
	srv := redeo.NewServer(nil)
	srv.Handle("custom", redeo.SubCommands{
		"ping": redeo.Ping(),
		"echo": redeo.Echo(),
	})
}

func ExamplePubSubBroker() {
	broker := redeo.NewPubSubBroker()

	srv := redeo.NewServer(nil)
	srv.Handle("publish", broker.Publish())
	srv.Handle("subscribe", broker.Subscribe())
}

func ExampleHandlerFunc() {
	mu := sync.RWMutex{}
	data := make(map[string]string)
	srv := redeo.NewServer(nil)

	srv.HandleFunc("set", func(w resp.ResponseWriter, c *resp.Command) {
		if c.ArgN() != 2 {
			w.AppendError(redeo.WrongNumberOfArgs(c.Name))
			return
		}

		key := c.Arg(0).String()
		val := c.Arg(1).String()

		mu.Lock()
		data[key] = val
		mu.Unlock()

		w.AppendInt(1)
	})

	srv.HandleFunc("get", func(w resp.ResponseWriter, c *resp.Command) {
		if c.ArgN() != 1 {
			w.AppendError(redeo.WrongNumberOfArgs(c.Name))
			return
		}

		key := c.Arg(0).String()
		mu.RLock()
		val, ok := data[key]
		mu.RUnlock()

		if ok {
			w.AppendBulkString(val)
			return
		}
		w.AppendNil()
	})
}

func ExampleWrapperFunc() {
	mu := sync.RWMutex{}
	data := make(map[string]string)
	srv := redeo.NewServer(nil)

	srv.Handle("set", redeo.WrapperFunc(func(c *resp.Command) interface{} {
		if c.ArgN() != 2 {
			return redeo.ErrWrongNumberOfArgs(c.Name)
		}

		key := c.Arg(0).String()
		val := c.Arg(1).String()

		mu.Lock()
		data[key] = val
		mu.Unlock()

		return 1
	}))

	srv.Handle("get", redeo.WrapperFunc(func(c *resp.Command) interface{} {
		if c.ArgN() != 1 {
			return redeo.ErrWrongNumberOfArgs(c.Name)
		}

		key := c.Arg(0).String()
		mu.RLock()
		val, ok := data[key]
		mu.RUnlock()

		if ok {
			return val
		}
		return nil
	}))
}
