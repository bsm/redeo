package main

import (
	"github.com/bsm/redeo"
	"log"
)

func main() {
	srv := redeo.NewServer(nil)
	srv.HandleFunc("ping", func(out *redeo.Responder, _ *redeo.Request) error {
		_, err := out.WriteInlineString("PONG")
		return err
	})

	log.Printf("Listening on %s://%s", srv.Proto(), srv.Addr())
	log.Fatal(srv.ListenAndServe())
}
