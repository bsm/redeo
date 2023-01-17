package info

import (
	"testing"

	. "github.com/bsm/ginkgo/v2"
	. "github.com/bsm/gomega"
)

var _ = Describe("Registry", func() {
	var subject *Registry

	BeforeEach(func() {
		subject = New()
		subject.FetchSection("Server").Register("version", StaticString("1.0.1"))
		subject.FetchSection("Server").Register("date", StaticString("2014-11-11"))
		subject.FetchSection("Clients").Register("count", StaticString("17"))
		subject.FetchSection("Clients").Register("total", StaticString("123456"))
		subject.FetchSection("Empty")
	})

	It("should generate info strings", func() {
		Expect(New().String()).To(Equal(""))
		Expect(subject.String()).To(Equal("# Server\nversion:1.0.1\ndate:2014-11-11\n\n# Clients\ncount:17\ntotal:123456\n"))
	})

	It("should clear", func() {
		subject.FetchSection("Clients").Clear()
		Expect(subject.sections[1].kvs).To(BeEmpty())
		subject.Clear()
		Expect(subject.sections).To(BeEmpty())
	})

	It("should replace", func() {
		subject.FetchSection("Server").Replace(func(s *Section) {
			s.Register("test", StaticString("string"))
		})
		s := subject.FindSection("server").String()
		Expect(s).To(Equal("# Server\ntest:string\n"))
	})

	It("should generate section strings", func() {
		s := subject.FindSection("clients").String()
		Expect(s).To(Equal("# Clients\ncount:17\ntotal:123456\n"))

		s = subject.FindSection("unknown").String()
		Expect(s).To(Equal(""))
	})

})

func TestSuite(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "redeo/info")
}
