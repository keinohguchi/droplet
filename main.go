package main

import (
	"flag"
	"log"
	"os"
	"sync"
)

type request struct {
	cmd  string
	args []string
}

type reply struct {
	cmd  string
	args []string
	err  error
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
	flag.Parse()
	if *token == "" {
		log.Fatal("usage: droplet -t YOUR_API_TOKEN")
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
