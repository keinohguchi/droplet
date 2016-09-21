// account related methods
package main

import (
	"github.com/digitalocean/godo"
)

func (c *Client) account() (account *godo.Account, err error) {
	account, _, err = c.do.Account.Get()
	return
}
