package info

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("StaticString", func() {
	var _ Value = StaticString("")

	It("should generate strings", func() {
		Expect(StaticString("x").String()).To(Equal("x"))
	})
})

var _ = Describe("StaticInt", func() {
	var _ Value = StaticInt(0)

	It("should generate strings", func() {
		Expect(StaticInt(12).String()).To(Equal("12"))
	})
})

var _ = Describe("Callback", func() {
	var _ Value = Callback(nil)

	It("should generate strings", func() {
		cb := Callback(func() string { return "x" })
		Expect(cb.String()).To(Equal("x"))
	})
})

var _ = Describe("IntValue", func() {
	var subject *IntValue
	var _ Value = subject

	BeforeEach(func() {
		subject = NewIntValue(0)
	})

	It("should have accessors", func() {
		Expect(subject.Inc(3)).To(Equal(int64(3)))
		Expect(subject.Inc(24)).To(Equal(int64(27)))
		Expect(subject.Value()).To(Equal(int64(27)))
		Expect(subject.Inc(-17)).To(Equal(int64(10)))
		Expect(subject.Value()).To(Equal(int64(10)))
		subject.Set(21)
		Expect(subject.Value()).To(Equal(int64(21)))
	})

	It("should generate strings", func() {
		Expect(subject.String()).To(Equal("0"))
	})
})

var _ = Describe("StringValue", func() {
	var subject *StringValue
	var _ Value = subject

	BeforeEach(func() {
		subject = NewStringValue("x")
	})

	It("should have accessors", func() {
		subject.Set("z")
		Expect(subject.String()).To(Equal("z"))
	})

	It("should generate strings", func() {
		Expect(subject.String()).To(Equal("x"))
	})
})
