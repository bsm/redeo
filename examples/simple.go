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
	srv.HandleFunc("client", func(out *redeo.Responder, req *redeo.Request) error {
		if len(req.Args) != 1 {
			return redeo.WrongNumberOfArgs(req.Name)
		}

		switch req.Args[0] {
		case "list":
			out.WriteString(srv.Info().ClientsString())
		default:
			return redeo.UnknownCommand(req.Name)
		}
		return nil
	})

	log.Printf("Listening on tcp://%s", srv.Addr())
	log.Fatal(srv.ListenAndServe())
}
