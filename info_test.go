package redeo

import (
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("ServerInfo", func() {
	var subject *ServerInfo

	BeforeEach(func() {
		c1 := newClient(&mockConn{Port: 10001})

		subject = newServerInfo()
		subject.connections.Inc(5)
		subject.commands.Inc(12)
		subject.clients.Add(c1)
		subject.clients.Add(newClient(&mockConn{Port: 10002}))
		subject.clients.Add(newClient(&mockConn{Port: 10004}))
		subject.clients.Cmd(c1.ID(), "get")
	})

	It("should generate info string", func() {
		str := subject.String()
		Expect(str).To(ContainSubstring("# Server\n"))
		Expect(str).To(MatchRegexp(`process_id:\d+\n`))
		Expect(str).To(MatchRegexp(`uptime_in_seconds:\d+\n`))
		Expect(str).To(MatchRegexp(`uptime_in_days:\d+\n`))

		Expect(str).To(ContainSubstring("# Clients\nconnected_clients:3\n"))
		Expect(str).To(ContainSubstring("# Stats\ntotal_connections_received:5\ntotal_commands_processed:12\n"))
	})

	It("should retrieve a list of clients", func() {
		stats := subject.ClientInfo()
		Expect(stats).To(HaveLen(3))
		Expect(stats[0].String()).To(MatchRegexp(`id=\d+ addr=1\.2\.3\.4\:10001 age=\d+ idle=\d+ cmd=get`))
	})

})

var _ = Describe("ClientInfo", func() {

	It("should init", func() {
		c := newClient(&mockConn{Port: 10001})
		c.id = 12

		info := newClientInfo(c, time.Now().Add(-3*time.Second))
		Expect(info.String()).To(Equal(`id=12 addr=1.2.3.4:10001 age=3 idle=3 cmd=`))
	})

})
