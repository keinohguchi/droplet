// client.go
package main

import (
	"golang.org/x/oauth2"

	"github.com/digitalocean/godo"
)

type client struct{
	token string
	c     *godo.Client
}

func (c *client) Token() (*oauth2.Token, error) {
	token := &oauth2.Token{ AccessToken: c.token }
	return token, nil
}

func (c *client) open(token string) error {
	c.token = token
	oauthClient := oauth2.NewClient(oauth2.NoContext, c)
	c.c = godo.NewClient(oauthClient)
	return nil
}
