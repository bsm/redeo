package redeo

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("ServerInfo", func() {
	var subject *ServerInfo

	BeforeEach(func() {
		subject = newServerInfo(&Config{
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
		Expect(str).To(ContainSubstring("tcp_port:9736\nunix_socket:/tmp/redeo.sock\n"))
		Expect(str).To(MatchRegexp(`uptime_in_seconds:\d+\n`))
		Expect(str).To(MatchRegexp(`uptime_in_days:\d+\n`))

		Expect(str).To(ContainSubstring("# Clients\nconnected_clients:3\n"))
		Expect(str).To(ContainSubstring("# Stats\ntotal_connections_received:5\ntotal_commands_processed:12\n"))
	})

	It("should generate client string", func() {
		str := subject.ClientsString()
		Expect(str).To(MatchRegexp(`id=\d+ addr=1\.2\.3\.4\:10002 age=\d+ idle=\d+ cmd=set`))
		Expect(str).To(MatchRegexp(`id=\d+ addr=1\.2\.3\.4\:10004 age=\d+ idle=\d+ cmd=info`))
		Expect(str).To(MatchRegexp(`id=\d+ addr=1\.2\.3\.4\:10001 age=\d+ idle=\d+ cmd=get`))
	})

})
