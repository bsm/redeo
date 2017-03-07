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

func ExamplePing() {
	srv := redeo.NewServer(nil)
	srv.Handle("ping", redeo.Ping())
}

func ExampleInfo() {
	srv := redeo.NewServer(nil)
	srv.Handle("info", redeo.Info(srv))
}

func ExampleHandlerFunc() {
	mu := sync.RWMutex{}
	myData := make(map[string]map[string]string)
	srv := redeo.NewServer(nil)

	// handle HSET
	srv.HandleFunc("hset", func(w resp.ResponseWriter, c *resp.Command) {
		// validate arguments
		if c.ArgN() != 3 {
			w.AppendError(redeo.WrongNumberOfArgs(c.Name))
			return
		}

		// lock for write
		mu.Lock()
		defer mu.Unlock()

		// fetch (find-or-create) key
		hash, ok := myData[c.Arg(0).String()]
		if !ok {
			hash = make(map[string]string)
			myData[c.Arg(0).String()] = hash
		}

		// check if field already exists
		_, ok = hash[c.Arg(1).String()]

		// set field
		hash[c.Arg(1).String()] = c.Arg(2).String()

		// respond
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
