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

func NewServer(token *string, n *sync.WaitGroup) (*Server, error) {
	s := &Server{token: *token}
	s.do = godo.NewClient(oauth2.NewClient(oauth2.NoContext, s))
	n.Add(1)
	go s.loop(n)
	return s, nil
}

func (s *Server) Token() (*oauth2.Token, error) {
	token := &oauth2.Token{AccessToken: s.token}
	return token, nil
}

// loop goroutine waiting for the client requests channel.
func (s *Server) loop(n *sync.WaitGroup) {
	defer n.Done()
	for {
		select {
		case req, ok := <-requests:
			if !ok {
				log.Printf("server: requests channel closed\n")
				return
			}
			s.handle(req)
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

func (s *Server) handle(req *request) {
	switch req.cmd {
	case "setting":
		go func() {
			acct, err := setting(s)
			if err != nil {
				log.Fatal(err)
			} else {
				log.Print(acct)
			}
			replies <- &reply{cmd: req.cmd, args: req.args, err: err}
		}()
	case "list":
		go func() {
			droplets, err := list(s)
			if err != nil {
				log.Fatal(err)
			} else {
				for i, d := range droplets {
					log.Printf("%d: %v, %v\n", i+1, d.Name, d.Image.Slug)
				}
			}
			replies <- &reply{cmd: req.cmd, args: req.args, err: err}
		}()
	case "create":
		go func() {
			d, err := create(s)
			if err != nil {
				log.Print(err)
			} else {
				log.Printf("%v, %v\n", d.Name, d.Image.Slug)
			}
			replies <- &reply{cmd: req.cmd, args: req.args, err: err}
		}()
	default:
		go func() {
			log.Print(req)
			replies <- &reply{cmd: req.cmd, args: req.args, err: nil}
		}()
	}
}
