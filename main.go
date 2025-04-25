package main

import (
	"context"
	"defibotgo/internal/config"
	"defibotgo/internal/models"
	protocolconfig "defibotgo/internal/protocols/config"
	"defibotgo/internal/protocols/tarot"
	"defibotgo/internal/web3"
	"flag"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

var validChains = map[models.Chain]bool{
	models.Optimism: true,
	models.Base:     true,
}

func main() {
	zerolog.TimeFieldFormat = time.DateTime

	var protocol *models.TarotOpts
	var protocolErr error

	chain := getChain()
	log.Info().Str("chain", string(chain)).Msg("Running with chain")

	ethClient, err := web3.BuildWeb3Client(chain, true)
	ethClientWriter, err2 := web3.BuildWeb3Client(chain, false)

	if err != nil || err2 != nil {
		log.Fatal().Err(err).Msg("Error building eth client")
	}
	log.Info().Str("chain", string(chain)).Msg("Successfully built eth client")

	walletPrivateKey := config.GetSecret(config.WalletTarotKeyOne)

	if walletPrivateKey == "" {
		log.Fatal().Msg("wallet test private key not found")
	}

	walletPrivateKeyCiph, err := crypto.HexToECDSA(walletPrivateKey)
	if err != nil {
		log.Fatal().Err(err).Msg("wallet private key error")
	}

	switch chain {
	case models.Base:
		protocol, protocolErr = protocolconfig.GetTarotBaseUsdcAero()
	case models.Optimism:
		protocol, protocolErr = protocolconfig.GetTarotOptimismUsdcTarot()
	default:
		log.Fatal().Str("chain", string(chain)).Msg("Unknown chain")
	}

	if protocolErr != nil {
		log.Fatal().Err(err).Msg("Error getting protocol")
	}

	rootCtx, rootCancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer rootCancel()
	tarot.Run(rootCtx, ethClient, ethClientWriter, protocol, walletPrivateKeyCiph)
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
		log.Fatal().Msg("Error: -chain parameter is required")
	}

	// Convert to uppercase and cast to Chain type
	chain := models.Chain(strings.ToUpper(chainStr))

	if !validChains[chain] {
		log.Fatal().Str("chain", chainStr).Msg("Error: Invalid chain")
	}

	return chain
}
