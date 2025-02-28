package main

import (
	"defibotgo/internal/config"
	"defibotgo/internal/models"
	"defibotgo/internal/protocols/tarot"
	"defibotgo/internal/web3"
	"flag"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"log"
	"math/big"
	"strings"
)

var validChains = map[models.Chain]bool{
	models.Optimism: true,
	models.Base:     true,
}

func main() {
	chain := getChain()
	log.Printf("Running with chain: %s\n", chain)

	ethClient, err := web3.BuildWeb3Client(chain, true)
	ethClientWriter, err2 := web3.BuildWeb3Client(chain, false)

	if err != nil || err2 != nil {
		log.Fatalf("Error building eth cient: %s", err)
	}
	log.Printf("Successfully built eth client on %s", chain)

	walletPrivateKey := config.GetSecret(config.WalletTarotKeyOne)
	senderWallet := config.GetSecret(config.WalletTarotAddressOne)

	if walletPrivateKey == "" {
		log.Fatalf("wallet test private key not found")
	}

	walletPrivateKeyCiph, err := crypto.HexToECDSA(walletPrivateKey)
	if err != nil {
		log.Fatalf("wallet private key error: %s", err)
	}

	var tarotOpA = tarot.TarotOpts{
		Sender:           common.HexToAddress(senderWallet),
		PriorityFee:      big.NewInt(5678),
		BlockRangeFilter: big.NewInt(20),
		ContractLender:   common.HexToAddress("0x80942A0066F72eFfF5900CF80C235dd32549b75d"),
		ContractGauge:    common.HexToAddress("0x73d5C2f4EB0E4EB15B3234f8B880A10c553DA1ea"),
	}

	tarot.Run(ethClient, ethClientWriter, chain, &tarotOpA, walletPrivateKeyCiph)
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
