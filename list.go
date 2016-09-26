// list command
package main

import (
	"github.com/digitalocean/godo"
)

func list(c *Server) ([]godo.Droplet, error) {
	var droplets []godo.Droplet

	opt := &godo.ListOptions{}
	for {
		dd, resp, err := c.do.Droplets.List(opt)
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
