package redeo

import (
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("ServerInfo", func() {
	var subject *ServerInfo

	BeforeEach(func() {
		subject = NewServerInfo(&Config{
			Addr:   "127.0.0.1:9736",
			Socket: "/tmp/redeo.sock",
		})

		a, b, c, d := NewClient("1.2.3.4:10001"), NewClient("1.2.3.4:10002"), NewClient("1.2.3.4:10003"), NewClient("1.2.3.4:10004")
		subject.Connected(a)
		subject.Called(a, "ping")
		subject.Called(a, "info")
		subject.Connected(b)
		subject.Called(b, "get")
		subject.Called(a, "ping")
		subject.Called(a, "get")
		subject.Called(b, "set")
		subject.Connected(c)
		subject.Called(c, "ping")
		subject.Called(c, "info")
		subject.Disconnected(a)
		subject.Connected(d)
		subject.Called(d, "ping")
		subject.Called(d, "info")
		subject.Connected(a)
		subject.Called(c, "quit")
		subject.Disconnected(c)
	})

	It("should generate info string", func() {
		str := subject.String()
		Expect(str).To(ContainSubstring("# Server\n"))
		Expect(str).To(MatchRegexp(`process_id:\d+\n`))
		Expect(str).To(ContainSubstring("tcp_port:9736\n"))
		Expect(str).To(ContainSubstring("unix_socket:/tmp/redeo.sock\n"))
		Expect(str).To(MatchRegexp(`uptime_in_seconds:\d+\n`))
		Expect(str).To(MatchRegexp(`uptime_in_days:\d+\n`))

		Expect(str).To(ContainSubstring("# Clients\n"))
		Expect(str).To(ContainSubstring("connected_clients:3\n"))

		Expect(str).To(ContainSubstring("# Stats\n"))
		Expect(str).To(ContainSubstring("total_connections_received:5\n"))
		Expect(str).To(ContainSubstring("total_commands_processed:11\n"))
	})

	It("should generate client string", func() {
		str := subject.ClientsString()
		Expect(str).To(MatchRegexp(`id=2 addr=1\.2\.3\.4\:10002 age=\d+ cmd=set`))
		Expect(str).To(MatchRegexp(`id=3 addr=1\.2\.3\.4\:10004 age=\d+ cmd=info`))
		Expect(str).To(MatchRegexp(`id=4 addr=1\.2\.3\.4\:10001 age=\d+ cmd=get`))
	})

	It("should expose stats", func() {
		Expect(subject.Uptime()).To(BeNumerically("~", time.Second, time.Second))
		Expect(subject.ClientLen()).To(Equal(3))
		Expect(subject.Connections()).To(Equal(int64(5)))
		Expect(subject.Processed()).To(Equal(int64(11)))
	})

})
