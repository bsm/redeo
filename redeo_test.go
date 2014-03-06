package redeo

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"testing"
)

/*************************************************************************
 * GINKGO TEST HOOK
 *************************************************************************/

func TestSuite(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "redeo")
}
