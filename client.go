// client
package main

import (
	"bufio"
	"fmt"
	"io"
	"strings"
	"sync"
)

const (
	prompt = "droplet> "
)

func clientHandler(in io.ReadCloser, out io.Writer, n *sync.WaitGroup) {
	defer n.Done()
	prompt := func(w io.Writer) { fmt.Fprint(w, prompt) }
	input := bufio.NewScanner(in)
	for prompt(out); input.Scan(); prompt(out) {
		args := strings.Split(input.Text(), " ")
		switch args[0] {
		case "quit":
			close(requests)
			close(abort)
			for range replies {
				// drain it!
			}
			return
		default:
			requests <- &request{cmd: args[0], args: args[1:]}
			reply, ok := <-replies
			if !ok {
				fmt.Fprintf(out, "Server disconnected\n")
				return
			}
			printReply(out, reply)
		}
	}
}

func printReply(out io.Writer, r *reply) {
	if r.err != nil {
		fmt.Fprint(out, r.err)
		return
	}
	switch r.dataType {
	case invalid:
		fmt.Fprintf(out, "%s\n", r.data)
	default:
		fmt.Fprintf(out, "%s\n", r)
	}
}
