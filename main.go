package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
)

type request struct {
	cmd string
	args []string
}

type reply struct {
	status string
}

const (
	prompt = "droplet"
)

var (
	token    = flag.String("t", "", "DO APIv2 access token")
	inputs   = make(chan []string)
	outputs  = make(chan []string)
	requests = make(chan *request)
	replies  = make(chan *reply)
	abort    = make(chan struct{})
)

func main() {
	n := &sync.WaitGroup{}
	defer n.Wait()

	flag.Parse()
	if *token == "" {
		fmt.Fprintf(os.Stderr, "usage: droplet -t YOUR_API_TOKEN\n")
		os.Exit(1)
	}

	n.Add(1)
	go server(n)
	go clientHandler(os.Stdin, os.Stdout)

	// main loop.
	for {
		select {
		case args := <-inputs:
			// Just a back and forth for now :)
			n.Add(1)
			go clientWriter(args, n)
		case <-replies:
		case <-abort:
			for range inputs {
				// draint it!
			}
			close(requests)
			return
		}
	}
}

func clientWriter(args []string, n *sync.WaitGroup) {
	defer n.Done()
	outputs <- args
}

func clientHandler(in io.ReadCloser, out io.Writer) {
	prompt := func(w io.Writer) { fmt.Fprintf(w, "%s$ ", prompt) }
	input := bufio.NewScanner(in)
	for prompt(out); input.Scan(); prompt(out) {
		args := strings.Split(input.Text(), " ")
		switch args[0] {
		case "quit":
			close(inputs)
			close(abort)
		default:
			inputs <- args
			lines := <-outputs
			fmt.Fprintf(out, "%v\n", lines)
		}
	}
}
