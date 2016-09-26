// create command
package main

import (
	"github.com/digitalocean/godo"
)

func requestParam(name string, region string, slug string) *godo.DropletCreateRequest {
	return &godo.DropletCreateRequest{
		Name:   name,
		Region: region,
		Size:   "512mb",
		Image: godo.DropletCreateImage{
			Slug: slug,
		},
	}
}

func create(c *Server) (d *godo.Droplet, err error) {
	d, _, err = c.do.Droplets.Create(requestParam("test", "sfo1", "ubuntu-14-04-x64"))
	return
}
