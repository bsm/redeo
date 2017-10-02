package redeotest_test

import (
	"fmt"

	"github.com/bsm/redeo"
	"github.com/bsm/redeo/redeotest"
	"github.com/bsm/redeo/resp"
)

func ExampleResponseRecorder() {
	handler := func(w resp.ResponseWriter, c *resp.Command) {
		if c.ArgN() != 1 {
			w.AppendError(redeo.WrongNumberOfArgs(c.Name))
			return
		}

		w.AppendArrayLen(3)
		w.AppendBulk(c.Arg(0).Bytes())
		w.AppendBulkString("responds to")
		w.AppendBulkString(c.Name)
	}

	w := redeotest.NewRecorder()
	handler(w, resp.NewCommand("call", resp.CommandArgument("bob")))
	fmt.Println(w.Quoted())
	fmt.Println(w.Response())

	// Output:
	// "*3\r\n$3\r\nbob\r\n$11\r\nresponds to\r\n$4\r\ncall\r\n"
	// [bob responds to call] <nil>
}
