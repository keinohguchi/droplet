// server: server goroutine to access the backend
package main

import (
	"fmt"
	"log"
	"strconv"
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
	opts := []godo.ClientOpt{}
	if *server != "" {
		opts = append(opts, godo.SetBaseURL(*server))
	}
	do, err := godo.New(oauth2.NewClient(oauth2.NoContext, s), opts...)
	if err != nil {
		return nil, err
	}
	s.do = do
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
			if err == nil {
				log.Print(acct)
			}
			replies <- &reply{cmd: req.cmd, args: req.args, err: err}
		}()
	case "ls", "list":
		go func() {
			droplets, err := s.list()
			if err == nil {
				for _, d := range droplets {
					log.Printf("%d: %v, %v, %v\n",
						d.ID, d.Name, d.Region.Slug, d.Image.Slug)
				}
			}
			replies <- &reply{cmd: req.cmd, args: req.args, err: err}
		}()
	case "create":
		go func() {
			var err error
			if len(req.args) < 2 {
				err = fmt.Errorf("server: create <name> <region>\n")
			} else {
				p := create_param(req.args[0], req.args[1])
				d, _, err := s.do.Droplets.Create(p)
				if err != nil {
					log.Print(err)
				} else {
					log.Printf("%d, %v, %v, %v\n",
						d.ID, d.Name, d.Region.Slug, d.Image.Slug)
				}
			}
			replies <- &reply{cmd: req.cmd, args: req.args, err: err}
		}()
	case "info":
		go func() {
			var d *godo.Droplet
			var err error
			var i int
			if len(req.args) < 1 {
				err = fmt.Errorf("server: delete <droplet_id>\n")
			} else {
				i, err = strconv.Atoi(req.args[0])
				if err == nil {
					d, _, err = s.do.Droplets.Get(i)
					if err == nil {
						log.Printf("%d, %v, %v, %v\n",
							d.ID, d.Name, d.Region.Slug, d.Image.Slug)
					}
				}
			}
			replies <- &reply{cmd: req.cmd, args: req.args, err: err}
		}()
	case "delete", "rm":
		go func() {
			var err error
			var i int
			if len(req.args) < 1 {
				err = fmt.Errorf("server: delete <droplet_id>\n")
			} else {
				i, err = strconv.Atoi(req.args[0])
				if err == nil {
					_, err = s.do.Droplets.Delete(i)
				}
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

func create_param(name, region string) *godo.DropletCreateRequest {
	return &godo.DropletCreateRequest{
		Name:   name,
		Region: region,
		Size:   "512mb",
		Image: godo.DropletCreateImage{
			Slug: "ubuntu-14-04-x64",
		},
	}
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
