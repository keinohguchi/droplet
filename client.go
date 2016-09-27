// client
package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"sync"

	"github.com/digitalocean/godo"
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
		case "quit", "q":
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
				fmt.Fprintln(out, "Server disconnected")
				return
			}
			if reply.err != nil {
				fmt.Fprintln(out, reply.err)
				break
			}
			printReplyData(out, reply)
		}
	}
}

func printReplyData(out io.Writer, r *reply) {
	switch r.dataType {
	case account:
		var a godo.Account
		if err := json.Unmarshal(r.data, &a); err != nil {
			fmt.Fprintln(out, err)
		}
		fmt.Fprintln(out, a.Email, a.Status, a.DropletLimit)
	case droplet:
		var d godo.Droplet
		if err := json.Unmarshal(r.data, &d); err != nil {
			fmt.Fprintln(out, err)
		}
		fmt.Fprintln(out, d.ID, d.Name)
	case droplets:
		var dd []godo.Droplet
		if err := json.Unmarshal(r.data, &dd); err != nil {
			fmt.Fprintln(out, err)
		}
		for _, d := range dd {
			fmt.Fprintln(out, d.ID, d.Name)
		}
	case httpStatus:
		var status string
		if err := json.Unmarshal(r.data, &status); err != nil {
			fmt.Fprintln(out, err)
		}
		fmt.Fprintln(out, status)
	case invalid:
		fmt.Fprintf(out, "%s\n", r.data)
	default:
		fmt.Fprintf(out, "%s\n", r)
	}
}
