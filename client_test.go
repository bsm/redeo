package redeo

import (
	. "github.com/bsm/ginkgo/v2"
	. "github.com/bsm/gomega"
)

var _ = Describe("Client", func() {
	It("should generate IDs", func() {
		a, b := newClient(&mockConn{}), newClient(&mockConn{})
		Expect(b.ID() - 1).To(Equal(a.ID()))
	})
})
