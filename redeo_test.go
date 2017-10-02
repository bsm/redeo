package redeo

import (
	"bytes"
	"net"
	"testing"
	"time"

	"github.com/bsm/redeo/redeotest"
	"github.com/bsm/redeo/resp"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Ping", func() {
	subject := Ping()

	It("should PONG", func() {
		w := redeotest.NewRecorder()
		subject.ServeRedeo(w, resp.NewCommand("PING"))
		Expect(w.Response()).To(Equal("PONG"))

		w = redeotest.NewRecorder()
		subject.ServeRedeo(w, resp.NewCommand("PING", resp.CommandArgument("eCHo")))
		Expect(w.Response()).To(Equal("eCHo"))

		w = redeotest.NewRecorder()
		subject.ServeRedeo(w, resp.NewCommand("PING", resp.CommandArgument("bad"), resp.CommandArgument("args")))
		Expect(w.Response()).To(Equal("ERR wrong number of arguments for 'PING' command"))
	})

})

// ------------------------------------------------------------------------

func TestSuite(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "redeo")
}

// --------------------------------------------------------------------

type mockConn struct {
	bytes.Buffer
	Port   int
	closed bool
}

func (m *mockConn) Close() error { m.closed = true; return nil }
func (m *mockConn) LocalAddr() net.Addr {
	return &net.TCPAddr{IP: net.IP{127, 0, 0, 1}, Port: 9736, Zone: ""}
}
func (m *mockConn) RemoteAddr() net.Addr {
	return &net.TCPAddr{IP: net.IP{1, 2, 3, 4}, Port: m.Port, Zone: ""}
}
func (m *mockConn) SetDeadline(_ time.Time) error      { return nil }
func (m *mockConn) SetReadDeadline(_ time.Time) error  { return nil }
func (m *mockConn) SetWriteDeadline(_ time.Time) error { return nil }
