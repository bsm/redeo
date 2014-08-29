package main

import (
	"log"

	"github.com/bsm/redeo"
)

func main() {
	srv := redeo.NewServer(nil)
	srv.HandleFunc("ping", func(out *redeo.Responder, _ *redeo.Request) error {
		out.WriteInlineString("PONG")
		return nil
	})
	srv.HandleFunc("info", func(out *redeo.Responder, _ *redeo.Request) error {
		out.WriteString(srv.Info().String())
		return nil
	})

	log.Printf("Listening on tcp://%s", srv.Addr())
	log.Fatal(srv.ListenAndServe())
}
