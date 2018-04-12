package info

import (
	"strconv"
	"sync/atomic"
)

// Value must be exportable as a string
type Value interface {
	String() string
}

// StaticString is the simplest value type
type StaticString string

func (v StaticString) String() string { return string(v) }

// StaticInt converts a static integer into a value
func StaticInt(n int64) Value { return StaticString(strconv.FormatInt(n, 10)) }

// --------------------------------------------------------------------

// Callback function
type Callback func() string

func (c Callback) String() string { return c() }

// --------------------------------------------------------------------

// IntValue is a int64 value with thread-safe atomic modifiers.
type IntValue struct{ n int64 }

// NewIntValue return a IntValue
func NewIntValue(n int64) *IntValue { return &IntValue{n: n} }

// Inc atomically increase the IntValue by delta
func (v *IntValue) Inc(delta int64) int64 { return atomic.AddInt64(&v.n, delta) }

// Set atomically set the value of IntValue to v
func (v *IntValue) Set(n int64) { atomic.StoreInt64(&v.n, n) }

// Value atomically get the value stored in IntValue
func (v *IntValue) Value() int64 { return atomic.LoadInt64(&v.n) }

// String return the value of IntValue as string
func (v *IntValue) String() string { return strconv.FormatInt(v.Value(), 10) }

// --------------------------------------------------------------------

// StringValue is a string value with thread-safe atomic modifiers.
type StringValue struct{ s atomic.Value }

// NewStringValue returns a StringValue
func NewStringValue(s string) *StringValue {
	v := &StringValue{}
	v.s.Store(s)
	return v
}

// Set atomically sets the value to s
func (v *StringValue) Set(s string) { v.s.Store(s) }

// String return the value.
func (v *StringValue) String() string { return v.s.Load().(string) }
