package replication

import (
	"io/ioutil"
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("InMemKeyValueStore", func() {
	var subject *InMemKeyValueStore
	var _ DataStore = subject

	BeforeEach(func() {
		subject = NewInMemKeyValueStore()
	})

	It("should get/set", func() {
		Expect(subject.Keys()).To(BeEmpty())

		val, ok := subject.Get("notthere")
		Expect(val).To(BeNil())
		Expect(ok).To(BeFalse())

		subject.Set("alpha", []byte("value"))
		val, ok = subject.Get("alpha")
		Expect(val).To(Equal([]byte("value")))
		Expect(ok).To(BeTrue())

		subject.Set("beta", []byte("valux"))
		val, ok = subject.Get("beta")
		Expect(val).To(Equal([]byte("valux")))
		Expect(ok).To(BeTrue())

		Expect(subject.Keys()).To(ConsistOf("alpha", "beta"))
	})

})

var _ = Describe("InMemStableStore", func() {
	var subject StableStore

	BeforeEach(func() {
		subject = NewInMemStableStore()
	})

	It("should decode/encode", func() {
		var v, w struct {
			A string
			B int
		}
		v.A = "test"
		v.B = 33

		Expect(subject.Decode("k1", &w)).To(MatchError(ErrNotFound))
		Expect(subject.Encode("k1", &v)).To(Succeed())
		Expect(subject.Decode("k1", &w)).To(Succeed())
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

		Expect(subject.Decode("k1", &w)).To(MatchError(ErrNotFound))
		Expect(subject.Encode("k1", &v)).To(Succeed())
		Expect(subject.Decode("k1", &w)).To(Succeed())
		Expect(w).To(Equal(v))
	})

})
