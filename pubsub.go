package redeo

import (
	"sync"
	"sync/atomic"

	"github.com/bsm/redeo/v2/resp"
)

// PubSubBroker can be used to emulate redis'
// native pub/sub functionality
type PubSubBroker struct {
	channels map[string]*pubSubChannel
	mu       sync.RWMutex
}

// NewPubSubBroker inits a new pub-sub broker
func NewPubSubBroker() *PubSubBroker {
	return &PubSubBroker{
		channels: make(map[string]*pubSubChannel),
	}
}

// Subscribe returns a subscribe handler
func (b *PubSubBroker) Subscribe() Handler {
	return HandlerFunc(func(w resp.ResponseWriter, c *resp.Command) {
		if c.ArgN() != 1 {
			w.AppendError(WrongNumberOfArgs(c.Name))
			return
		}
		b.subscribe(c.Arg(0).String(), w)
	})
}

// Publish acts as a publish handler
func (b *PubSubBroker) Publish() Handler {
	return HandlerFunc(func(w resp.ResponseWriter, c *resp.Command) {
		if c.ArgN() != 2 {
			w.AppendError(WrongNumberOfArgs(c.Name))
			return
		}

		n := b.PublishMessage(c.Arg(0).String(), c.Arg(1).String())
		w.AppendInt(n)
	})
}

// PublishMessage allows to publish a message to the broker
// outside the command-cycle. Returns the number of subscribers
func (b *PubSubBroker) PublishMessage(name, msg string) int64 {
	b.mu.RLock()
	ch, ok := b.channels[name]
	b.mu.RUnlock()

	if ok {
		return ch.Publish(name, msg)
	}
	return 0
}

func (b *PubSubBroker) subscribe(name string, w resp.ResponseWriter) {
	b.mu.RLock()
	ch, ok := b.channels[name]
	b.mu.RUnlock()

	if !ok {
		b.mu.Lock()
		if ch, ok = b.channels[name]; !ok {
			ch = &pubSubChannel{
				subscribers: make(map[int64]resp.ResponseWriter),
			}
			b.channels[name] = ch
		}
		b.mu.Unlock()
	}

	ch.Subscribe(w)
	w.AppendArrayLen(3)
	w.AppendBulkString("subscribe")
	w.AppendBulkString(name)
	w.AppendInt(1)
}

// --------------------------------------------------------------------

type pubSubChannel struct {
	subscribers map[int64]resp.ResponseWriter
	mu          sync.RWMutex
	nextID      int64
}

func (c *pubSubChannel) Subscribe(w resp.ResponseWriter) {
	sid := atomic.AddInt64(&c.nextID, 1)

	c.mu.Lock()
	c.subscribers[sid] = w
	c.mu.Unlock()
}

func (c *pubSubChannel) Publish(name, msg string) (n int64) {
	var failed []int64

	c.mu.RLock()
	for sid, w := range c.subscribers {
		w.AppendArrayLen(3)
		w.AppendBulkString("message")
		w.AppendBulkString(name)
		w.AppendBulkString(msg)

		if err := w.Flush(); err != nil {
			failed = append(failed, sid)
		} else {
			n++
		}
	}
	c.mu.RUnlock()

	if len(failed) != 0 {
		c.evict(failed)
	}
	return
}

func (c *pubSubChannel) evict(failed []int64) {
	c.mu.Lock()
	for _, sid := range failed {
		delete(c.subscribers, sid)
	}
	c.mu.Unlock()
}
