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
		subject.OnConnect(a)
		subject.OnCommand(a, "ping")
		subject.OnCommand(a, "info")
		subject.OnConnect(b)
		subject.OnCommand(b, "get")
		subject.OnCommand(a, "ping")
		subject.OnCommand(a, "get")
		subject.OnCommand(b, "set")
		subject.OnConnect(c)
		subject.OnCommand(c, "ping")
		subject.OnCommand(c, "info")
		subject.OnDisconnect(a)
		subject.OnConnect(d)
		subject.OnCommand(d, "ping")
		subject.OnCommand(d, "info")
		subject.OnConnect(a)
		subject.OnCommand(c, "quit")
		subject.OnDisconnect(c)
		subject.OnCommand(nil, "noop")
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
		Expect(str).To(ContainSubstring("total_commands_processed:12\n"))
	})

	It("should generate client string", func() {
		str := subject.ClientsString()
		Expect(str).To(MatchRegexp(`id=\d+ addr=1\.2\.3\.4\:10002 age=\d+ idle=\d+ cmd=set`))
		Expect(str).To(MatchRegexp(`id=\d+ addr=1\.2\.3\.4\:10004 age=\d+ idle=\d+ cmd=info`))
		Expect(str).To(MatchRegexp(`id=\d+ addr=1\.2\.3\.4\:10001 age=\d+ idle=\d+ cmd=get`))
	})

	It("should expose stats", func() {
		Expect(subject.UpTime()).To(BeNumerically(">", 0))
		Expect(subject.UpTime()).To(BeNumerically("<", time.Second))
		Expect(subject.ClientsLen()).To(Equal(3))
		Expect(subject.TotalConnections()).To(Equal(uint64(5)))
		Expect(subject.TotalProcessed()).To(Equal(uint64(12)))
	})

})
