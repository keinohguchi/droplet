package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"sync"
)

const (
	prompt = "droplet"
)

var (
	token    = flag.String("t", "", "DO APIv2 access token")
	server   = flag.String("s", "",
		"DO APIv2 server endpoint e.g. https://api.digitalocean.com")
	inputs   = make(chan []string)
	outputs  = make(chan []string)
	abort    = make(chan struct{})
)

func main() {
	flag.Parse()
	if *token == "" {
		fmt.Fprintf(os.Stderr, "usage: droplet -t YOUR_API_TOKEN\n")
		os.Exit(1)
	}

	n := &sync.WaitGroup{}
	defer func() {
		for range inputs {
			// draint inputs channel
		}
		close(requests)
		close(outputs)
		log.Print("Waiting for all goroutines to finish...")
		n.Wait()
	}()

	// server to interact with the backend
	if _, err := NewServer(token, n); err != nil {
		log.Fatal("main: can't create server\n")
	}
	n.Add(1)
	go clientHandler(os.Stdin, os.Stdout, n)

	// main loop.
	for {
		select {
		case args, ok := <-inputs:
			if !ok {
				log.Printf("main: inputs channel has been closed\n")
				return
			}
			req := request{cmd: args[0], args: args[1:]}
			go func(req request) {
				requests <- &req
			}(req)
		case reply, ok := <-replies:
			if !ok {
				log.Printf("main: replies channel has been closed\n")
				return
			}
			go func() {
				if reply.err != nil {
					outputs <- []string{reply.err.Error()}
				} else {
					outputs <- reply.args
				}
			}()
		case <-abort:
			return
		}
	}
}
