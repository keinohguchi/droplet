// server: server goroutine to access the backend
package main

import (
	"log"
	"sync"
)

// server goroutine waiting for the client requests
// channel.
func server(n *sync.WaitGroup) {
	defer n.Done()
	for {
		select {
		case req := <-requests:
			log.Print(req)
		case <-abort:
			for range requests {
				// drain it!
			}
			return
		}
	}
}
