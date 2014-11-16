package redeo

import (
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Server", func() {
	var subject *Server
	var pong = func(out *Responder, _ *Request) error {
		out.WriteInlineString("PONG")
		return nil
	}
	var echo = func(out *Responder, req *Request) error {
		if len(req.Args) != 1 {
			return WrongNumberOfArgs(req.Name)
		}
		out.WriteString(req.Args[0])
		return nil
	}

	BeforeEach(func() {
		subject = NewServer(nil)
	})

	It("should fallback on default config", func() {
		Expect(subject.config).To(Equal(DefaultConfig))
	})

	It("should register handlers", func() {
		subject.HandleFunc("pInG", pong)
		Expect(subject.commands).To(HaveLen(1))
		Expect(subject.commands).To(HaveKey("ping"))
	})

	It("should apply requests", func() {
		subject.HandleFunc("echo", echo)
		client := NewClient("1.2.3.4:10001")
		res, err := subject.Apply(&Request{Name: "echo", client: client})
		Expect(err).To(Equal(WrongNumberOfArgs("echo")))
		Expect(client.lastCommand).To(Equal("echo"))

		res, err = subject.Apply(&Request{Name: "echo", Args: []string{"SAY HI!"}})
		Expect(err).NotTo(HaveOccurred())
		Expect(res.String()).To(Equal("$7\r\nSAY HI!\r\n"))

		res, err = subject.Apply(&Request{Name: "echo", Args: []string{strings.Repeat("x", 100000)}})
		Expect(err).NotTo(HaveOccurred())
		Expect(res.Len()).To(Equal(100011))
		Expect(res.String()[:9]).To(Equal("$100000\r\n"))
	})

})
