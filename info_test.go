package redeo

import (
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Info", func() {
	var subject *Info

	BeforeEach(func() {
		subject = NewInfo(&Config{
			Addr:   "127.0.0.1:9736",
			Socket: "/tmp/redeo.sock",
		})
		subject.Connected()
		subject.Called()
		subject.Called()
		subject.Connected()
		subject.Called()
		subject.Called()
		subject.Called()
		subject.Called()
		subject.Connected()
		subject.Called()
		subject.Called()
		subject.Disconnected()
		subject.Connected()
		subject.Called()
		subject.Called()
		subject.Connected()
		subject.Called()
		subject.Disconnected()
		subject.Disconnected()
	})

	It("should generate strings", func() {
		str := subject.String()
		Expect(str).To(ContainSubstring("# Server\n"))
		Expect(str).To(MatchRegexp(`process_id:\d+\n`))
		Expect(str).To(ContainSubstring("tcp_port:9736\n"))
		Expect(str).To(ContainSubstring("unix_socket:/tmp/redeo.sock\n"))
		Expect(str).To(MatchRegexp(`uptime_in_seconds:\d+\n`))
		Expect(str).To(MatchRegexp(`uptime_in_days:\d+\n`))

		Expect(str).To(ContainSubstring("# Clients\n"))
		Expect(str).To(ContainSubstring("connected_clients:2\n"))

		Expect(str).To(ContainSubstring("# Stats\n"))
		Expect(str).To(ContainSubstring("total_connections_received:5\n"))
		Expect(str).To(ContainSubstring("total_commands_processed:11\n"))
	})

	It("should calculate uptime", func() {
		Expect(subject.Uptime()).To(BeNumerically("~", time.Second, time.Second))
	})

})
