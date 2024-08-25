package main

import (
	"defibotgo/internal/models"
	"defibotgo/internal/web3"
	"log"
)

func main() {
	chain := models.Optimism
	client, err := web3.BuildWeb3Client(chain)

	if err != nil {
		log.Fatalf("Error building web3 client: %s", err)
	}
	log.Printf("Successfully built web3 client on %s %p", chain, client)
}
