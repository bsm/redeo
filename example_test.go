package redeo_test

import (
	"net"
	"sync"

	"github.com/bsm/redeo"
	"github.com/bsm/redeo/resp"
)

func ExampleServer() {
	// Init server and define handlers
	srv := redeo.NewServer(nil)
	srv.HandleFunc("ping", func(w resp.ResponseWriter, _ *resp.Command) {
		w.AppendInlineString("PONG")
	})
	srv.HandleFunc("info", func(w resp.ResponseWriter, _ *resp.Command) {
		w.AppendBulkString(srv.Info().String())
	})

	// Open a new listener
	lis, err := net.Listen("tcp", ":9736")
	if err != nil {
		panic(err)
	}
	defer lis.Close()

	// Start serving (blocking)
	srv.Serve(lis)
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

func ExamplePubSub() {
	broker := redeo.NewPubSubBroker()

	srv := redeo.NewServer(nil)
	srv.Handle("publish", broker.Publish())
	srv.Handle("subscribe", broker.Subscribe())
}

func ExampleHandlerFunc() {
	mu := sync.RWMutex{}
	myData := make(map[string]map[string]string)
	srv := redeo.NewServer(nil)

	// handle HSET
	srv.HandleFunc("hset", func(w resp.ResponseWriter, c *resp.Command) {
		if c.ArgN() != 3 {
			w.AppendError(redeo.WrongNumberOfArgs(c.Name))
			return
		}

		key := c.Arg(0).String()
		field := c.Arg(1).String()
		value := c.Arg(2).String()

		// lock for write
		mu.Lock()
		defer mu.Unlock()

		// fetch hash @ key
		hash, ok := myData[key]
		if !ok {
			hash = make(map[string]string)
			myData[key] = hash
		}

		// check if set and replace
		_, ok = hash[field]
		hash[field] = value

		if ok {
			w.AppendInt(0)
		} else {
			w.AppendInt(1)
		}
	})

	// handle HGET
	srv.HandleFunc("hget", func(w resp.ResponseWriter, c *resp.Command) {
		if c.ArgN() != 2 {
			w.AppendError(redeo.WrongNumberOfArgs(c.Name))
			return
		}

		mu.RLock()
		defer mu.RUnlock()

		hash, ok := myData[c.Arg(0).String()]
		if !ok {
			w.AppendNil()
			return
		}

		val, ok := hash[c.Arg(1).String()]
		if !ok {
			w.AppendNil()
			return
		}

		w.AppendBulkString(val)
	})
}

func ExampleCommandHandlerFunc() {
	data := make(map[string]string)

	srv := redeo.NewServer(nil)
	srv.HandleCommandFunc("get", func(c *resp.Command) interface{} {
		if c.ArgN() != 1 {
			return redeo.ErrWrongNumberOfArgs(c.Name)
		}

		val, ok := data[c.Arg(0).String()]
		if !ok {
			return nil
		}
		return val
	})
}
