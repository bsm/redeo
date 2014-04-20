package main

import (
	"github.com/bsm/redeo"
	"log"
)

func main() {
	srv := redeo.NewServer(nil)
	srv.HandleFunc("ping", func(out *redeo.Responder, _ *redeo.Request, ctx interface{}) error {
		out.WriteInlineString("PONG")
		return nil
	})

	log.Printf("Listening on tcp://%s", srv.Addr())
	log.Fatal(srv.ListenAndServe())
}
