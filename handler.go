// handler
package main

import (
	"fmt"
)

func handle(c *Server, cmd string) error {
	switch cmd {
	case "setting":
		acct, err := setting(c)
		if err != nil {
			return err
		}
		fmt.Printf("%s: %v\n", cmd, acct)
	case "list":
		droplets, err := list(c)
		if err != nil {
			return err
		}
		for i, d := range droplets {
			fmt.Printf("%d: %v, %v\n", i + 1, d.Name, d.Image.Slug)
		}
	case "create":
		d, err := create(c)
		if err != nil {
			return err
		}
		fmt.Printf("%v, %v\n", d.Name, d.Image.Slug)
	default:
		return fmt.Errorf("%q: not supported", cmd)
	}
	return nil
}
