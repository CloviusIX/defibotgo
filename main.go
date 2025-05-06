package main

import (
	"context"
	"defibotgo/internal/config"
	"defibotgo/internal/models"
	protocolconfig "defibotgo/internal/protocols/config"
	"defibotgo/internal/protocols/tarot"
	"defibotgo/internal/web3"
	"flag"
	"fmt"
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

var validProtocols = map[models.Protocol]bool{
	models.Tarot:    true,
	models.Impermax: true,
}

var validPools = map[models.Pool]bool{
	models.FbombCbbtc: true,
	models.UsdcAero:   true,
	models.WethTarot:  true,
}

var poolRegistry = map[string]map[string]map[string]models.TarotOpts{
	string(models.Base): {
		string(models.Tarot): {
			string(models.UsdcAero):  protocolconfig.TarotBaseUsdcAero,
			string(models.WethTarot): protocolconfig.TarotBaseWethTarot,
		},
		string(models.Impermax): {
			string(models.FbombCbbtc): protocolconfig.ImpermaxBaseSTKDUNIV2,
		},
	},
}

var walletRegistry = map[string]string{
	strings.ToUpper(config.GetSecret(config.WalletTarotAddressOne)):    config.GetSecret(config.WalletTarotKeyOne),
	strings.ToUpper(config.GetSecret(config.WalletTarotAddressTwo)):    config.GetSecret(config.WalletTarotKeyTwo),
	strings.ToUpper(config.GetSecret(config.WalletImpermaxAddressOne)): config.GetSecret(config.WalletImpermaxKeyOne),
}

func main() {
	zerolog.TimeFieldFormat = time.DateTime
	rootCtx, rootCancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer rootCancel()

	// Get command args to build the pool opts
	chain, protocol, poolID := getCmdArgs()

	ethClient, err := web3.BuildWeb3Client(chain, true)
	ethClientWriter, err2 := web3.BuildWeb3Client(chain, false)

	if err != nil || err2 != nil {
		log.Fatal().Err(err).Msg("Error building eth client")
	}

	poolOpts, poolErr := getPoolOpts(chain, protocol, poolID)
	if poolErr != nil {
		log.Fatal().Err(poolErr).Msg("Error getting pool")
	}

	if poolOpts.Sender == protocolconfig.ZeroAddress {
		log.Fatal().Msg("pool options sender is null")
	}

	senderAddress := strings.ToUpper(poolOpts.Sender.Hex())
	walletPrivateKey := walletRegistry[senderAddress]

	if walletPrivateKey == "" {
		log.Fatal().Msg("wallet test private key not found")
	}

	walletPrivateKeyCiph, err := crypto.HexToECDSA(walletPrivateKey)
	if err != nil {
		log.Fatal().Err(err).Msg("wallet private key error")
	}

	blockNumber, err := ethClient.BlockNumber(rootCtx)
	if err != nil {
		log.Fatal().Err(err).Msg("Error getting block number")
	}

	startBlock := blockNumber
	log.Debug().Msgf("Start at %v", blockNumber)

	for {
		currentBlock, err := ethClient.BlockNumber(rootCtx)
		if err != nil {
			log.Error().Err(err).Msg("Error getting block number in loop")
			continue
		}

		if currentBlock > startBlock {
			blockNumber = currentBlock
			break
		}
		time.Sleep(100 * time.Millisecond)
	}

	log.Info().Uint64("block number", blockNumber).Str("wallet address", senderAddress).Str("chain", string(chain)).Msgf("Running on %s on %s %s", string(protocol), string(chain), string(poolID))
	tarot.Run(rootCtx, ethClient, ethClientWriter, &poolOpts, walletPrivateKeyCiph)
}

// getCmdArgs parses and validates command-line arguments for chain, protocol, and pool.
//
// It defines the expected flags (-chain, -protocol, -pool), parses the input once,
// and validates each argument against a predefined set of allowed values.
//
// If any argument is missing or invalid, the function logs a fatal error and exits the program.
//
// Returns:
// - models.Chain: the validated blockchain chain
// - models.Protocol: the validated DeFi protocol
// - models.Pool: the validated pool identifier
func getCmdArgs() (models.Chain, models.Protocol, models.Pool) {
	var chainStr string
	var protocolStr string
	var poolStr string

	flag.StringVar(&chainStr, "chain", "", "Blockchain to connect to (required)")
	flag.StringVar(&protocolStr, "protocol", "", "Protocol to connect to (required)")
	flag.StringVar(&poolStr, "pool", "", "Pool to connect to (required)")

	flag.Parse()

	chain := validateArg[models.Chain](chainStr, "chain", validChains)
	protocol := validateArg[models.Protocol](protocolStr, "protocol", validProtocols)
	poolID := validateArg[models.Pool](poolStr, "pool", validPools)

	return chain, protocol, poolID
}

// validateArg parses and validates a command-line flag input against a set of allowed values.
//
// T must be a string-like type (e.g., models.Chain, models.Protocol, models.Pool).
//
// Parameters:
// - valueStr: the raw string input from the command line
// - name: the human-readable name of the flag (for error messages)
// - validValues: a map containing the set of allowed values for this type
//
// If validation fails, the function logs a fatal error and exits the program.
// Otherwise, it returns the validated and correctly typed value.
func validateArg[T ~string](valueStr, name string, validValues map[T]bool) T {
	if valueStr == "" {
		log.Fatal().Msgf("Error: -%s parameter is required", name)
	}

	value := T(strings.ToUpper(valueStr))

	if !validValues[value] {
		log.Fatal().Str(name, valueStr).Msgf("Error: Invalid %s", name)
	}

	return value
}

// getPoolOpts retrieves the configuration options for a specific pool based on the given chain, protocol, and pool ID.
//
// Parameters:
// - chain: the blockchain network (e.g., Base, Optimism)
// - protocol: the DeFi protocol (e.g., Tarot, Impermax)
// - poolID: the pool identifier
//
// Returns:
// - models.TarotOpts: the configuration options for the selected pool
// - error: non-nil if the pool was not found
func getPoolOpts(chain models.Chain, protocol models.Protocol, poolID models.Pool) (models.TarotOpts, error) {
	protocols, ok := poolRegistry[string(chain)]
	if !ok {
		return models.TarotOpts{}, fmt.Errorf("no protocols found for chain=%s", chain)
	}

	pools, ok := protocols[string(protocol)]
	if !ok {
		return models.TarotOpts{}, fmt.Errorf("no pools found for chain=%s protocol=%s", chain, protocol)
	}

	pool, ok := pools[string(poolID)]
	if !ok {
		return models.TarotOpts{}, fmt.Errorf("pool not found for chain=%s protocol=%s poolID=%s", chain, protocol, poolID)
	}

	return pool, nil
}
