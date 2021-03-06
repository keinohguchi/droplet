package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"sync"
)

var (
	token  = flag.String("t", "", "DO APIv2 access token")
	finger = flag.String("f", "", "SSH fingerprint")
	server = flag.String("s", "",
		"DO APIv2 server endpoint e.g. https://api.digitalocean.com")
	abort = make(chan struct{})
)

func main() {
	flag.Parse()
	if *token == "" {
		usage()
	}
	n := &sync.WaitGroup{}
	defer func() {
		log.Print("Waiting for all goroutines to finish...")
		n.Wait()
	}()

	// server to interact with the backend
	if _, err := NewServer(context.Background(), token, finger, n); err != nil {
		log.Fatal("main: can't create server\n")
	}
	n.Add(1)
	go clientHandler(os.Stdin, os.Stdout, n)

	// Waiting for the abort...
	select {
	case <-abort:
		return
	}
}

func usage() {
	fmt.Fprintln(os.Stderr, "usage: droplet -t YOUR_API_KEY -f YOUR_SSH_FINGERPRINT")
	os.Exit(1)
}
