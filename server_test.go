package redeo

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"strings"
)

var _ = Describe("Server", func() {
	var subject *Server
	var pong = func(out *Responder, _ *Request, _ interface{}) error {
		out.WriteInlineString("PONG")
		return nil
	}
	var echo = func(out *Responder, req *Request, _ interface{}) error {
		if len(req.Args) != 1 {
			return ErrWrongNumberOfArgs
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
		res, err := subject.Apply(&Request{Name: "echo"}, struct{}{})
		Expect(err).To(Equal(ErrWrongNumberOfArgs))

		res, err = subject.Apply(&Request{Name: "echo", Args: []string{"SAY HI!"}}, struct{}{})
		Expect(err).To(BeNil())
		Expect(res.String()).To(Equal("$7\r\nSAY HI!\r\n"))

		res, err = subject.Apply(&Request{Name: "echo", Args: []string{strings.Repeat("x", 100000)}}, struct{}{})
		Expect(err).To(BeNil())
		Expect(res.Len()).To(Equal(100011))
		Expect(res.String()[:9]).To(Equal("$100000\r\n"))
	})

})
