// server: server goroutine to access the backend
package main

import (
	"fmt"
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
	case "who", "whoami":
		go func() {
			acct, _, err := s.do.Account.Get()
			if err != nil {
				log.Fatal(err)
			} else {
				log.Print(acct)
			}
			replies <- &reply{cmd: req.cmd, args: req.args, err: err}
		}()
	case "ls", "list":
		go func() {
			droplets, err := s.list()
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
		if len(req.args) < 1 {
			e := fmt.Errorf("server: create needs droplet's name\n")
			replies <- &reply{cmd: req.cmd, args: req.args, err: e}
		} else {
			go func() {
				d, err := s.create()
				if err != nil {
					log.Print(err)
				} else {
					log.Printf("%v, %v\n", d.Name, d.Image.Slug)
				}
				replies <- &reply{cmd: req.cmd, args: req.args, err: err}
			}()
		}
	default:
		go func() {
			log.Print(req)
			replies <- &reply{cmd: req.cmd, args: req.args, err: nil}
		}()
	}
}

func (s *Server) create() (d *godo.Droplet, err error) {
	param := func (name string, region string, slug string) *godo.DropletCreateRequest {
		return &godo.DropletCreateRequest{
			Name:   name,
			Region: region,
			Size:   "512mb",
			Image: godo.DropletCreateImage{
				Slug: slug,
			},
		}
	}
	d, _, err = s.do.Droplets.Create(param("test", "sfo1", "ubuntu-14-04-x64"))
	return d, err
}

func (s *Server) list() ([]godo.Droplet, error) {
	var droplets []godo.Droplet

	opt := &godo.ListOptions{}
	for {
		dd, resp, err := s.do.Droplets.List(opt)
		if err != nil {
			return nil, err
		}
		for _, d := range dd {
			droplets = append(droplets, d)
		}

		if resp.Links == nil || resp.Links.IsLastPage() {
			break
		}

		page, err := resp.Links.CurrentPage()
		if err != nil {
			return nil, err
		}

		opt.Page = page + 1
	}

	return droplets, nil
}
