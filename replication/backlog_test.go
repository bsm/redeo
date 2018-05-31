package replication

import (
	"io"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("backlog", func() {
	var subject *backlog
	var alphabet = []byte("abcdefghijklmnopqrstuvwxyz")

	BeforeEach(func() {
		subject = newBacklog(320640, 0)
	})

	It("should write", func() {
		subject.buffer = subject.buffer[:60] // reduce the buffer to 60 bytes

		Expect(subject.Write(alphabet)).To(Equal(26))
		Expect(subject.pos).To(Equal(26))
		Expect(subject.histlen).To(Equal(26))
		Expect(subject.offset).To(Equal(int64(320640)))

		Expect(subject.Write(alphabet)).To(Equal(26))
		Expect(subject.pos).To(Equal(52))
		Expect(subject.histlen).To(Equal(52))
		Expect(subject.offset).To(Equal(int64(320640)))

		Expect(subject.Write(alphabet)).To(Equal(26))
		Expect(subject.pos).To(Equal(18))
		Expect(subject.histlen).To(Equal(60))
		Expect(subject.offset).To(Equal(int64(320658)))
	})

	It("should read-at (simple)", func() {
		subject.buffer = subject.buffer[:60] // reduce the buffer to 26 bytes
		buf := make([]byte, 8)

		// --- write 20 bytes ---
		Expect(subject.Write(alphabet[:20])).To(Equal(20))

		n, err := subject.ReadAt(buf, 320639)
		Expect(n).To(Equal(0))
		Expect(err).To(Equal(ErrOffsetOutOfBounds))

		n, err = subject.ReadAt(buf, 320661)
		Expect(n).To(Equal(0))
		Expect(err).To(Equal(ErrOffsetOutOfBounds))

		Expect(subject.ReadAt(buf, 320640)).To(Equal(8))
		Expect(string(buf)).To(Equal(`abcdefgh`))

		Expect(subject.ReadAt(buf, 320641)).To(Equal(8))
		Expect(string(buf)).To(Equal(`bcdefghi`))

		Expect(subject.ReadAt(buf, 320650)).To(Equal(8))
		Expect(string(buf)).To(Equal(`klmnopqr`))

		n, err = subject.ReadAt(buf, 320656)
		Expect(n).To(Equal(4))
		Expect(err).To(Equal(io.EOF))
		Expect(string(buf[:n])).To(Equal(`qrst`))

		n, err = subject.ReadAt(buf, 320659)
		Expect(n).To(Equal(1))
		Expect(err).To(Equal(io.EOF))
		Expect(string(buf[:n])).To(Equal(`t`))

		n, err = subject.ReadAt(buf, 320660)
		Expect(n).To(Equal(0))
		Expect(err).To(Equal(io.EOF))

		// --- write 6 more bytes ---
		Expect(subject.Write(alphabet[20:])).To(Equal(6))

		n, err = subject.ReadAt(buf, 320660)
		Expect(n).To(Equal(6))
		Expect(err).To(Equal(io.EOF))
		Expect(string(buf[:n])).To(Equal(`uvwxyz`))

		n, err = subject.ReadAt(buf, 320665)
		Expect(n).To(Equal(1))
		Expect(err).To(Equal(io.EOF))
		Expect(string(buf[:n])).To(Equal(`z`))

		n, err = subject.ReadAt(buf, 320667)
		Expect(n).To(Equal(0))
		Expect(err).To(Equal(ErrOffsetOutOfBounds))
	})

	It("should read-at (overflow)", func() {
		subject.buffer = subject.buffer[:16] // reduce the buffer to 16 bytes
		buf := make([]byte, 8)

		// --- write 26 bytes ---
		Expect(subject.Write(alphabet)).To(Equal(26))

		n, err := subject.ReadAt(buf, 320649)
		Expect(n).To(Equal(0))
		Expect(err).To(Equal(ErrOffsetOutOfBounds))

		n, err = subject.ReadAt(buf, 320667)
		Expect(n).To(Equal(0))
		Expect(err).To(Equal(ErrOffsetOutOfBounds))

		Expect(subject.ReadAt(buf, 320650)).To(Equal(8))
		Expect(string(buf)).To(Equal(`klmnopqr`))
		Expect(subject.ReadAt(buf, 320654)).To(Equal(8))
		Expect(string(buf)).To(Equal(`opqrstuv`))
		Expect(subject.ReadAt(buf, 320658)).To(Equal(8))
		Expect(string(buf)).To(Equal(`stuvwxyz`))

		n, err = subject.ReadAt(buf, 320660)
		Expect(n).To(Equal(6))
		Expect(err).To(Equal(io.EOF))
		Expect(string(buf[:n])).To(Equal(`uvwxyz`))

		n, err = subject.ReadAt(buf, 320664)
		Expect(n).To(Equal(2))
		Expect(err).To(Equal(io.EOF))
		Expect(string(buf[:n])).To(Equal(`yz`))

		n, err = subject.ReadAt(buf, 320666)
		Expect(n).To(Equal(0))
		Expect(err).To(Equal(io.EOF))
	})

	It("should resize", func() {
		Expect(subject.Write(alphabet)).To(Equal(26))

		subject.Resize(minBacklogSize)
		Expect(subject.offset).To(Equal(int64(320640)))
		Expect(subject.pos).To(Equal(26))
		Expect(subject.histlen).To(Equal(26))

		subject.Resize(minBacklogSize * 2)
		Expect(subject.offset).To(Equal(int64(320666)))
		Expect(subject.pos).To(Equal(0))
		Expect(subject.histlen).To(Equal(0))
	})

})
