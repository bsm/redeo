package resp_test

import (
	"bytes"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/bsm/redeo/resp"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
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

	DescribeTable("Append",
		func(v interface{}, exp string) {
			subject.Append(v)
			Expect(subject.Flush()).To(Succeed())
			Expect(strconv.Quote(buf.String())).To(Equal(strconv.Quote(exp)))
		},

		Entry("nil", nil, "$-1\r\n"),
		Entry("error", errors.New("failed"), "-ERR failed\r\n"),
		Entry("standard error", errors.New("ERR failed"), "-ERR failed\r\n"),
		Entry("int", 33, ":33\r\n"),
		Entry("int64", int64(33), ":33\r\n"),
		Entry("uint", uint(33), ":33\r\n"),
		Entry("bool (true)", true, ":1\r\n"),
		Entry("bool (false)", false, ":0\r\n"),
		Entry("float32", float32(0.1231), "+0.1231\r\n"),
		Entry("float64", 0.7357, "+0.7357\r\n"),
		Entry("negative float64", -0.4214, "+-0.4214\r\n"),
		Entry("string", "many words", "$10\r\nmany words\r\n"),
		Entry("[]byte", []byte("many words"), "$10\r\nmany words\r\n"),
		Entry("[]string", []string{"a", "b", "c"}, "*3\r\n$1\r\na\r\n$1\r\nb\r\n$1\r\nc\r\n"),
		Entry("[][]byte", [][]byte{{'a'}, {'b'}, {'c'}}, "*3\r\n$1\r\na\r\n$1\r\nb\r\n$1\r\nc\r\n"),
		Entry("[]int", []int{3, 5, 2}, "*3\r\n:3\r\n:5\r\n:2\r\n"),
		Entry("[]int64", []int64{7, 8, 3}, "*3\r\n:7\r\n:8\r\n:3\r\n"),
		Entry("[][]int64", [][]int64{{1, 2}, {3, 4}}, "*2\r\n*2\r\n:1\r\n:2\r\n*2\r\n:3\r\n:4\r\n"),
		Entry("map[string]string", map[string]string{"a": "b"}, "*2\r\n$1\r\na\r\n$1\r\nb\r\n"),
		Entry("map[int64]float64", map[int64]float64{1: 1.1}, "*2\r\n:1\r\n+1.1\r\n"),
		Entry("custom response", &customResponse{Host: "foo", Port: 8888}, "$17\r\ncustom 'foo:8888'\r\n"),
		Entry("custom error", customErrorResponse("bar"), "-WRONG bar\r\n"),
	)

	It("should reject bad custom types", func() {
		Expect(subject.Append(time.Time{})).To(MatchError(`resp: unsupported type time.Time`))
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

	Describe("Scan", func() {

		DescribeTable("success",
			func(s string, v interface{}, exp interface{}) {
				_, err := buf.WriteString(s)
				Expect(err).NotTo(HaveOccurred())
				Expect(subject.Scan(v)).To(Succeed())
				Expect(reflect.ValueOf(v).Elem().Interface()).To(Equal(exp))
			},

			Entry("bool (numeric true)", ":1\r\n", new(bool), true),
			Entry("bool (numeric false)", ":0\r\n", new(bool), false),
			Entry("bool (OK)", "+OK\r\n", new(bool), true),
			Entry("bool (true from inline)", "+1\r\n", new(bool), true),
			Entry("bool (true from bulk)", "$1\r\n1\r\n", new(bool), true),
			Entry("bool (false from inline)", "+0\r\n", new(bool), false),
			Entry("bool (false from bulk)", "$1\r\n0\r\n", new(bool), false),

			Entry("int64", ":123\r\n", new(int64), int64(123)),
			Entry("int32", ":123\r\n", new(int32), int32(123)),
			Entry("int16", ":123\r\n", new(int16), int16(123)),
			Entry("int8", ":123\r\n", new(int8), int8(123)),
			Entry("int", ":123\r\n", new(int), int(123)),
			Entry("int (from inline)", "+123\r\n", new(int), int(123)),
			Entry("int (from bulk)", "$3\r\n123\r\n", new(int), int(123)),

			Entry("uint64", ":123\r\n", new(uint64), uint64(123)),
			Entry("uint32", ":123\r\n", new(uint32), uint32(123)),
			Entry("uint16", ":123\r\n", new(uint16), uint16(123)),
			Entry("uint8", ":123\r\n", new(uint8), uint8(123)),
			Entry("uint", ":123\r\n", new(uint), uint(123)),
			Entry("uint (from inline)", "+123\r\n", new(uint), uint(123)),
			Entry("uint (from bulk)", "$3\r\n123\r\n", new(uint), uint(123)),

			Entry("float64 (from string)", "+2.312\r\n", new(float64), 2.312),
			Entry("float64 (from int)", ":123\r\n", new(float64), 123.0),
			Entry("float32 (from string)", "$5\r\n2.312\r\n", new(float32), float32(2.312)),
			Entry("float32 (from int)", ":123\r\n", new(float32), float32(123.0)),

			Entry("string (inline)", "+hello\r\n", new(string), "hello"),
			Entry("string (bulk)", "$5\r\nhello\r\n", new(string), "hello"),
			Entry("string (from int)", ":123\r\n", new(string), "123"),

			Entry("bytes (inline)", "+hello\r\n", new([]byte), []byte("hello")),
			Entry("bytes (bulk)", "$5\r\nhello\r\n", new([]byte), []byte("hello")),
			Entry("bytes (from int)", ":123\r\n", new([]byte), []byte("123")),
			Entry("bytes (from nil)", "$-1\r\n", new([]byte), ([]byte)(nil)),

			Entry("string slices", "*2\r\n+hello\r\n$5\r\nworld\r\n", new([]string), []string{"hello", "world"}),
			Entry("string slices (with ints)", "*2\r\n+hello\r\n:123\r\n", new([]string), []string{"hello", "123"}),
			Entry("number slices", "*2\r\n:1\r\n:2\r\n", new([]int64), []int64{1, 2}),
			Entry("number slices (from strings)", "*2\r\n:1\r\n+2\r\n", new([]int64), []int64{1, 2}),
			Entry("nested slices", "*2\r\n*2\r\n:1\r\n:2\r\n*2\r\n:3\r\n:4\r\n", new([][]int64), [][]int64{
				{1, 2},
				{3, 4},
			}),

			Entry("maps", "*2\r\n+hello\r\n$5\r\nworld\r\n", new(map[string]string), map[string]string{
				"hello": "world",
			}),
			Entry("maps (mixed)", "*4\r\n+foo\r\n+bar\r\n+baz\r\n:3\r\n", new(map[string]string), map[string]string{
				"foo": "bar",
				"baz": "3",
			}),
			Entry("maps (nested)", "*4\r\n+foo\r\n*2\r\n+bar\r\n:1\r\n+baz\r\n*2\r\n+boo\r\n:2\r\n", new(map[string]map[string]int), map[string]map[string]int{
				"foo": {"bar": 1},
				"baz": {"boo": 2},
			}),
			Entry("slice of maps", "*2\r\n*2\r\n+bar\r\n:1\r\n*2\r\n+boo\r\n:2\r\n", new([]map[string]int), []map[string]int{
				{"bar": 1},
				{"boo": 2},
			}),

			Entry("nullable (from nil)", "$-1\r\n", new(resp.NullString), resp.NullString{}),
			Entry("nullable (inline)", "+foo\r\n", new(resp.NullString), resp.NullString{Value: "foo", Valid: true}),

			Entry("scannable", "*2\r\n"+
				"*6\r\n+llen\r\n:2\r\n*2\r\n+readonly\r\n+fast\r\n:1\r\n:1\r\n:1\r\n"+
				"*6\r\n+mset\r\n:-3\r\n*1\r\n+write\r\n:1\r\n:-1\r\n:2\r\n",
				new([]ScannableStruct), []ScannableStruct{
					{Name: "llen", Arity: 2, Flags: []string{"readonly", "fast"}, FirstKey: 1, LastKey: 1, KeyStep: 1},
					{Name: "mset", Arity: -3, Flags: []string{"write"}, FirstKey: 1, LastKey: -1, KeyStep: 2},
				}),
		)

		DescribeTable("failure",
			func(s string, v interface{}, exp string) {
				_, err := buf.WriteString(s)
				Expect(err).NotTo(HaveOccurred())
				Expect(subject.Scan(v)).To(MatchError(exp))
			},
			Entry("errors", "-ERR something bad\r\n", new(string), `resp: server error "ERR something bad"`),
			Entry("bad type", "+hello\r\n", new(time.Time), `resp: error on Scan into *time.Time: unsupported conversion from "hello"`),
			Entry("not a pointer", "+hello\r\n", "value", `resp: error on Scan into string: destination not a pointer`),

			Entry("bool (bad type)", "*3\r\n", new(bool), `resp: error on Scan into *bool: unsupported conversion from array[3]`),
			Entry("bool (string)", "+hello\r\n", new(bool), `resp: error on Scan into *bool: unsupported conversion from "hello"`),
			Entry("bool (bad numeric)", ":2\r\n", new(bool), `resp: error on Scan into *bool: unsupported conversion from 2`),
			Entry("bool (from nil)", "$-1\r\n", new(bool), `resp: error on Scan into *bool: unsupported conversion from <nil>`),

			Entry("int64 (bad type)", "*3\r\n", new(int64), `resp: error on Scan into *int64: unsupported conversion from array[3]`),
			Entry("int64 (string)", "+hello\r\n", new(int64), `resp: error on Scan into *int64: unsupported conversion from "hello"`),
			Entry("int64 (from nil)", "$-1\r\n", new(int64), `resp: error on Scan into *int64: unsupported conversion from <nil>`),

			Entry("string (bad type)", "*3\r\n", new(string), `resp: error on Scan into *string: unsupported conversion from array[3]`),
			Entry("string (nil)", "$-1\r\n", new(string), `resp: error on Scan into *string: unsupported conversion from <nil>`),

			Entry("float64 (bad type)", "*3\r\n", new(float64), `resp: error on Scan into *float64: unsupported conversion from array[3]`),
			Entry("float64 (bad string)", "+hello\r\n", new(float64), `resp: error on Scan into *float64: unsupported conversion from "hello"`),

			Entry("slices (bad type)", "+hello\r\n", new([]string), `resp: error on Scan into *[]string: unsupported conversion from "hello"`),
			Entry("maps (odd number)", "*3\r\n+foo\r\n+bar\r\n+ba\r\n", new(map[string]string), `resp: error on Scan into *map[string]string: unsupported conversion from array[3]`),
		)

		DescribeTable("nil",
			func(s string, v interface{}) {
				_, err := buf.WriteString(s)
				Expect(err).NotTo(HaveOccurred())
				Expect(subject.Scan(v)).To(Succeed())
			},
			Entry("nil (from nil)", "$-1\r\n", nil),
			Entry("nil (from int)", ":123\r\n", nil),
			Entry("nil (from inline)", "+foo\r\n", nil),
			Entry("nil (from array)", "*1\r\n+foo\r\n", nil),
		)

	})

})

type customResponse struct {
	Host string
	Port int
}

func (r *customResponse) AppendTo(w resp.ResponseWriter) {
	w.AppendBulkString(fmt.Sprintf("custom '%s:%d'", r.Host, r.Port))
}

type customErrorResponse string

func (r customErrorResponse) AppendTo(w resp.ResponseWriter) {
	w.AppendError(r.Error())
}

func (r customErrorResponse) Error() string {
	return "WRONG " + string(r)
}
