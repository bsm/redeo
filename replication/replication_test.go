package replication

import (
	"testing"

	"github.com/bsm/redeo/resp"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Replication", func() {
	var subject *Replication
	var stable StableStore
	var data *InMemKeyValueStore

	cmd1 := resp.NewCommand("SET", resp.CommandArgument("key"), resp.CommandArgument("value"))

	BeforeEach(func() {
		stable = NewInMemStableStore()
		data = NewInMemKeyValueStore()

		var err error
		subject, err = NewReplication("127.0.0.1:0", stable, data, nil)
		Expect(err).NotTo(HaveOccurred())
	})

	AfterEach(func() {
		Expect(subject.Close()).To(Succeed())
	})

	It("should persist on close", func() {
		subject.Propagate(cmd1)
		Expect(subject.Close()).To(Succeed())

		var state runtimeState
		Expect(stable.Decode("replication", &state)).To(Succeed())
		Expect(state).To(Equal(runtimeState{
			ID:     subject.masterID,
			Offset: 33,
		}))
	})

})

// ------------------------------------------------------------------------

func TestSuite(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "redeo/replication")
}
