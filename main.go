package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
)

var token = flag.String("t", "", "DO APIv2 access token")

func main() {
	var c Client

	flag.Parse()
	if *token == "" {
		fmt.Fprintf(os.Stderr, "usage: droplet -t YOUR_API_TOKEN\n")
		os.Exit(1)
	}

	if err := c.open(*token); err != nil {
		log.Fatal(err)
	}

	// Simple command handler.
	input := bufio.NewScanner(os.Stdin)
	for prompt(os.Stdout); input.Scan(); prompt(os.Stdout) {
		if err := c.handle(input.Text()); err != nil {
			log.Fatal(err)
		}
	}
}

func prompt(out io.Writer) {
	const prompt = "Droplet"
	fmt.Fprintf(out, "%s> ", prompt)
}
