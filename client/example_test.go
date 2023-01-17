package client_test

import (
	"fmt"

	"github.com/bsm/pool"
	"github.com/bsm/redeo/v2/client"
	"github.com/bsm/redeo/v2/resp"
)

func ExamplePool() {
	pool, _ := client.New(&pool.Options{
		InitialSize: 1,
	}, nil)
	defer pool.Close()

	cn, _ := pool.Get()
	defer pool.Put(cn)

	// Build pipeline
	cn.WriteCmdString("PING")
	cn.WriteCmdString("ECHO", "HEllO")
	cn.WriteCmdString("GET", "key")
	cn.WriteCmdString("SET", "key", "value")
	cn.WriteCmdString("DEL", "key")

	// Flush pipeline to socket
	if err := cn.Flush(); err != nil {
		cn.MarkFailed()
		panic(err)
	}

	// Consume responses
	for i := 0; i < 5; i++ {
		t, err := cn.PeekType()
		if err != nil {
			return
		}

		switch t {
		case resp.TypeInline:
			s, _ := cn.ReadInlineString()
			fmt.Println(s)
		case resp.TypeBulk:
			s, _ := cn.ReadBulkString()
			fmt.Println(s)
		case resp.TypeInt:
			n, _ := cn.ReadInt()
			fmt.Println(n)
		case resp.TypeNil:
			_ = cn.ReadNil()
			fmt.Println(nil)
		default:
			panic("unexpected response type")
		}
	}

	// Output:
	// PONG
	// HEllO
	// <nil>
	// OK
	// 1
}
