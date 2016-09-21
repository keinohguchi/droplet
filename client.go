// client.go
package main

import (
	"fmt"

	"golang.org/x/oauth2"

	"github.com/digitalocean/godo"
)

type Client struct {
	token string
	do    *godo.Client
}

func (c *Client) Token() (*oauth2.Token, error) {
	token := &oauth2.Token{AccessToken: c.token}
	return token, nil
}

func (c *Client) open(token string) error {
	c.token = token
	oauthClient := oauth2.NewClient(oauth2.NoContext, c)
	c.do = godo.NewClient(oauthClient)
	return nil
}

func (c *Client) handle(cmd string) error {
	switch cmd {
	case "account":
		obj, err := c.account()
		if err != nil {
			return err
		}
		fmt.Printf("%s: %v\n", cmd, obj)
		return nil
	default:
		return fmt.Errorf("%q: not supported", cmd)
	}
}
