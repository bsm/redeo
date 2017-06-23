package resp_test

import (
	"bytes"
	"fmt"
	"reflect"
	"testing"

	"github.com/bsm/redeo/resp"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/types"
)

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
	return fmt.Sprintf("Expected\n\t%#v\nto match\n\t%#v", m.actual, m.expected)
}
func (m *commandMatcher) NegatedFailureMessage(actual interface{}) string {
	return fmt.Sprintf("Expected\n\t%#v\nnot to match\n\t%#v", m.actual, m.expected)
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

	for i := 0; i < cmd.ArgN(); i++ {
		ar, err := cmd.NextArg()
		if err != nil {
			return false, fmt.Errorf("MatchStream failed to parse argument #%d: %v", i+1, err)
		}

		buf.Reset()
		if _, err = buf.ReadFrom(ar); err != nil {
			return false, fmt.Errorf("MatchStream failed to read argument into buffer: %v", err)
		}
		m.actual = append(m.actual, buf.String())
	}
	return reflect.DeepEqual(m.expected, m.actual), nil
}

func (m *streamMatcher) FailureMessage(actual interface{}) string {
	return fmt.Sprintf("Expected\n\t%#v\nto match\n\t%#v", m.actual, m.expected)
}
func (m *streamMatcher) NegatedFailureMessage(actual interface{}) string {
	return fmt.Sprintf("Expected\n\t%#v\nnot to match\n\t%#v", m.actual, m.expected)
}

func cmdToSlice(cmd *resp.Command) []string {
	res := []string{cmd.Name}
	for _, arg := range cmd.Args() {
		res = append(res, string(arg))
	}
	return res
}
