// account related methods
package main

import (
	"github.com/digitalocean/godo"
)

func account(c *Client) (account *godo.Account, err error) {
	account, _, err = c.do.Account.Get()
	return
}
