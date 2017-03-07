package redeo

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Client", func() {
	var subject *Client

	BeforeEach(func() {
		subject = newClient(&mockConn{Port: 10001})
	})

	It("should generate IDs", func() {
		a, b := newClient(&mockConn{}), newClient(&mockConn{})
		Expect(b.ID() - 1).To(Equal(a.ID()))
	})

})
