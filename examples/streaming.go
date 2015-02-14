/*

Streaming example:

  $ redis-cli -p 9736 ping
  PONG

  $ redis-cli -p 9736 file
  (error) ERR wrong number of arguments for 'file' command
  $ redis-cli -p 9736 file bad.txt
  (error) ERR no such file or directory

  $ echo -n "it works!" > /tmp/hi.txt
  $ redis-cli -p 9736 file hi.txt
  "it works!"

*/
package main

import (
	"log"
	"net/http"
	"path"

	"github.com/bsm/redeo"
)

func main() {
	dir := http.Dir("/tmp")
	srv := redeo.NewServer(nil)
	srv.HandleFunc("ping", func(out *redeo.Responder, _ *redeo.Request) error {
		out.WriteInlineString("PONG")
		return nil
	})
	srv.HandleFunc("file", func(out *redeo.Responder, req *redeo.Request) error {
		if len(req.Args) != 1 {
			return redeo.WrongNumberOfArgs(req.Name)
		}

		file, err := dir.Open(path.Clean(req.Args[0]))
		if err != nil {
			return err
		}

		stat, err := file.Stat()
		if err != nil {
			return err
		}

		_, err = out.StreamN(file, stat.Size())
		return err
	})

	log.Printf("Listening on tcp://%s", srv.Addr())
	log.Fatal(srv.ListenAndServe())
}
