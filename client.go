// client
package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"sync"
	"text/tabwriter"

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
			break
		}
		printAccounts(out, a)
	case droplet:
		var d godo.Droplet
		if err := json.Unmarshal(r.data, &d); err != nil {
			fmt.Fprintln(out, err)
			break
		}
		printDroplets(out, d)
	case droplets:
		var dd []godo.Droplet
		if err := json.Unmarshal(r.data, &dd); err != nil {
			fmt.Fprintln(out, err)
			break
		}
		printDroplets(out, dd...)
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

func printAccounts(out io.Writer, accounts ...godo.Account) {
	const format = "%v\t%v\t%v\n"
	tw := new(tabwriter.Writer).Init(out, 0, 8, 2, ' ', 0)
	fmt.Fprintf(tw, format, "E-mail", "Status", "Droplet Limit")
	fmt.Fprintf(tw, format, "------", "------", "-------------")
	for _, a := range accounts {
		fmt.Fprintf(tw, format, a.Email, a.Status, a.DropletLimit)
	}
	tw.Flush()
}

func printDroplets(out io.Writer, droplets ...godo.Droplet) {
	const format = "%v\t%v\t%v\n"
	tw := new(tabwriter.Writer).Init(out, 0, 8, 2, ' ', 0)
	fmt.Fprintf(tw, format, "Identifier", "Droplet Name", "Region")
	fmt.Fprintf(tw, format, "----------", "------------", "------")
	for _, d := range droplets {
		fmt.Fprintf(tw, format, d.ID, d.Name, d.Region.Slug)
	}
	tw.Flush()
}
