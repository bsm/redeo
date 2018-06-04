package replication

import (
	"io/ioutil"
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("InMemStore", func() {
	var subject StableStore

	BeforeEach(func() {
		subject = NewInMemStableStore()
	})

	It("should get/set", func() {
		var v, w struct {
			A string
			B int
		}
		v.A = "test"
		v.B = 33

		Expect(subject.Get("k1", &w)).To(MatchError(ErrNotFound))
		Expect(subject.Set("k1", &v)).To(Succeed())
		Expect(subject.Get("k1", &w)).To(Succeed())
		Expect(w).To(Equal(v))
	})

})

var _ = Describe("FSStableStore", func() {
	var subject StableStore
	var dir string

	BeforeEach(func() {
		var err error
		dir, err = ioutil.TempDir("", "redeo-replication-test")
		Expect(err).NotTo(HaveOccurred())

		subject = NewFSStableStore(dir)
	})

	AfterEach(func() {
		Expect(os.RemoveAll(dir)).To(Succeed())
	})

	It("should get/set", func() {
		var v, w struct {
			A string
			B int
		}
		v.A = "test"
		v.B = 33

		Expect(subject.Get("k1", &w)).To(MatchError(ErrNotFound))
		Expect(subject.Set("k1", &v)).To(Succeed())
		Expect(subject.Get("k1", &w)).To(Succeed())
		Expect(w).To(Equal(v))
	})

})
