// server: server goroutine to access the backend
package main

import (
	"log"
	"sync"

	"golang.org/x/oauth2"

	"github.com/digitalocean/godo"
)

type Server struct {
	token string
	do    *godo.Client
}

func (s *Server) Token() (*oauth2.Token, error) {
	token := &oauth2.Token{AccessToken: s.token}
	return token, nil
}

func open(token *string) (*godo.Client, error) {
	s := Server{token: *token}
	oauthClient := oauth2.NewClient(oauth2.NoContext, &s)
	s.do = godo.NewClient(oauthClient)
	return s.do, nil
}

// server goroutine waiting for the client requests channel.
func server(token *string, n *sync.WaitGroup) {
	defer n.Done()
	if _, err := open(token); err != nil {
		log.Fatal(err)
	}
	for {
		select {
		case req, ok := <-requests:
			if !ok {
				log.Printf("server: requests channel closed\n")
				return
			}
			log.Print(req)
			reply := reply{status: "good", args: req.args}
			go func() {
				replies <- &reply
			}()
		case <-abort:
			log.Printf("server: aborting...\n")
			for range requests {
				// drain it!
			}
			log.Printf("server: requests channel is clean\n")
			return
		}
	}
}
