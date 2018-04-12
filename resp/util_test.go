package resp

import (
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
)

var _ = DescribeTable("appendArgument",
	func(src string, exp string, expN int) {
		dst, n := appendArgument(nil, []byte(src))
		Expect(string(dst)).To(Equal(exp))
		Expect(n).To(Equal(expN))
	},

	Entry("empty",
		"", "", 0),
	Entry("blank",
		" \t ", "", 3),
	Entry("words",
		"  hello world", "hello", 7),
	Entry("words with tabs",
		"hello\tworld", "hello", 5),
	Entry("words with nl",
		"hello\nworld", "hello", 5),
	Entry("words interrupted by quotes",
		`he"llo" world`, "he", 2),
	Entry("words interrupted by single quotes",
		`he'llo' world`, "he", 2),
	Entry("quoted",
		` "hello my" world`, "hello my", 11),
	Entry("quoted with quotes",
		`"hello \"my\" " world`, `hello "my" `, 15),
	Entry("quoted with escaped hex chars",
		`"hello \x6dy" world`, `hello my`, 13),
	Entry("single quoted",
		` 'hello my' world`, "hello my", 11),
)
