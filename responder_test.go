package redeo

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"io"
	"strings"
	"testing"
)

var _ = Describe("Responder", func() {
	var subject *Responder

	BeforeEach(func() {
		subject = NewResponder()
	})

	It("should write inline strings", func() {
		n := subject.WriteInlineString("HELLO")
		Expect(n).To(Equal(8))
		Expect(subject.String()).To(Equal("+HELLO\r\n"))
	})

	It("should write strings", func() {
		n := subject.WriteString("HELLO")
		Expect(n).To(Equal(11))
		Expect(subject.String()).To(Equal("$5\r\nHELLO\r\n"))
	})

	It("should write ints", func() {
		Expect(subject.WriteInt(345)).To(Equal(6))
		Expect(subject.WriteZero()).To(Equal(4))
		Expect(subject.WriteOne()).To(Equal(4))
		Expect(subject.String()).To(Equal(":345\r\n:0\r\n:1\r\n"))
	})

	It("should write error strings", func() {
		n := subject.WriteErrorString("ERR some error")
		Expect(n).To(Equal(17))
		Expect(subject.String()).To(Equal("-ERR some error\r\n"))
	})

	It("should write errors", func() {
		n := subject.WriteError(ErrInvalidRequest)
		Expect(n).To(Equal(22))
		Expect(subject.String()).To(Equal("-ERR invalid request\r\n"))

		n = subject.WriteError(io.EOF)
		Expect(n).To(Equal(10))
		Expect(subject.String()[22:]).To(Equal("-ERR EOF\r\n"))
	})

	It("should write OK", func() {
		n := subject.WriteOK()
		Expect(n).To(Equal(5))
		Expect(subject.String()).To(Equal("+OK\r\n"))
	})

	It("should write nils", func() {
		n := subject.WriteNil()
		Expect(n).To(Equal(5))
		Expect(subject.String()).To(Equal("$-1\r\n"))
	})

	It("should write bulk lens", func() {
		n := subject.WriteBulkLen(4)
		Expect(n).To(Equal(4))
		Expect(subject.String()).To(Equal("*4\r\n"))
	})

})

func BenchmarkResponder_WriteOK(b *testing.B) {
	r := NewResponder()
	for i := 0; i < b.N; i++ {
		r.WriteOK()
	}
}

func BenchmarkResponder_WriteNil(b *testing.B) {
	r := NewResponder()
	for i := 0; i < b.N; i++ {
		r.WriteNil()
	}
}

func BenchmarkResponder_WriteInlineString(b *testing.B) {
	r := NewResponder()
	s := strings.Repeat("x", 64)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r.WriteInlineString(s)
	}
}

func BenchmarkResponder_WriteString(b *testing.B) {
	r := NewResponder()
	s := strings.Repeat("x", 64)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r.WriteString(s)
		r.Truncate(0)
	}
}

func BenchmarkResponder_WriteInt(b *testing.B) {
	r := NewResponder()
	for i := 0; i < b.N; i++ {
		r.WriteInt(98765)
	}
}
