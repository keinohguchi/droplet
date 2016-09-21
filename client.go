// client.go
package main

import (
	"golang.org/x/oauth2"

	"github.com/digitalocean/godo"
)

type client struct{
	AccessToken string
	c           *godo.Client
}

func (c *client) Token() (*oauth2.Token, error) {
	token := &oauth2.Token{
		AccessToken: c.AccessToken,
	}
	return token, nil
}

func (c *client) open(token string) error {
	c.AccessToken = token
	oauthClient := oauth2.NewClient(oauth2.NoContext, c)
	c.c = godo.NewClient(oauthClient)
	return nil
}
