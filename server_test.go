package redeo

import (
	"encoding/json"
	"fmt"
	"net"
	"strings"
	"testing"
	"time"

	"github.com/bsm/redeo/resp"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Server", func() {
	var subject *Server

	var (
		pong = func(w resp.ResponseWriter, _ *resp.Command) { w.AppendInlineString("PONG") }

		echo = func(w resp.ResponseWriter, cmd *resp.Command) {
			if cmd.ArgN() != 1 {
				w.AppendError(WrongNumberOfArgs(cmd.Name))
				return
			}
			w.AppendBulk(cmd.Arg(0))
		}

		flush = func(w resp.ResponseWriter, _ *resp.Command) {
			w.AppendOK()
			w.Flush()
		}

		stream = func(w resp.ResponseWriter, cmd *resp.CommandStream) {
			if cmd.ArgN() != 1 {
				w.AppendError(WrongNumberOfArgs(cmd.Name))
				return
			}

			rd, err := cmd.NextArg()
			if err != nil {
				w.AppendErrorf("ERR unable to parse argument: %s", err.Error())
				return
			}

			data := struct {
				N int
				S string
			}{}

			if err := json.NewDecoder(rd).Decode(&data); err != nil {
				w.AppendErrorf("ERR unable to decode argument: %s", err.Error())
				return
			}

			w.AppendInlineString(fmt.Sprintf("%s.%d", data.S, data.N))
			w.AppendOK()
		}
	)

	var runServer = func(srv *Server, fn func(net.Conn, *resp.RequestWriter, resp.ResponseReader)) {
		lis, err := net.Listen("tcp", "127.0.0.1:0")
		Expect(err).NotTo(HaveOccurred())
		defer lis.Close()

		// start listening
		go srv.Serve(lis)

		// connect client
		cn, err := net.Dial("tcp", lis.Addr().String())
		Expect(err).NotTo(HaveOccurred())
		defer cn.Close()

		fn(cn, resp.NewRequestWriter(cn), resp.NewResponseReader(cn))
	}

	BeforeEach(func() {
		subject = NewServer(&Config{
			Timeout: 100 * time.Millisecond,
		})
		subject.HandleFunc("pInG", pong)
		subject.HandleFunc("echo", echo)
		subject.HandleFunc("flush", flush)
		subject.HandleStreamFunc("stream", stream)
	})

	It("should register handlers", func() {
		Expect(subject.cmds).To(HaveLen(4))
		Expect(subject.cmds).To(HaveKey("ping"))
	})

	It("should serve", func() {
		runServer(subject, func(cn net.Conn, cw *resp.RequestWriter, cr resp.ResponseReader) {
			cw.WriteCmd("PING")
			Expect(cw.Flush()).To(Succeed())

			s, err := cr.ReadInlineString()
			Expect(err).NotTo(HaveOccurred())
			Expect(s).To(Equal("PONG"))

			info := subject.Info()
			Expect(info.NumClients()).To(Equal(1))
			Expect(info.TotalCommands()).To(Equal(int64(1)))
			Expect(info.TotalConnections()).To(Equal(int64(1)))
			Expect(info.ClientInfo()[0].LastCmd).To(Equal("ping"))

			cw.WriteCmdString("echo", strings.Repeat("x", 10000))
			Expect(cw.Flush()).To(Succeed())

			s, err = cr.ReadBulkString()
			Expect(err).NotTo(HaveOccurred())
			Expect(len(s)).To(Equal(10000))

			info = subject.Info()
			Expect(info.NumClients()).To(Equal(1))
			Expect(info.TotalCommands()).To(Equal(int64(2)))
			Expect(info.TotalConnections()).To(Equal(int64(1)))
			Expect(info.ClientInfo()[0].LastCmd).To(Equal("echo"))
		})
	})

	It("should serve streams", func() {
		runServer(subject, func(cn net.Conn, cw *resp.RequestWriter, cr resp.ResponseReader) {
			cw.WriteCmdString("STREAM", `{"n":8,"s":"hello"}`)
			Expect(cw.Flush()).To(Succeed())

			s, err := cr.ReadInlineString()
			Expect(err).NotTo(HaveOccurred())
			Expect(s).To(Equal("hello.8"))
		})
	})

	It("should handle pipelines", func() {
		runServer(subject, func(cn net.Conn, cw *resp.RequestWriter, cr resp.ResponseReader) {
			cw.WriteCmd("PING")
			cw.WriteCmd("PING")
			cw.WriteCmd("PING")
			Expect(cw.Flush()).To(Succeed())

			for i := 0; i < 3; i++ {
				s, err := cr.ReadInlineString()
				Expect(err).NotTo(HaveOccurred())
				Expect(s).To(Equal("PONG"))
			}
		})
	})

	It("should handle invalid commands", func() {
		runServer(subject, func(cn net.Conn, cw *resp.RequestWriter, cr resp.ResponseReader) {
			cw.WriteCmd("nOOp")
			Expect(cw.Flush()).To(Succeed())

			s, err := cr.ReadError()
			Expect(err).NotTo(HaveOccurred())
			Expect(s).To(Equal("ERR unknown command 'nOOp'"))

			// connection should still be open
			cw.WriteCmd("PING")
			Expect(cw.Flush()).To(Succeed())
			s, err = cr.ReadInlineString()
			Expect(err).NotTo(HaveOccurred())
			Expect(s).To(Equal("PONG"))
		})
	})

	It("should handle invalid commands in pipelines", func() {
		runServer(subject, func(cn net.Conn, cw *resp.RequestWriter, cr resp.ResponseReader) {
			cw.WriteCmd("PING")
			cw.WriteCmd("BAD")
			cw.WriteCmd("PING")
			Expect(cw.Flush()).To(Succeed())

			s, err := cr.ReadInlineString()
			Expect(err).NotTo(HaveOccurred())
			Expect(s).To(Equal("PONG"))

			s, err = cr.ReadError()
			Expect(err).NotTo(HaveOccurred())
			Expect(s).To(Equal("ERR unknown command 'BAD'"))

			s, err = cr.ReadInlineString()
			Expect(err).NotTo(HaveOccurred())
			Expect(s).To(Equal("PONG"))
		})
	})

	It("should handle client errors", func() {
		runServer(subject, func(cn net.Conn, cw *resp.RequestWriter, cr resp.ResponseReader) {
			cw.WriteCmd("ECHO")
			Expect(cw.Flush()).To(Succeed())

			s, err := cr.ReadError()
			Expect(err).NotTo(HaveOccurred())
			Expect(s).To(Equal("ERR wrong number of arguments for 'ECHO' command"))

			// connection should still be open
			cw.WriteCmd("PING")
			Expect(cw.Flush()).To(Succeed())
			s, err = cr.ReadInlineString()
			Expect(err).NotTo(HaveOccurred())
			Expect(s).To(Equal("PONG"))
		})
	})

	It("should handle client errors in pipelines", func() {
		runServer(subject, func(cn net.Conn, cw *resp.RequestWriter, cr resp.ResponseReader) {
			cw.WriteCmd("PING")
			cw.WriteCmd("echo")
			cw.WriteCmd("PING")
			Expect(cw.Flush()).To(Succeed())

			s, err := cr.ReadInlineString()
			Expect(err).NotTo(HaveOccurred())
			Expect(s).To(Equal("PONG"))

			s, err = cr.ReadError()
			Expect(err).NotTo(HaveOccurred())
			Expect(s).To(Equal("ERR wrong number of arguments for 'echo' command"))

			s, err = cr.ReadInlineString()
			Expect(err).NotTo(HaveOccurred())
			Expect(s).To(Equal("PONG"))
		})
	})

	It("should handle protocol errors", func() {
		runServer(subject, func(cn net.Conn, cw *resp.RequestWriter, cr resp.ResponseReader) {
			_, err := cn.Write([]byte("*x\r\n"))
			Expect(err).NotTo(HaveOccurred())

			x, _ := cr.PeekType()
			Expect(x).To(Equal(resp.TypeError))

			s, err := cr.ReadError()
			Expect(err).NotTo(HaveOccurred())
			Expect(s).To(Equal("ERR Protocol error: invalid multibulk length"))

			// connection should still be open
			cw.WriteCmd("PING")
			Expect(cw.Flush()).To(Succeed())
			s, err = cr.ReadInlineString()
			Expect(err).NotTo(HaveOccurred())
			Expect(s).To(Equal("PONG"))
		})
	})

	It("should handle protocol errors in pipelines", func() {
		runServer(subject, func(cn net.Conn, cw *resp.RequestWriter, cr resp.ResponseReader) {
			_, err := cn.Write([]byte("*1\r\n$4\r\nPING\r\n*1\r\n$x\r\nPING\r\n*1\r\n$4\r\nPING\r\n"))
			Expect(err).NotTo(HaveOccurred())

			s, err := cr.ReadInlineString()
			Expect(err).NotTo(HaveOccurred())
			Expect(s).To(Equal("PONG"))

			s, err = cr.ReadError()
			Expect(err).NotTo(HaveOccurred())
			Expect(s).To(Equal("ERR Protocol error: invalid bulk length"))

			s, err = cr.ReadInlineString()
			Expect(err).NotTo(HaveOccurred())
			Expect(s).To(Equal("PONG"))
		})
	})

	It("should close connections on EOF errors", func() {
		runServer(subject, func(cn net.Conn, cw *resp.RequestWriter, cr resp.ResponseReader) {
			_, err := cn.Write([]byte("*1\r\n$4\r\nPI"))
			Expect(err).NotTo(HaveOccurred())

			// connection should be closed
			_, err = cr.PeekType()
			Expect(err).To(MatchError("EOF"))
		})
	})

})

// --------------------------------------------------------------------

func BenchmarkServer_inline(b *testing.B) {
	benchmarkServer(b, []byte(
		"ECHO HELLO\r\n"+
			"ECHO CRUEL\r\n"+
			"ECHO WORLD\r\n",
	), 24)
}

func BenchmarkServer_bulk(b *testing.B) {
	benchmarkServer(b, []byte(
		"*2\r\n$4\r\nECHO\r\n$5\r\nHELLO\r\n"+
			"*2\r\n$4\r\nECHO\r\n$5\r\nCRUEL\r\n"+
			"*2\r\n$4\r\nECHO\r\n$5\r\nWORLD\r\n",
	), 24)
}

func benchmarkServer(b *testing.B, pipe []byte, expN int) {
	lis, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		b.Fatal(err)
	}
	defer lis.Close()

	srv := NewServer(nil)
	srv.HandleFunc("echo", func(w resp.ResponseWriter, cmd *resp.Command) {
		if cmd.ArgN() != 1 {
			w.AppendError(WrongNumberOfArgs(cmd.Name))
		}
		w.AppendInline(cmd.Arg(0))
	})

	go srv.Serve(lis)

	conn, err := net.Dial("tcp", lis.Addr().String())
	if err != nil {
		b.Fatal(err)
	}
	defer conn.Close()

	buf := make([]byte, 1024)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := conn.Write(pipe); err != nil {
			b.Fatal(err)
		}
		if n, err := conn.Read(buf); err != nil {
			b.Fatal(err)
		} else if n != expN {
			b.Fatalf("expected response to be %d bytes long, not %d", expN, n)
		}
	}
}
