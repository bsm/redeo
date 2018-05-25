package replication

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("backlog", func() {
	var subject *backlog

	BeforeEach(func() {
		subject = newBacklog(320640, 0)
	})

	It("should write", func() {
		Expect(subject.Write([]byte("testdata1234"))).To(Equal(12))
		Expect(subject.pos).To(Equal(12))
		Expect(subject.histlen).To(Equal(12))
		Expect(subject.offset).To(Equal(int64(320652)))
		Expect(subject.MinOffset()).To(Equal(int64(320640)))

		Expect(subject.Write([]byte("testdata1234"))).To(Equal(12))
		Expect(subject.pos).To(Equal(24))
		Expect(subject.histlen).To(Equal(24))
		Expect(subject.offset).To(Equal(int64(320664)))
		Expect(subject.MinOffset()).To(Equal(int64(320640)))

		for i := 0; i < 10920; i++ {
			Expect(subject.Write([]byte("testdata1234"))).To(Equal(12))
		}
		Expect(subject.pos).To(Equal(131064))
		Expect(subject.histlen).To(Equal(131064))
		Expect(subject.offset).To(Equal(int64(451704)))
		Expect(subject.MinOffset()).To(Equal(int64(320640)))

		Expect(subject.Write([]byte("testdata1234"))).To(Equal(12))
		Expect(subject.pos).To(Equal(4))
		Expect(subject.histlen).To(Equal(131072))
		Expect(subject.offset).To(Equal(int64(451716)))
		Expect(subject.MinOffset()).To(Equal(int64(320644)))

		Expect(subject.Write([]byte("testdata1234"))).To(Equal(12))
		Expect(subject.pos).To(Equal(16))
		Expect(subject.histlen).To(Equal(131072))
		Expect(subject.offset).To(Equal(int64(451728)))
		Expect(subject.MinOffset()).To(Equal(int64(320656)))
	})

})
