package main

import (
	"flag"
	"log"
	"net"

	"github.com/bsm/redeo"
)

var flags struct {
	addr string
}

func init() {
	flag.StringVar(&flags.addr, "addr", ":9736", "The TCP address to bind to")
}

func main() {
	flag.Parse()

	if err := run(); err != nil {
		log.Fatalln(err)
	}
}

func run() error {
	broker := redeo.NewPubSubBroker()
	srv := redeo.NewServer(nil)
	srv.Handle("ping", redeo.Ping())
	srv.Handle("info", redeo.Info(srv))
	srv.Handle("publish", broker.Publish())
	srv.Handle("subscribe", broker.Subscribe())

	lis, err := net.Listen("tcp", flags.addr)
	if err != nil {
		return err
	}
	defer lis.Close()

	log.Printf("waiting for connections on %s", lis.Addr().String())
	return srv.Serve(lis)
}
