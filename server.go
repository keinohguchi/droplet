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

type Server struct {
	token string
	*godo.Client
	*sync.WaitGroup
}

type action func (s *Server, req *request)

var (
	requests = make(chan *request)
	replies  = make(chan *reply)
	actions  = make(map[string]action)
)

func NewServer(token *string, n *sync.WaitGroup) (*Server, error) {
	opts := []godo.ClientOpt{}
	if *server != "" {
		opts = append(opts, godo.SetBaseURL(*server))
	}
	s := &Server{token: *token}
	c, err := godo.New(oauth2.NewClient(oauth2.NoContext, s), opts...)
	if err != nil {
		return nil, err
	}
	s.Client = c
	s.WaitGroup = n
	s.Add(1)
	go s.loop()
	return s, nil
}

func (s *Server) Token() (*oauth2.Token, error) {
	token := &oauth2.Token{AccessToken: s.token}
	return token, nil
}

// loop goroutine waiting for the client requests channel.
func (s *Server) loop() {
	defer func() {
		close(replies)
		log.Printf("server: close replies channel\n")
		s.Done()
	}()
	for {
		select {
		case req, ok := <-requests:
			if !ok {
				log.Printf("server: requests channel was closed\n")
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
	if a, ok := actions[req.cmd]; ok {
		go a(s, req)
	} else {
		go noop(s, req)
	}
}

func init() {
	// action map, used in s.handle()
	actions["who"]     = who
	actions["account"] = who
	actions["add"]     = add
	actions["create"]  = add
	actions["del"]     = del
	actions["delete"]  = del
	actions["get"]     = get
	actions["info"]    = get
	actions["list"]    = list
	actions["ls"]      = list
}

func who(s *Server, req *request) {
	r := &reply{dataType: account}
	defer func() { replies <- r }()

	acct, _, err := s.Account.Get()
	if err != nil {
		r.err = err
	} else {
		r.data, r.err = json.Marshal(acct)
	}
}

func add(s *Server, req *request) {
	r := &reply{dataType: droplet}
	defer func() { replies <- r }()

	if len(req.args) < 2 {
		r.err = fmt.Errorf("droplet <name> <region>")
		return
	}
	p := &godo.DropletCreateRequest{
		Name:   req.args[0],
		Region: req.args[1],
		Size:   "512mb",
		Image: godo.DropletCreateImage{
			Slug: "ubuntu-14-04-x64",
		},
	}
	d, _, err := s.Droplets.Create(p)
	if err != nil {
		r.err = err
	} else {
		r.data, r.err = json.Marshal(d)
	}
}

func del(s *Server, req *request) {
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
	resp, err := s.Droplets.Delete(i)
	if err != nil {
		r.err = err
		return
	}
	r.data, r.err = json.Marshal(resp.Status)
}

func get(s *Server, req *request) {
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
	d, _, err := s.Droplets.Get(i)
	if err != nil {
		r.err = err
		return
	}
	r.data, r.err = json.Marshal(d)
}

func list(s *Server, req *request) {
	r := &reply{dataType: droplets}
	defer func() { replies <- r }()

	droplets, err := func() ([]godo.Droplet, error) {
		var droplets []godo.Droplet

		opt := &godo.ListOptions{}
		for {
			dd, resp, err := s.Droplets.List(opt)
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
	}()
	if err != nil {
		r.err = err
	} else {
		r.data, r.err = json.Marshal(droplets)
	}
}

func noop(s *Server, req *request) {
	replies <- &reply{
		dataType: invalid,
		data:     nil,
		err:      fmt.Errorf("%q is not supported\n", req.cmd),
	}
}
