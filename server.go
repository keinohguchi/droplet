// server: server goroutine to access the backend
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"sync"

	"golang.org/x/oauth2"

	"github.com/digitalocean/godo"
)

type DataType int

const (
	invalid DataType = iota
	account
	droplet
	droplets
	httpStatus
)

type request struct {
	cmd  string
	args []string
}

type reply struct {
	dataType DataType
	data     []byte
	err      error
}

var (
	requests = make(chan *request)
	replies  = make(chan *reply)
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
	defer func() {
		close(replies)
		log.Printf("server: replies channel has been closed\n")
		n.Done()
	}()
	for {
		select {
		case req, ok := <-requests:
			if !ok {
				log.Printf("server: requests channel was closed already\n")
				return
			}
			s.handle(req)
		case <-abort:
			log.Printf("server: aborting...\n")
			for range requests {
				// drain it!
			}
			return
		}
	}
}

func (s *Server) handle(req *request) {
	switch req.cmd {
	case "who", "account":
		go func() {
			r := &reply{dataType: account}
			defer func() { replies <- r }()

			acct, _, err := s.do.Account.Get()
			if err != nil {
				r.err = err
			} else {
				r.data, r.err = json.Marshal(acct)
			}
		}()
	case "ls", "list":
		go func() {
			r := &reply{dataType: droplets}
			defer func() { replies <- r }()

			droplets, err := s.list()
			if err != nil {
				r.err = err
			} else {
				r.data, r.err = json.Marshal(droplets)
			}
		}()
	case "add", "create":
		go func() {
			r := &reply{dataType: droplet}
			defer func() { replies <- r }()

			if len(req.args) < 2 {
				r.err = fmt.Errorf("droplet <name> <region>")
				return
			}
			p := create_param(req.args[0], req.args[1])
			d, _, err := s.do.Droplets.Create(p)
			if err != nil {
				r.err = err
			} else {
				r.data, r.err = json.Marshal(d)
			}
		}()
	case "get", "info":
		go func() {
			r := &reply{dataType: droplet}
			defer func() { replies <- r }()

			if len(req.args) < 1 {
				r.err = fmt.Errorf("server: delete <droplet_id>\n")
				return
			}
			i, err := strconv.Atoi(req.args[0])
			if err != nil {
				r.err = err
				return
			}
			d, _, err := s.do.Droplets.Get(i)
			if err != nil {
				r.err = err
				return
			}
			r.data, r.err = json.Marshal(d)
		}()
	case "rm", "delete":
		go func() {
			r := &reply{dataType: httpStatus}
			defer func() { replies <- r }()

			if len(req.args) < 1 {
				r.err = fmt.Errorf("server: delete <droplet_id>\n")
				return
			}
			i, err := strconv.Atoi(req.args[0])
			if err != nil {
				r.err = err
				return
			}
			resp, err := s.do.Droplets.Delete(i)
			if err != nil {
				r.err = err
				return
			}
			r.data, r.err = json.Marshal(resp.Status)
		}()
	default:
		go func() {
			replies <- &reply{dataType: invalid, data: nil,
				err: fmt.Errorf("%s not supported\n", req.cmd)}
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
