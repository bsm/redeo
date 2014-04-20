package redeo

// Abstract handler interface
type Handler interface {
	ServeClient(out *Responder, req *Request, ctx interface{}) error
}

// Abstract handler function
type HandlerFunc func(out *Responder, req *Request, ctx interface{}) error

// ServeClient calls f(out, req).
func (f HandlerFunc) ServeClient(out *Responder, req *Request, ctx interface{}) error {
	return f(out, req, ctx)
}
