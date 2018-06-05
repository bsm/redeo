package replication

import (
	"net"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("masterLink", func() {
	var subject *masterLink
	var lis net.Listener

	BeforeEach(func() {
		subject = &masterLink{timeout: time.Hour}

		var err error
		lis, err = net.Listen("tcp", "127.0.0.1:0")
		Expect(err).NotTo(HaveOccurred())
	})

	AfterEach(func() {
		Expect(subject.Close()).To(Succeed())
		Expect(lis.Close()).To(Succeed())
	})

	It("should maintain conn", func() {
		Expect(subject.Addr()).To(Equal(""))
		Expect(subject.active).To(Equal(""))
		Expect(subject.MaintainConn()).To(Succeed())

		subject.SetAddr(lis.Addr().String())
		Expect(subject.Addr()).To(Equal(lis.Addr().String()))
		Expect(subject.active).To(Equal(""))
		Expect(subject.MaintainConn()).To(Succeed())
		Expect(subject.Addr()).To(Equal(lis.Addr().String()))
		Expect(subject.active).To(Equal(lis.Addr().String()))

		original := subject.conn
		subject.SetAddr("127.0.0.1:52151")
		Expect(subject.Addr()).To(Equal("127.0.0.1:52151"))
		Expect(subject.active).To(Equal(lis.Addr().String()))

		Expect(subject.MaintainConn()).To(HaveOccurred())
		Expect(subject.Addr()).To(Equal("127.0.0.1:52151"))
		Expect(subject.active).To(Equal(lis.Addr().String()))
		Expect(subject.conn).To(Equal(original))

		subject.SetAddr(lis.Addr().String())
		Expect(subject.Addr()).To(Equal(lis.Addr().String()))
		Expect(subject.active).To(Equal(lis.Addr().String()))
	})

})
