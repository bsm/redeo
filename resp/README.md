# RESP

Low-level primitives for dealing with RESP (REdis Serialization Protocol), client and server-side.

## Server Examples

Reading requests:

```go
package main

import (
  "fmt"
  "strings"

  "github.com/bsm/redeo/resp"
)

func main() {
	cn := strings.NewReader("*1\r\n$4\r\nPING\r\n*2\r\n$4\r\nEcHO\r\n$5\r\nHeLLO\r\n")
	r := resp.NewRequestReader(cn)

	// read command
	cmd, _ := r.ReadCmd(nil)
	fmt.Println(cmd.Name)
	for i := 0; i < cmd.ArgN(); i++ {
		fmt.Println(i, cmd.Arg(i))
	}

	// read command, recycle previous instance
	cmd, _ = r.ReadCmd(cmd)
	fmt.Println(cmd.Name)
	for i := 0; i < cmd.ArgN(); i++ {
		fmt.Println(i, cmd.Arg(i))
	}

}
```

Writing responses:

```go
package main

import (
  "bytes"
  "fmt"

  "github.com/bsm/redeo/resp"
)

func main() {
	buf := new(bytes.Buffer)
	w := resp.NewResponseWriter(buf)

	// Append OK response
	w.AppendOK()

	// Append a number
	w.AppendInt(33)

	// Append an array
	w.AppendArrayLen(3)
	w.AppendBulkString("Adam")
	w.AppendBulkString("Had'em")
	w.AppendNil()

	// Writer data must be flushed manually
	fmt.Println(buf.Len(), w.Buffered())
	if err := w.Flush(); err != nil {
		panic(err)
	}

	// Once flushed, it will be sent to the underlying writer
	// as a bulk
	fmt.Println(buf.Len(), w.Buffered())
	fmt.Printf("%q\n", buf.String())

}
```

## Client Examples

Reading requests:

```go
package main

import (
  "fmt"
  "net"

  "github.com/bsm/redeo/resp"
)

func main() {
	cn, _ := net.Dial("tcp", "127.0.0.1:6379")
	defer cn.Close()

	// Wrap connection
	w := resp.NewRequestWriter(cn)
	r := resp.NewResponseReader(cn)

	// Write pipeline
	w.WriteCmdString("PING")
	w.WriteCmdString("ECHO", "HEllO")
	w.WriteCmdString("GET", "key")
	w.WriteCmdString("SET", "key", "value")
	w.WriteCmdString("DEL", "key")

	// Flush pipeline
	if err := w.Flush(); err != nil {
		panic(err)
	}

	// Consume responses
	for i := 0; i < 5; i++ {
		t, err := r.PeekType()
		if err != nil {
			return
		}

		switch t {
		case resp.TypeInline:
			s, _ := r.ReadInlineString()
			fmt.Println(s)
		case resp.TypeBulk:
			s, _ := r.ReadBulkString()
			fmt.Println(s)
		case resp.TypeInt:
			n, _ := r.ReadInt()
			fmt.Println(n)
		case resp.TypeNil:
			_ = r.ReadNil()
			fmt.Println(nil)
		default:
			panic("unexpected response type")
		}
	}

}
```
