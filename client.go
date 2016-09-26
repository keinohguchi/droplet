// client
package main

import (
	"bufio"
	"fmt"
	"io"
	"strings"
	"sync"
)

func clientHandler(in io.ReadCloser, out io.Writer, n *sync.WaitGroup) {
	defer n.Done()
	prompt := func(w io.Writer) { fmt.Fprintf(w, "%s$ ", prompt) }
	input := bufio.NewScanner(in)
	for prompt(out); input.Scan(); prompt(out) {
		args := strings.Split(input.Text(), " ")
		switch args[0] {
		case "quit":
			close(inputs)
			close(abort)
			for range outputs {
				// drain it!
			}
			return
		default:
			inputs <- args
			lines := <-outputs
			fmt.Fprintf(out, "%v\n", lines)
		}
	}
}
