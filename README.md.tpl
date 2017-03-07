# Redeo

[![GoDoc](https://godoc.org/github.com/bsm/redeo?status.svg)](https://godoc.org/github.com/bsm/redeo)
[![Build Status](https://travis-ci.org/bsm/redeo.png?branch=master)](https://travis-ci.org/bsm/redeo)
[![Go Report Card](https://goreportcard.com/badge/github.com/bsm/redeo)](https://goreportcard.com/report/github.com/bsm/redeo)
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

  "github.com/bsm/redeo"
)

func main() {{ "ExampleServer" | code }}
```

More complex handlers:

```go
func main() {{ "ExampleHandlerFunc" | code }}
```
