package main

import (
	"flag"
	"fmt"
	"log"

	"golang.org/x/oauth2"

	"github.com/digitalocean/godo"
)

type TokenSource struct{
	AccessToken string
}
func (t *TokenSource) Token() (*oauth2.Token, error) {
	token := &oauth2.Token{
		AccessToken: t.AccessToken,
	}
	return token, nil
}

var token = flag.String("t", "", "DO APIv2 access token")

func main() {
	flag.Parse()
	tokenSource := &TokenSource{
		AccessToken: *token,
	}
	oauthClient := oauth2.NewClient(oauth2.NoContext, tokenSource)
	client := godo.NewClient(oauthClient)
	account, _, err := client.Account.Get()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(account)
}
