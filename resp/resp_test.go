package resp_test

import (
	"bytes"
	"fmt"
	"reflect"
	"testing"

	. "github.com/bsm/ginkgo/v2"
	. "github.com/bsm/gomega"
	"github.com/bsm/gomega/types"
	"github.com/bsm/redeo/v2/resp"
)

var _ = Describe("ResponseType", func() {

	It("should implement stringer", func() {
		Expect(resp.TypeArray.String()).To(Equal("Array"))
		Expect(resp.TypeNil.String()).To(Equal("Nil"))
		Expect(resp.TypeUnknown.String()).To(Equal("Unknown"))
	})

})

// --------------------------------------------------------------------

func TestSuite(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "resp")
}

// --------------------------------------------------------------------

func MatchCommand(expected ...string) types.GomegaMatcher {
	return &commandMatcher{expected: expected}
}

func MatchStream(expected ...string) types.GomegaMatcher {
	return &streamMatcher{expected: expected}
}

type commandMatcher struct {
	expected []string
	actual   []string
}

func (m *commandMatcher) Match(actual interface{}) (bool, error) {
	cmd, ok := actual.(*resp.Command)
	if !ok {
		return false, fmt.Errorf("MatchCommand matcher expects a Command, but was %T", actual)
	}
	m.actual = cmdToSlice(cmd)
	return reflect.DeepEqual(m.expected, m.actual), nil
}

func (m *commandMatcher) FailureMessage(actual interface{}) string {
	return fmt.Sprintf("Expected\n\t%#v\n to match\n\t%#v", m.actual, m.expected)
}
func (m *commandMatcher) NegatedFailureMessage(actual interface{}) string {
	return fmt.Sprintf("Expected\n\t%#v\n not to match\n\t%#v", m.actual, m.expected)
}

type streamMatcher struct {
	expected []string
	actual   []string
}

func (m *streamMatcher) Match(actual interface{}) (bool, error) {
	cmd, ok := actual.(*resp.CommandStream)
	if !ok {
		return false, fmt.Errorf("MatchStream matcher expects a CommandStream, but was %T", actual)
	}

	buf := new(bytes.Buffer)
	m.actual = []string{cmd.Name}

	for cmd.More() {
		arg, err := cmd.Next()
		if err != nil {
			return false, fmt.Errorf("MatchStream failed to parse argument: %v", err)
		}

		buf.Reset()
		if _, err = buf.ReadFrom(arg); err != nil {
			return false, fmt.Errorf("MatchStream failed to read argument into buffer: %v", err)
		}
		m.actual = append(m.actual, buf.String())
	}
	return reflect.DeepEqual(m.expected, m.actual), nil
}

func (m *streamMatcher) FailureMessage(actual interface{}) string {
	return fmt.Sprintf("Expected\n\t%#v\n to match\n\t%#v", m.actual, m.expected)
}
func (m *streamMatcher) NegatedFailureMessage(actual interface{}) string {
	return fmt.Sprintf("Expected\n\t%#v\n not to match\n\t%#v", m.actual, m.expected)
}

func cmdToSlice(cmd *resp.Command) []string {
	res := []string{cmd.Name}
	for _, arg := range cmd.Args {
		res = append(res, arg.String())
	}
	return res
}

type ScannableStruct struct {
	Name     string
	Arity    int
	Flags    []string
	FirstKey int
	LastKey  int
	KeyStep  int
}

func (s *ScannableStruct) ScanResponse(t resp.ResponseType, r resp.ResponseReader) error {
	if t != resp.TypeArray {
		return fmt.Errorf("resp_test: unable to scan response, bad type: %s", t.String())
	}

	sz, err := r.ReadArrayLen()
	if err != nil {
		return err
	} else if sz != 6 {
		return fmt.Errorf("resp_test: expected 6 attributes, but received %d", sz)
	}

	return r.Scan(&s.Name, &s.Arity, &s.Flags, &s.FirstKey, &s.LastKey, &s.KeyStep)
}
