package redeo

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"net"
)

type mockConn struct {
	addr string
}

func NewMockConn(addr string) *mockConn {
	return &mockConn{addr: addr}
}

func (c *mockConn) Close() error {
	return nil
}

func (c *mockConn) Network() string {
	return ""
}

func (c *mockConn) String() string {
	return c.addr
}

func (c *mockConn) RemoteAddr() net.Addr {
	return c
}

var _ = Describe("Client", func() {
	var subject *Client

	BeforeEach(func() {
		subject = NewClient(NewMockConn("1.2.3.4:10001"))
	})

	It("should generate IDs", func() {
		a, b := NewClient(NewMockConn("1.2.3.4:10001")), NewClient(NewMockConn("1.2.3.4:10002"))
		Expect(b.ID - 1).To(Equal(a.ID))
	})

	It("should generate info string", func() {
		subject.ID = 12
		Expect(subject.String()).To(Equal(`id=12 addr=1.2.3.4:10001 age=0 idle=0 cmd=`))
	})

})
