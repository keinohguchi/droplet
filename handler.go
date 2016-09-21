// handler
package main

import (
	"fmt"
)

func handle(c *Client, cmd string) error {
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
