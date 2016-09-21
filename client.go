// client.go
package main

import (
	"fmt"
	"log"

	"golang.org/x/oauth2"

	"github.com/digitalocean/godo"
)

type Client struct{
	token string
	c     *godo.Client
}

func (c *Client) Token() (*oauth2.Token, error) {
	token := &oauth2.Token{ AccessToken: c.token }
	return token, nil
}

func (c *Client) open(token string) error {
	c.token = token
	oauthClient := oauth2.NewClient(oauth2.NoContext, c)
	c.c = godo.NewClient(oauthClient)
	return nil
}

func (c *Client) handle(cmd string) error {
	switch cmd {
	case "account":
		account, _, err := c.c.Account.Get()
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("%s: %v\n", cmd, account)
		return nil
	default:
		return fmt.Errorf("%q: not supported", cmd)
	}
}
