package redeo

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Client", func() {
	var subject *Client

	BeforeEach(func() {
		subject = NewClient("1.2.3.4:10001")
	})

	It("should generate IDs", func() {
		a, b := NewClient("1.2.3.4:10001"), NewClient("1.2.3.4:10002")
		Expect(b.ID - 1).To(Equal(a.ID))
	})

	It("should generate info string", func() {
		subject.ID = 12
		Expect(subject.String()).To(Equal(`id=12 addr=1.2.3.4:10001 age=0 idle=0 cmd=`))
	})

})
