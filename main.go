package main

import (
	"flag"
	"fmt"
	"log"
)

var token = flag.String("t", "", "DO APIv2 access token")

func main() {
	var c client

	flag.Parse()
	c.open(*token)
	account, _, err := c.c.Account.Get()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(account)
}
