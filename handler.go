package redeo

// Abstract handler interface
type Handler interface {
	ServeClient(out *Responder, req *Request) error
}

// Abstract handler function
type HandlerFunc func(out *Responder, req *Request) error

// ServeClient calls f(out, req).
func (f HandlerFunc) ServeClient(out *Responder, req *Request) error {
	return f(out, req)
}
