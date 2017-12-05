/*
Package redeo provides a toolkit for building redis-protocol compatible services
optimised for high thoughput and low latency.

A simple server example with two commands:

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
	srv.Serve(lis)

*/
package redeo
