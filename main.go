package main

import (
	"defibotgo/internal/models"
	"defibotgo/internal/web3"
	"log"
)

func main() {
	chain := models.Optimism
	ethClient, err := web3.BuildWeb3Client(chain, true)

	if err != nil {
		log.Fatalf("Error building eth cient: %s", err)
	}
	log.Printf("Successfully built eth client on %s %p", chain, ethClient)
}
