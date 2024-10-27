package main

import (
	"defibotgo/internal/models"
	"defibotgo/internal/protocols/tarot"
	"defibotgo/internal/web3"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/joho/godotenv"
	"log"
	"math/big"
	"os"
)

func main() {
	chain := models.Optimism
	ethClient, err := web3.BuildWeb3Client(chain, true)
	ethClientWriter, err2 := web3.BuildWeb3Client(chain, false)

	if err != nil || err2 != nil {
		log.Fatalf("Error building eth cient: %s", err)
	}
	log.Printf("Successfully built eth client on %s %p", chain, ethClient)

	err = godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	walletPrivateKey := os.Getenv("ACCOUNT_PRIVATE_KEY_TAROT")
	senderWallet := os.Getenv("ACCOUNT_SENDER_ADDRESS_TAROT")

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
