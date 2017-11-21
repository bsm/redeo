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

var _ = Describe("Commands", func() {
	subject := Commands([]CommandDetails{
		{Name: "GeT", Arity: 2, Flags: []string{"readonly", "fast"}, FirstKey: 1, LastKey: 1, KeyStepCount: 1},
		{Name: "randomkey", Arity: 1, Flags: []string{"readonly", "random"}},
		{Name: "mset", Arity: -3, Flags: []string{"write", "denyoom"}, FirstKey: 1, LastKey: -1, KeyStepCount: 2},
		{Name: "quit", Arity: 1},
	})

	It("should enumerate", func() {
		w := redeotest.NewRecorder()
		subject.ServeRedeo(w, resp.NewCommand("COMMAND"))
		Expect(w.Response()).To(Equal([]interface{}{
			[]interface{}{"get", int64(2), []interface{}{"readonly", "fast"}, int64(1), int64(1), int64(1)},
			[]interface{}{"randomkey", int64(1), []interface{}{"readonly", "random"}, int64(0), int64(0), int64(0)},
			[]interface{}{"mset", int64(-3), []interface{}{"write", "denyoom"}, int64(1), int64(-1), int64(2)},
			[]interface{}{"quit", int64(1), []interface{}{}, int64(0), int64(0), int64(0)},
		}))
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
