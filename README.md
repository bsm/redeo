# Redeo

[![Go Reference](https://pkg.go.dev/badge/github.com/bsm/redeo/v2)](https://pkg.go.dev/github.com/bsm/redeo/v2)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)

The high-performance Swiss Army Knife for building redis-protocol compatible servers/services.

## Parts

This repository is organised into multiple components:

* [root](./) package contains the framework for building redis-protocol compatible,
  high-performance servers.
* [resp](./resp/) implements low-level primitives for dealing with
  RESP (REdis Serialization Protocol), client and server-side. It
  contains basic wrappers for readers and writers to read/write requests and
  responses.
* [client](./client/) contains a minimalist pooled client.

For full documentation and examples, please see the individual packages and the
official API documentation: https://godoc.org/github.com/bsm/redeo.

## Examples

A simple server example with two commands:

```go
package main

import (
  "net"

  "github.com/bsm/redeo/v2"
)

func main() {
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
}
```

More complex handlers:

```go
func main() {
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
```

Redeo also supports command wrappers:

```go
func main() {
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
```
