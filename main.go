package main

import (
	"flag"
	"log"
	"os"
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
	flag.Parse()
	if *token == "" {
		log.Fatal("usage: droplet -t YOUR_API_TOKEN")
	}

	n := &sync.WaitGroup{}
	defer func() {
		log.Print("Waiting all goroutines to finish...")
		n.Wait()
	}()

	n.Add(1)
	go server(token, n)
	n.Add(1)
	go clientHandler(os.Stdin, os.Stdout, n)

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
			close(outputs)
			return
		}
	}
}
