package resp_test

import (
	"bytes"
	"strings"

	"github.com/bsm/redeo/resp"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("ResponseWriter", func() {
	var subject resp.ResponseWriter
	var buf = new(bytes.Buffer)

	BeforeEach(func() {
		buf.Reset()
		subject = resp.NewResponseWriter(buf)
	})

	It("should append bulks", func() {
		subject.AppendBulk([]byte("dAtA"))
		Expect(buf.String()).To(BeEmpty())
		Expect(subject.Flush()).To(Succeed())
		Expect(buf.String()).To(Equal("$4\r\ndAtA\r\n"))
	})

	It("should append bulk strings", func() {
		subject.AppendBulkString("PONG")
		Expect(buf.String()).To(BeEmpty())
		Expect(subject.Flush()).To(Succeed())
		Expect(buf.String()).To(Equal("$4\r\nPONG\r\n"))

		subject.AppendBulkString("日本")
		Expect(subject.Flush()).To(Succeed())
		Expect(buf.String()).To(Equal("$4\r\nPONG\r\n$6\r\n日本\r\n"))
	})

	It("should append inline bytes", func() {
		subject.AppendInline([]byte("dAtA"))
		Expect(buf.String()).To(BeEmpty())
		Expect(subject.Flush()).To(Succeed())
		Expect(buf.String()).To(Equal("+dAtA\r\n"))
	})

	It("should append inline strings", func() {
		subject.AppendInlineString("PONG")
		Expect(buf.String()).To(BeEmpty())
		Expect(subject.Flush()).To(Succeed())
		Expect(buf.String()).To(Equal("+PONG\r\n"))
	})

	It("should append errors", func() {
		subject.AppendError("WRONGTYPE not a number")
		Expect(buf.String()).To(BeEmpty())
		Expect(subject.Flush()).To(Succeed())
		Expect(buf.String()).To(Equal("-WRONGTYPE not a number\r\n"))
	})

	It("should append ints", func() {
		subject.AppendInt(27)
		Expect(buf.String()).To(BeEmpty())
		Expect(subject.Flush()).To(Succeed())
		Expect(buf.String()).To(Equal(":27\r\n"))

		subject.AppendInt(1)
		Expect(subject.Flush()).To(Succeed())
		Expect(buf.String()).To(Equal(":27\r\n:1\r\n"))
	})

	It("should append nils", func() {
		subject.AppendNil()
		Expect(buf.String()).To(BeEmpty())
		Expect(subject.Flush()).To(Succeed())
		Expect(buf.String()).To(Equal("$-1\r\n"))
	})

	It("should append OK", func() {
		subject.AppendOK()
		Expect(buf.String()).To(BeEmpty())
		Expect(subject.Flush()).To(Succeed())
		Expect(buf.String()).To(Equal("+OK\r\n"))
	})

	It("should copy from readers", func() {
		src := strings.NewReader("this is a streaming data source")
		subject.AppendArrayLen(1)
		Expect(buf.String()).To(BeEmpty())
		Expect(subject.CopyBulk(src, 16)).To(Succeed())
		Expect(subject.Flush()).To(Succeed())
		Expect(buf.String()).To(Equal("*1\r\n$16\r\nthis is a stream\r\n"))
	})

})

var _ = Describe("ResponseReader", func() {
	var subject resp.ResponseReader
	var buf = new(bytes.Buffer)

	BeforeEach(func() {
		buf.Reset()
		subject = resp.NewResponseReader(buf)
	})

	It("should read nils", func() {
		buf.WriteString("$-1\r\n+OK\r\n")

		t, err := subject.PeekType()
		Expect(err).NotTo(HaveOccurred())
		Expect(t).To(Equal(resp.TypeNil))

		err = subject.ReadNil()
		Expect(err).NotTo(HaveOccurred())

		// ensure we have consumed everything
		t, err = subject.PeekType()
		Expect(err).NotTo(HaveOccurred())
		Expect(t).To(Equal(resp.TypeInline))
	})

	It("should read strings", func() {
		buf.WriteString("$4\r\nPING\r\n+OK\r\n")

		t, err := subject.PeekType()
		Expect(err).NotTo(HaveOccurred())
		Expect(t).To(Equal(resp.TypeBulk))

		s, err := subject.ReadBulkString()
		Expect(err).NotTo(HaveOccurred())
		Expect(s).To(Equal("PING"))

		// ensure we have consumed everything
		t, err = subject.PeekType()
		Expect(err).NotTo(HaveOccurred())
		Expect(t).To(Equal(resp.TypeInline))
	})

	It("should read bytes", func() {
		buf.WriteString("$4\r\nPiNG\r\n+OK\r\n")

		t, err := subject.PeekType()
		Expect(err).NotTo(HaveOccurred())
		Expect(t).To(Equal(resp.TypeBulk))

		s, err := subject.ReadBulk(nil)
		Expect(err).NotTo(HaveOccurred())
		Expect(s).To(Equal([]byte("PiNG")))

		// ensure we have consumed everything
		t, err = subject.PeekType()
		Expect(err).NotTo(HaveOccurred())
		Expect(t).To(Equal(resp.TypeInline))
	})

	It("should read ints", func() {
		buf.WriteString(":21412\r\n+OK\r\n")

		t, err := subject.PeekType()
		Expect(err).NotTo(HaveOccurred())
		Expect(t).To(Equal(resp.TypeInt))

		n, err := subject.ReadInt()
		Expect(err).NotTo(HaveOccurred())
		Expect(n).To(Equal(int64(21412)))

		// ensure we have consumed everything
		t, err = subject.PeekType()
		Expect(err).NotTo(HaveOccurred())
		Expect(t).To(Equal(resp.TypeInline))
	})

	It("should read negative ints", func() {
		buf.WriteString(":-321\r\n+OK\r\n")

		t, err := subject.PeekType()
		Expect(err).NotTo(HaveOccurred())
		Expect(t).To(Equal(resp.TypeInt))

		n, err := subject.ReadInt()
		Expect(err).NotTo(HaveOccurred())
		Expect(n).To(Equal(int64(-321)))

		// ensure we have consumed everything
		t, err = subject.PeekType()
		Expect(err).NotTo(HaveOccurred())
		Expect(t).To(Equal(resp.TypeInline))
	})

	It("should read arrays", func() {
		buf.WriteString("*2\r\n$5\r\nHeLLo\r\n$5\r\nwOrld\r\n+OK\r\n")

		t, err := subject.PeekType()
		Expect(err).NotTo(HaveOccurred())
		Expect(t).To(Equal(resp.TypeArray))

		n, err := subject.ReadArrayLen()
		Expect(err).NotTo(HaveOccurred())
		Expect(n).To(Equal(2))

		for i := 0; i < n; i++ {
			t, err = subject.PeekType()
			Expect(err).NotTo(HaveOccurred())
			Expect(t).To(Equal(resp.TypeBulk))

			_, err = subject.ReadBulk(nil)
			Expect(err).NotTo(HaveOccurred())
		}

		// ensure we have consumed everything
		t, err = subject.PeekType()
		Expect(err).NotTo(HaveOccurred())
		Expect(t).To(Equal(resp.TypeInline))
	})

	It("should read errors", func() {
		buf.WriteString("-WRONGTYPE expected hash\r\n+OK\r\n")

		t, err := subject.PeekType()
		Expect(err).NotTo(HaveOccurred())
		Expect(t).To(Equal(resp.TypeError))

		s, err := subject.ReadError()
		Expect(err).NotTo(HaveOccurred())
		Expect(s).To(Equal("WRONGTYPE expected hash"))

		// ensure we have consumed everything
		t, err = subject.PeekType()
		Expect(err).NotTo(HaveOccurred())
		Expect(t).To(Equal(resp.TypeInline))
	})

	It("should read statuses", func() {
		buf.WriteString("+OK\r\n+OK\r\n")

		t, err := subject.PeekType()
		Expect(err).NotTo(HaveOccurred())
		Expect(t).To(Equal(resp.TypeInline))

		s, err := subject.ReadInlineString()
		Expect(err).NotTo(HaveOccurred())
		Expect(s).To(Equal("OK"))

		// ensure we have consumed everything
		t, err = subject.PeekType()
		Expect(err).NotTo(HaveOccurred())
		Expect(t).To(Equal(resp.TypeInline))
	})

	It("should read statuses across buffer overflows", func() {
		s := strings.Repeat("x", 4000)
		buf.WriteString("+")
		buf.WriteString(s)
		buf.WriteString("\r\n")
		buf.WriteString("+")
		buf.WriteString(s)
		buf.WriteString("\r\n")

		t, err := subject.PeekType()
		Expect(err).NotTo(HaveOccurred())
		Expect(t).To(Equal(resp.TypeInline))

		s, err = subject.ReadInlineString()
		Expect(err).NotTo(HaveOccurred())
		Expect(len(s)).To(Equal(4000))

		s, err = subject.ReadInlineString()
		Expect(err).NotTo(HaveOccurred())
		Expect(len(s)).To(Equal(4000))

		_, err = subject.PeekType()
		Expect(err).To(MatchError("EOF"))
	})

})
