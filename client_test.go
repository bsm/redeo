package redeo

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Client", func() {
	var subject *Client

	BeforeEach(func() {
		subject = NewClient("1.2.3.4:10001")
		subject.ID = 12
	})

	It("should generate info string", func() {
		Expect(subject.String()).To(Equal(`id=12 addr=1.2.3.4:10001 age=0 cmd=`))
	})

})
