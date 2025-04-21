package tarot

import (
	"context"
	"crypto/ecdsa"
	"defibotgo/internal/contract_abi"
	"defibotgo/internal/models"
	"defibotgo/internal/services/asyncservices"
	"defibotgo/internal/utils"
	"defibotgo/internal/web3"
	"defibotgo/internal/web3/web3Async"
	"fmt"
	"github.com/dgraph-io/ristretto"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/rs/zerolog/log"
	"math/big"
	"strings"
	"sync"
	"time"
)

type TarotCalculationOpts struct {
	VaultPendingReward models.WeiResult
	BaseFeePerGas      models.WeiResult
	EstimateGasLimit   models.GasLimitResult
	RewardPair         models.WeiResult
	PriorityFee        models.WeiResult

	// Cached fields for direct access (populated after error checking)
	EstimateGasLimitValue   uint64
	VaultPendingRewardValue *big.Int
	BaseFeeValue            *big.Int
	RewardPairValue         *big.Int
	PriorityFeeValue        *big.Int
}

var (
	reinvestFunctionName    = "reinvest"
	zeroValue               = big.NewInt(0)
	transactionsBlockRange  = big.NewInt(50)
	gasLimitExtraPercent    = uint64(30)
	gasLimitUsedExpectedMin = uint64(100000)
)

func Run(rootCtx context.Context, ethClient *ethclient.Client, ethClientWriter *ethclient.Client, tarotOpts *models.TarotOpts, walletPrivateKey *ecdsa.PrivateKey) {
	chainID, err := ethClientWriter.ChainID(rootCtx)
	if err != nil {
		panic(err)
	}

	cache, err := ristretto.NewCache(&ristretto.Config{
		NumCounters: 50,  // 10x the number of items we expect to store
		MaxCost:     320, // Approx. 64 bytes per *big.Int, 5 keys in total
		BufferItems: 64,  // Recommended size for eviction buffer
	})
	if err != nil {
		panic(err)
	}

	// Init the channels needed for computing reward and gas fee
	vaultPendingRewardChan := make(chan models.WeiResult, 1)
	baseFeePerGasChan := make(chan models.WeiResult, 1)
	estimateGasChan := make(chan models.GasLimitResult, 1)
	rewardPairValueChan := make(chan models.WeiResult, 1)
	priorityFeeChan := make(chan models.WeiResult, 1)

	contractGauge, contractGasPriceOracle, callOpts, callMsg := buildOpts(ethClient, tarotOpts)
	log.Info().Str("contract lender", tarotOpts.ContractLender.Hex()).Msg("Calling contract lender")

	for {
		select {
		case <-rootCtx.Done():
			log.Info().Msg("ctx canceled, exiting tarot.Run")
			return
		default:
			// no cancellation signal, proceed
		}

		iterCtx, loopCancel := context.WithTimeout(rootCtx, time.Second*10)
		callOpts.Context = iterCtx

		// Keep the code in a block to avoid overhead from additional function calls (optimizing execution time)
		tarotCalculationOpts := &TarotCalculationOpts{}
		var wg sync.WaitGroup
		wg.Add(5)

		// Call web3 api asynchronously
		go web3Async.EthCallAsync(contractGauge, "earned", callOpts, vaultPendingRewardChan, &wg, tarotOpts.ContractLender)
		go web3Async.GetBaseFeePerGasAsync(ethClient, callOpts.BlockNumber, cache, "1", baseFeePerGasChan, &wg)
		go web3Async.EstimateGasAsync(ethClient, callMsg, cache, "2", estimateGasChan, &wg)
		go web3Async.GetPriorityFeeAsync(ethClient, tarotOpts.Sender, tarotOpts.ContractLender, transactionsBlockRange, callOpts.BlockNumber, cache, "3", priorityFeeChan, &wg)
		go asyncservices.GetPoolPriceAsync(tarotOpts.Chain, cache, "4", rewardPairValueChan, &wg)

		// Wait for goroutines
		wg.Wait()

		// Get the channels result
		tarotCalculationOpts.VaultPendingReward = <-vaultPendingRewardChan
		tarotCalculationOpts.BaseFeePerGas = <-baseFeePerGasChan
		tarotCalculationOpts.EstimateGasLimit = <-estimateGasChan
		tarotCalculationOpts.RewardPair = <-rewardPairValueChan
		tarotCalculationOpts.PriorityFee = <-priorityFeeChan

		if tarotCalculationOpts.VaultPendingReward.Err != nil || tarotCalculationOpts.BaseFeePerGas.Err != nil || tarotCalculationOpts.EstimateGasLimit.Err != nil || tarotCalculationOpts.RewardPair.Err != nil || tarotCalculationOpts.PriorityFee.Err != nil {
			log.Error().
				Str("chain", string(tarotOpts.Chain)).
				AnErr("pendingRewardError", tarotCalculationOpts.VaultPendingReward.Err).
				AnErr("baseFeeError", tarotCalculationOpts.BaseFeePerGas.Err).
				AnErr("gasLimitError", tarotCalculationOpts.EstimateGasLimit.Err).
				AnErr("rewardPairError", tarotCalculationOpts.RewardPair.Err).
				AnErr("priorityFeeError", tarotCalculationOpts.PriorityFee.Err).
				Msg("Failed to calculate transaction parameters")
			loopCancel()
			time.Sleep(utils.RetryErrorSleep)
			continue
		}

		// Set values for direct access to avoid deeply nested references
		tarotCalculationOpts.VaultPendingRewardValue = tarotCalculationOpts.VaultPendingReward.Value
		tarotCalculationOpts.BaseFeeValue = tarotCalculationOpts.BaseFeePerGas.Value
		tarotCalculationOpts.EstimateGasLimitValue = tarotCalculationOpts.EstimateGasLimit.Value
		tarotCalculationOpts.PriorityFeeValue = tarotCalculationOpts.PriorityFee.Value
		tarotCalculationOpts.RewardPairValue = tarotCalculationOpts.RewardPair.Value

		// Add 8 up to 20% of extra priority fees to make it unpredictable
		priorityFeeExtraPercent := utils.RandomNumberInRange(8, 20)
		isL2Worth, l2GasOpts, rewardEth, err := GetL2TransactionGasFees(tarotOpts, tarotCalculationOpts, priorityFeeExtraPercent, gasLimitExtraPercent)
		if err != nil {
			log.Error().Err(err).Str("chain", string(tarotOpts.Chain)).Msg("Error getting gas on Tarot")
			loopCancel()
			time.Sleep(utils.RetryErrorSleep)
			continue
		}

		if !isL2Worth {
			loopCancel()
			time.Sleep(utils.RetryMainSleep)
			continue
		}

		// Estimate L1 gas fee
		isWorth, signedTx, err := getL1TransactionGasFees(iterCtx, ethClient, chainID, callOpts, l2GasOpts, tarotOpts, contractGasPriceOracle, rewardEth, walletPrivateKey)
		loopCancel()
		if err != nil {
			log.Error().Err(err).Str("chain", string(tarotOpts.Chain)).Msg("Error getting l1 gas fee")
			time.Sleep(utils.RetryErrorSleep)
			continue
		}

		// The reward is lower than the transaction fee estimated
		if !isWorth {
			time.Sleep(utils.RetryMainSleep)
			continue
		}

		// Send transaction on chain
		txCtx, txCancel := context.WithTimeout(rootCtx, time.Second*20)
		err = ethClientWriter.SendTransaction(txCtx, signedTx)

		if err != nil {
			log.Error().Err(err).Str("chain", string(tarotOpts.Chain)).Msg("Failed to send transaction on Tarot")
			txCancel()
			time.Sleep(utils.RetryErrorSleep)
			continue
		}

		waitTransaction(ethClient, txCtx, signedTx, tarotOpts.Chain)

		// free resources
		txCancel()
	}
}

func GetL2TransactionGasFees(
	tarotOpts *models.TarotOpts,
	tarotCalculationOpts *TarotCalculationOpts,
	priorityFeeExtraPercent int,
	gasLimitExtraPercent uint64,
) (bool, *web3.GasOpts, *big.Int, error) {
	// Gas limit is too low to be correct
	if tarotCalculationOpts.EstimateGasLimitValue < gasLimitUsedExpectedMin {
		return false, nil, nil, fmt.Errorf("estimated gas limit is too low => skip")
	}

	// Set new priority fee depending on competitors
	if tarotCalculationOpts.PriorityFeeValue == nil {
		tarotCalculationOpts.PriorityFeeValue = zeroValue
		return false, nil, nil, fmt.Errorf("priority fee is set to 0")
	}

	newPriorityFee_ := tarotCalculationOpts.PriorityFeeValue
	if tarotOpts.PriorityFee.Cmp(tarotCalculationOpts.PriorityFeeValue) == 1 {
		// Use the highest priority fee for the transaction
		newPriorityFee_ = tarotOpts.PriorityFee
	}

	newPriorityFee := utils.IncreaseAmount(newPriorityFee_, priorityFeeExtraPercent)
	rewardToken := ComputeReward(tarotCalculationOpts.VaultPendingRewardValue, tarotOpts.ReinvestBounty)
	rewardEth := utils.ConvertToEth(rewardToken, tarotCalculationOpts.RewardPairValue)

	gasOpts := web3.BuildTransactionFeeArgs(tarotCalculationOpts.BaseFeeValue, newPriorityFee, tarotCalculationOpts.EstimateGasLimitValue)
	isWorth := utils.ComputeDifference(rewardEth, gasOpts.TransactionFee) > 0

	log.Info().Str("vault pending reward", tarotCalculationOpts.VaultPendingRewardValue.String()).
		Str("reward erc20", rewardToken.String()).
		Str("reward weth", rewardEth.String()).
		Str("l2 transaction fee", gasOpts.TransactionFee.String()).
		Str("l2 base fee", tarotCalculationOpts.BaseFeeValue.String()).
		Str("max fee", gasOpts.GasFeeCap.String()).
		Str("priority fee", gasOpts.GasTipCap.String()).
		Uint64("gas limit", gasOpts.GasLimit).
		Str("reward pair", tarotCalculationOpts.RewardPairValue.String()).
		Msg("")

	// Increase gas limit to ensure the success of the transaction
	// TODO: use utils.IncreaseAmount
	gasOpts.GasLimit = tarotCalculationOpts.EstimateGasLimitValue + (tarotCalculationOpts.EstimateGasLimitValue*gasLimitExtraPercent)/100

	return isWorth, gasOpts, rewardEth, nil
}

func getL1TransactionGasFees(
	ctx context.Context,
	ethClient *ethclient.Client,
	chainId *big.Int,
	callOpts *bind.CallOpts,
	gasOpts *web3.GasOpts,
	tarotOpts *models.TarotOpts,
	contractGasPriceOracle *bind.BoundContract,
	rewardEth *big.Int,
	walletPrivateKey *ecdsa.PrivateKey,
) (bool, *types.Transaction, error) {
	l1GasFee, signedTx, err := web3.GetL1GasFee(ctx, ethClient, chainId, callOpts, gasOpts, contractGasPriceOracle, &tarotOpts.ContractLender, reinvestFunctionName, walletPrivateKey)
	if err != nil {
		return false, nil, err
	}

	// The gas-price oracle may apply an extra buffer in the final block
	// to account for last‑second basefee volatility.
	scaledL1GasFee := utils.DecreaseAmount(l1GasFee, 10)
	transactionFee := new(big.Int).Add(gasOpts.TransactionFee, scaledL1GasFee)
	diff := utils.ComputeDifference(rewardEth, transactionFee)

	// TODO: set value though params
	isWorth := diff > -6
	log.Info().Str("l1GasFee", l1GasFee.String()).Str("scaledL1GasFee", scaledL1GasFee.String()).Str("transaction fee", transactionFee.String()).Float64("diff", diff).Msg("")

	return isWorth, signedTx, nil
}

func ComputeReward(vaultPendingReward *big.Int, reinvestBounty *big.Int) *big.Int {
	// Calculate bounty = reward * REINVEST_BOUNTY / 1e18
	bounty := new(big.Int).Mul(vaultPendingReward, reinvestBounty) // reward * reinvestBounty
	bounty.Div(bounty, utils.OneE18)                               // divide by 1e18

	return bounty
}

// buildOpts initializes and returns the contract bindings and call options
// for interacting with the Tarot Gauge and Gas Price Oracle contracts.
// Note: The returned CallOpts.Context must be set manually by the caller
//
//	(e.g., using context.WithTimeout or context.WithCancel) before use.
func buildOpts(ethClient *ethclient.Client, tarotOpts *models.TarotOpts) (*bind.BoundContract, *bind.BoundContract, *bind.CallOpts, ethereum.CallMsg) {
	contractGauge, err := web3.BuildContractInstance(ethClient, tarotOpts.ContractGauge, contract_abi.CONTRACT_ABI_GAUGE)
	if err != nil {
		log.Fatal().Err(err).Str("gauge contract", tarotOpts.ContractGauge.String()).Msg("Error building tarot contract gauge instance")
	}

	contractGasPriceOracle, err := web3.BuildContractInstance(ethClient, tarotOpts.ContractGasPriceOracle, contract_abi.CONTRACT_ABI_GAS_PRICE_ORACLE)
	if err != nil {
		log.Fatal().Err(err).Str("gauge contract", tarotOpts.ContractGasPriceOracle.String()).Msg("Error building tarot contract L1 Block instance")
	}

	abiJson, err := web3.LoadAbi(contract_abi.CONTRACT_ABI_LENDER)
	if err != nil {
		log.Fatal().Err(err).Str("chain", string(tarotOpts.Chain)).Msg("Error loading Tarot contract_abi")
	}

	data, err := abiJson.Pack(reinvestFunctionName)
	if err != nil {
		log.Fatal().Err(err).Str("chain", string(tarotOpts.Chain)).Msg("Failed to pack Tarot contract_abi")
	}

	// TODO manage cancel for this context ? Set new global context ?
	callOpts := &bind.CallOpts{
		Pending:     true,
		BlockNumber: nil,
		// the context must set manually after calling this function
	}

	// Create a message to simulate the transaction
	callMsg := ethereum.CallMsg{
		From:  tarotOpts.Sender,
		To:    &tarotOpts.ContractLender,
		Data:  data, // ABI-encoded function call data
		Value: zeroValue,
	}

	return contractGauge, contractGasPriceOracle, callOpts, callMsg
}

func waitTransaction(ethClient *ethclient.Client, ctx context.Context, tx *types.Transaction, chain models.Chain) {
	log.Info().Str("hash", tx.Hash().Hex()).Msg("Sent transaction on Tarot")

	// Wait for the transaction's validation
	receipt, err := bind.WaitMined(ctx, ethClient, tx)

	if err != nil {
		log.Error().Err(err).Str("chain", string(chain)).Msg("Failed to wait for receipt on Tarot")

		if strings.Contains(err.Error(), "context deadline exceeded") {
			log.Error().Msgf("Wait for %v", utils.RetryExpiredContextSleep)
			time.Sleep(utils.RetryExpiredContextSleep)
			return
		}

		if strings.Contains(err.Error(), "replacement transaction underpriced") {
			//TODO: send eth to the wallet with higher priority fee.
		}

		return
	}

	if receipt.Status == types.ReceiptStatusSuccessful {
		log.Info().Str("hash", tx.Hash().Hex()).Msg("Successfully sent transaction on Tarot")
		time.Sleep(utils.RetrySuccessSleep)
	} else {
		log.Error().Err(err).Str("chain", string(chain)).Msg("Failed to send transaction on Tarot")
		time.Sleep(utils.RetryErrorSleep)
	}
}
