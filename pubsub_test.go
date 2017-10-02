package redeo

import (
	"github.com/bsm/redeo/redeotest"
	"github.com/bsm/redeo/resp"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("PubSubBroker", func() {
	var subject *PubSubBroker

	BeforeEach(func() {
		subject = NewPubSubBroker()
	})

	var publish = func(name, msg string) int64 {
		c := resp.NewCommand("subscribe", resp.CommandArgument(name), resp.CommandArgument(msg))
		w := redeotest.NewRecorder()
		subject.Publish().ServeRedeo(w, c)

		n, err := w.Response()
		Expect(err).NotTo(HaveOccurred())
		return n.(int64)
	}

	It("should publish/subscribe", func() {
		sub1 := redeotest.NewRecorder()
		sub2 := redeotest.NewRecorder()
		subc := resp.NewCommand("subscribe", resp.CommandArgument("chan"))

		Expect(subject.channels).To(HaveLen(0))
		Expect(publish("chan", "msg1")).To(Equal(int64(0)))

		subject.Subscribe().ServeRedeo(sub1, subc)
		Expect(sub1.Responses()).To(Equal([]interface{}{
			[]interface{}{"subscribe", "chan", int64(1)},
		}))

		Expect(subject.channels).To(HaveLen(1))
		Expect(subject.channels).To(HaveKey("chan"))
		Expect(subject.channels["chan"].subscribers).To(HaveLen(1))
		Expect(publish("chan", "msg2")).To(Equal(int64(1)))

		subject.Subscribe().ServeRedeo(sub2, subc)
		Expect(sub2.Responses()).To(Equal([]interface{}{
			[]interface{}{"subscribe", "chan", int64(1)},
		}))

		Expect(subject.channels).To(HaveLen(1))
		Expect(subject.channels).To(HaveKey("chan"))
		Expect(subject.channels["chan"].subscribers).To(HaveLen(2))
		Expect(publish("chan", "msg3")).To(Equal(int64(2)))

		Expect(sub1.Responses()).To(Equal([]interface{}{
			[]interface{}{"subscribe", "chan", int64(1)},
			[]interface{}{"message", "chan", "msg2"},
			[]interface{}{"message", "chan", "msg3"},
		}))
		Expect(sub2.Responses()).To(Equal([]interface{}{
			[]interface{}{"subscribe", "chan", int64(1)},
			[]interface{}{"message", "chan", "msg3"},
		}))
	})

})
