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
	"github.com/digitalocean/godo/context"
)

type DataType int

const (
	invalid DataType = iota
	account
	droplet
	droplets
	keys
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
	ctx   context.Context
	token string
	*godo.Client
	*sync.WaitGroup
}

type actor func(s *Server, req *request)

var (
	requests = make(chan *request)
	replies  = make(chan *reply)
	actors   = make(map[string]actor)
)

func NewServer(ctx context.Context, token *string, n *sync.WaitGroup) (*Server, error) {
	opts := []godo.ClientOpt{}
	if *server != "" {
		opts = append(opts, godo.SetBaseURL(*server))
	}
	s := &Server{token: *token}
	c, err := godo.New(oauth2.NewClient(oauth2.NoContext, s), opts...)
	if err != nil {
		return nil, err
	}
	s.ctx = ctx
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
	if a, ok := actors[req.cmd]; ok {
		go a(s, req)
	} else {
		go noop(s, req)
	}
}

func init() {
	// action map, used in s.handle()
	actors["who"] = who
	actors["account"] = who
	actors["add"] = add
	actors["create"] = add
	actors["del"] = del
	actors["delete"] = del
	actors["delete-all"] = deleteAll
	actors["get"] = get
	actors["info"] = get
	actors["list"] = list
	actors["ls"] = list
	actors["keys"] = listKeys
}

func who(s *Server, req *request) {
	r := &reply{dataType: account}
	defer func() { replies <- r }()

	acct, _, err := s.Account.Get(s.ctx)
	if err != nil {
		r.err = err
		return
	}
	r.data, r.err = json.Marshal(acct)
}

func add(s *Server, req *request) {
	r := &reply{dataType: droplet}
	defer func() { replies <- r }()

	if len(req.args) < 3 {
		r.err = fmt.Errorf("droplet <name> <region> <fingerprint>")
		return
	}
	p := &godo.DropletCreateRequest{
		Name:              req.args[0],
		Region:            req.args[1],
		IPv6:              true,
		PrivateNetworking: true,
		Size:              "512mb",
		Image: godo.DropletCreateImage{
			Slug: "ubuntu-16-04-x64",
		},
		SSHKeys: []godo.DropletCreateSSHKey{
			{
				Fingerprint: req.args[2],
			},
		},
	}
	d, _, err := s.Droplets.Create(s.ctx, p)
	if err != nil {
		r.err = err
		return
	}
	r.data, r.err = json.Marshal(d)
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
	resp, err := s.Droplets.Delete(s.ctx, i)
	if err != nil {
		r.err = err
		return
	}
	r.data, r.err = json.Marshal(resp.Status)
}

func deleteAll(s *Server, req *request) {
	r := &reply{dataType: httpStatus}
	defer func() { replies <- r }()

	droplets, err := func() ([]godo.Droplet, error) {
		var droplets []godo.Droplet

		opt := &godo.ListOptions{}
		for {
			dd, resp, err := s.Droplets.List(s.ctx, opt)
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
		return
	}
	var resp *godo.Response
	for _, d := range droplets {
		resp, err = s.Droplets.Delete(s.ctx, d.ID)
		if err != nil {
			r.err = err
			return
		}
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
	d, _, err := s.Droplets.Get(s.ctx, i)
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
			dd, resp, err := s.Droplets.List(s.ctx, opt)
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
		return
	}
	r.data, r.err = json.Marshal(droplets)
}

func listKeys(s *Server, req *request) {
	r := &reply{dataType: keys}
	defer func() { replies <- r }()

	keys, err := func() ([]godo.Key, error) {
		var keys []godo.Key

		opt := &godo.ListOptions{}
		for {
			kk, resp, err := s.Keys.List(s.ctx, opt)
			if err != nil {
				return nil, err
			}
			for _, k := range kk {
				keys = append(keys, k)
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
		return keys, nil
	}()
	if err != nil {
		r.err = err
		return
	}
	r.data, r.err = json.Marshal(keys)
}

func noop(s *Server, req *request) {
	replies <- &reply{
		dataType: invalid,
		data:     nil,
		err:      fmt.Errorf("%q is not supported\n", req.cmd),
	}
}
