package main

import (
	"defibotgo/internal/config"
	"defibotgo/internal/models"
	protocolconfig "defibotgo/internal/protocols/config"
	"defibotgo/internal/protocols/tarot"
	"defibotgo/internal/web3"
	"flag"
	"github.com/ethereum/go-ethereum/crypto"
	"log"
	"strings"
)

var validChains = map[models.Chain]bool{
	models.Optimism: true,
	models.Base:     true,
}

func main() {
	var protocol *models.TarotOpts
	var protocolErr error

	chain := getChain()
	log.Printf("Running with chain: %s\n", chain)

	ethClient, err := web3.BuildWeb3Client(chain, true)
	ethClientWriter, err2 := web3.BuildWeb3Client(chain, false)

	if err != nil || err2 != nil {
		log.Fatalf("Error building eth cient: %s", err)
	}
	log.Printf("Successfully built eth client on %s", chain)

	walletPrivateKey := config.GetSecret(config.WalletTarotKeyOne)

	if walletPrivateKey == "" {
		log.Fatalf("wallet test private key not found")
	}

	walletPrivateKeyCiph, err := crypto.HexToECDSA(walletPrivateKey)
	if err != nil {
		log.Fatalf("wallet private key error: %s", err)
	}

	switch chain {
	case models.Base:
		protocol, protocolErr = protocolconfig.GetTarotBaseUsdcAero()
	case models.Optimism:
		protocol, protocolErr = protocolconfig.GetTarotOptimismUsdcTarot()
	default:
		log.Fatalf("Unknown chain: %s", chain)
	}

	if protocolErr != nil {
		log.Fatalf("Error getting protocol: %s", err)
	}

	tarot.Run(ethClient, ethClientWriter, protocol, walletPrivateKeyCiph)
}

// getChain retrieves and validates the blockchain chain parameter from command-line flags.
//
// If validation fails at any point, the function logs an error message and
// exits the program with a non-zero status code.
func getChain() models.Chain {
	var chainStr string
	flag.StringVar(&chainStr, "chain", "", "Blockchain to connect to (required)")
	flag.Parse()

	if chainStr == "" {
		log.Fatalln("Error: -chain parameter is required")
	}

	// Convert to uppercase and cast to Chain type
	chain := models.Chain(strings.ToUpper(chainStr))

	if !validChains[chain] {
		log.Fatalf("Error: Invalid chain '%s'\n", chainStr)
	}

	return chain
}
