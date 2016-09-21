// handler
package main

import (
	"fmt"
)

func handle(c *Client, cmd string) error {
	switch cmd {
	case "account":
		acct, err := account(c)
		if err != nil {
			return err
		}
		fmt.Printf("%s: %v\n", cmd, acct)
		return nil
	default:
		return fmt.Errorf("%q: not supported", cmd)
	}
}
