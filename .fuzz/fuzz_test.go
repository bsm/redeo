package redeofuzz_test

import (
	"math/rand"
	"net"
	"sync"
	"testing"

	"github.com/bsm/redeo"
	"github.com/bsm/redeo/resp"
	"github.com/go-redis/redis"
)

func TestFuzz(t *testing.T) {
	lis, err := net.Listen("tcp", "127.0.0.1:")
	if err != nil {
		t.Fatalf("could not open a listener: %v", err)
		return
	}
	defer lis.Close()

	srv := initServer()
	go srv.Serve(lis)

	cln := redis.NewClient(&redis.Options{
		Addr:     lis.Addr().String(),
		PoolSize: 20,
	})
	defer cln.Close()

	if err := cln.Ping().Err(); err != nil {
		t.Fatalf("could not ping server: %v", err)
		return
	}

	n := 10000
	if testing.Short() {
		n = 1000
	}

	var wg sync.WaitGroup
	for k := 0; k < 10; k++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			for i := 0; i < n; i++ {
				if !fuzzIteration(t, cln, i) {
					return
				}
			}
		}()
	}
	wg.Wait()
}

func fuzzIteration(t *testing.T, c *redis.Client, i int) bool {
	if cmd, act, err, xargs, xbytes := fuzzCallMB(c); err != nil {
		t.Fatalf("fuzzmb failed with: %v, command: %q", err, cmd.String())
		return false
	} else if act["na"] != xargs {
		t.Fatalf("fuzzmb expected the number of processed arguments to be %d but was %d, command: %q", xargs, act["na"], cmd.String())
		return false
	} else if act["nb"] != xbytes {
		t.Fatalf("fuzzmb expected the number ofprocessed  bytes to be %d but was %d, command: %q", xbytes, act["nb"], cmd.String())
		return false
	}

	if i%3 == 0 {
		if err := fuzzCallErr(c); err == nil {
			t.Fatal("fuzzerr expected error but received none")
			return false
		} else if err.Error() != "ERR wrong number of arguments for 'fuzzerr' command" {
			t.Fatalf("fuzzerr returned unexpected error %v", err)
			return false
		}
	}

	if i%4 == 0 {
		if err := fuzzCallUnknown(c); err == nil {
			t.Fatal("fuzzunknown expected error but received none")
			return false
		} else if err.Error() != "ERR unknown command 'fuzzunknown'" {
			t.Fatalf("fuzzunknown returned unexpected error %v", err)
			return false
		}
	}

	if cmd, act, err, exp := fuzzCallStream(c); err != nil {
		t.Fatalf("fuzzstream failed with: %v, command: %q", err, cmd.String())
		return false
	} else if act != exp {
		t.Fatalf("fuzzstream expected the number of processed arguments to be %d but was %d, command: %q", exp, act, cmd.String())
		return false
	}

	return true
}

// --------------------------------------------------------------------

func fuzzCallMB(c *redis.Client) (cmd *redis.StringIntMapCmd, act map[string]int64, err error, xargs int64, xbytes int64) {
	xargs = rand.Int63n(20)
	args := append(make([]interface{}, 0, int(xargs+1)), "fuzzmb")
	for i := int64(0); i < xargs; i++ {
		b := make([]byte, rand.Intn(1024))
		n, _ := rand.Read(b)
		args = append(args, b[:n])
		xbytes += int64(n)
	}
	cmd = redis.NewStringIntMapCmd(args...)
	c.Process(cmd)
	act, err = cmd.Result()
	return
}

func fuzzCallErr(c *redis.Client) error {
	cmd := redis.NewStatusCmd("fuzzerr")
	c.Process(cmd)
	return cmd.Err()
}

func fuzzCallStream(c *redis.Client) (cmd *redis.IntCmd, act int64, err error, exp int64) {
	exp = rand.Int63n(3)
	args := append(make([]interface{}, 0, int(exp+1)), "fuzzstream")
	for i := int64(0); i < exp; i++ {
		b := make([]byte, rand.Intn(32*1024))
		n, _ := rand.Read(b)
		args = append(args, b[:n])
	}
	cmd = redis.NewIntCmd(args...)
	c.Process(cmd)
	act, err = cmd.Result()
	return
}

func fuzzCallUnknown(c *redis.Client) error {
	cmd := redis.NewStatusCmd("fuzzunknown")
	c.Process(cmd)
	return cmd.Err()
}

// --------------------------------------------------------------------

func initServer() *redeo.Server {
	s := redeo.NewServer(nil)
	s.Handle("ping", redeo.Ping())

	s.HandleFunc("fuzzmb", func(w resp.ResponseWriter, c *resp.Command) {
		sz := 0
		for _, a := range c.Args() {
			sz += len(a)
		}

		w.AppendArrayLen(4)
		w.AppendBulkString("na")
		w.AppendInt(int64(c.ArgN()))
		w.AppendBulkString("nb")
		w.AppendInt(int64(sz))
	})

	s.HandleFunc("fuzzerr", func(w resp.ResponseWriter, c *resp.Command) {
		w.AppendError(redeo.WrongNumberOfArgs(c.Name))
	})

	s.HandleStreamFunc("fuzzstream", func(w resp.ResponseWriter, c *resp.CommandStream) {
		if c.ArgN() != 0 {
			for i := 0; i < rand.Intn(c.ArgN()); i++ {
				rd, err := c.NextArg()
				if err != nil {
					w.AppendErrorf("ERR %v", err)
					return
				}
				if _, err := rd.Read(make([]byte, 16*1024)); err != nil {
					w.AppendErrorf("ERR %v", err)
					return
				}
			}
		}
		w.AppendInt(int64(c.ArgN()))
	})

	return s
}
