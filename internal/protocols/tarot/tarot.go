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

type ProtocolCalculationOpts struct {
	// Hot/Cached fields
	VaultPendingRewardValue *big.Int //  8 bytes
	BaseFeeValue            *big.Int //  8 bytes
	RewardPairValue         *big.Int //  8 bytes
	PriorityFeeValue        *big.Int //  8 bytes
	EstimateGasLimitValue   uint64   //  8 bytes

	// “Cold” result structs
	VaultPendingReward models.WeiResult      // 24 bytes
	BaseFeePerGas      models.WeiResult      // 24 bytes
	RewardPair         models.WeiResult      // 24 bytes
	PriorityFee        models.WeiResult      // 24 bytes
	EstimateGasLimit   models.GasLimitResult // 24 bytes
}

var (
	reinvestFunctionName    = "reinvest"
	zeroValue               = big.NewInt(0)
	transactionsBlockRange  = big.NewInt(50)
	gasLimitExtraPercent    = uint64(30)
	gasLimitUsedExpectedMin = uint64(100000)
	blockTime               = int64(2)
)

func Run(rootCtx context.Context, ethClient *ethclient.Client, ethClientWriter *ethclient.Client, tarotOpts *models.TarotOpts, walletPrivateKey *ecdsa.PrivateKey) {
	chainID, err := ethClientWriter.ChainID(rootCtx)
	if err != nil {
		panic(err)
	}

	cache, err := ristretto.NewCache(&ristretto.Config{
		NumCounters: 100, // ~16× counters to minimize collisions and maximize hit rate
		MaxCost:     768, // 6 keys × 128 bytes each (generous overhead to avoid evictions)
		BufferItems: 64,  // Recommended default for smooth eviction buffering
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

	balanceChan := make(chan models.WeiResult, 1)
	totalSupplyChan := make(chan models.WeiResult, 1)

	rateRewardChan := make(chan *big.Int, 1)

	contractGauge, contractGasPriceOracle, callOpts, callMsg, lenderCallData := buildOpts(ethClient, tarotOpts)
	minExtraPriorityFeePercent, maxExtraPriorityFeePercent := tarotOpts.ExtraPriorityFeePercent[0], tarotOpts.ExtraPriorityFeePercent[1]
	rewardRate := tarotOpts.RewardRate

	// get the reward rate every 10 minutes
	rateRewardCallOpts := &bind.CallOpts{
		Pending:     callOpts.Pending,
		BlockNumber: callOpts.BlockNumber,
		From:        callOpts.From,
		Context:     rootCtx,
	}
	go startRateRewardFetcher(rootCtx, contractGauge, "rewardRate", rateRewardCallOpts, 10*time.Minute, rateRewardChan)

	for {
		select {
		case <-rootCtx.Done():
			log.Info().Msg("ctx canceled, exiting tarot.Run")
			return
		case rr := <-rateRewardChan:
			rewardRate = rr
			log.Debug().
				Str("chain", string(tarotOpts.Chain)).
				Str("newRateReward", rr.String()).
				Msg("updated rewardRate")
		default:
			// no cancellation signal, proceed
		}

		iterCtx, iterCancelCtx := context.WithTimeout(rootCtx, time.Second*10)
		callOpts.Context = iterCtx

		// Keep the code in a block to avoid overhead from additional function calls (optimizing execution time)
		tarotCalculationOpts := &ProtocolCalculationOpts{}
		var wg sync.WaitGroup
		wg.Add(7)

		// Call web3 api asynchronously
		go web3Async.EthCallAsync(contractGauge, "earned", callOpts, vaultPendingRewardChan, &wg, tarotOpts.ContractLender)
		go web3Async.GetBaseFeePerGasAsync(ethClient, callOpts.BlockNumber, cache, "1", baseFeePerGasChan, &wg)
		go web3Async.EstimateGasAsync(ethClient, callMsg, cache, "2", estimateGasChan, &wg)
		go web3Async.GetPriorityFeeAsync(ethClient, tarotOpts.Sender, tarotOpts.ContractLender, transactionsBlockRange, callOpts.BlockNumber, cache, "3", priorityFeeChan, &wg)
		go asyncservices.GetPoolPriceAsync(tarotOpts.Chain, cache, "4", rewardPairValueChan, &wg)

		go web3Async.EthCallWithCacheAsync(contractGauge, "balanceOf", callOpts, cache, "5", balanceChan, &wg, tarotOpts.ContractLender)
		go web3Async.EthCallWithCacheAsync(contractGauge, "totalSupply", callOpts, cache, "6", totalSupplyChan, &wg)

		// Wait for goroutines
		wg.Wait()

		// Get the channels result
		tarotCalculationOpts.VaultPendingReward = <-vaultPendingRewardChan
		tarotCalculationOpts.BaseFeePerGas = <-baseFeePerGasChan
		tarotCalculationOpts.EstimateGasLimit = <-estimateGasChan
		tarotCalculationOpts.RewardPair = <-rewardPairValueChan
		tarotCalculationOpts.PriorityFee = <-priorityFeeChan

		gaugeBalance := <-balanceChan
		gaugeTotalSupply := <-totalSupplyChan

		if tarotCalculationOpts.VaultPendingReward.Err != nil || tarotCalculationOpts.BaseFeePerGas.Err != nil || tarotCalculationOpts.EstimateGasLimit.Err != nil || tarotCalculationOpts.RewardPair.Err != nil || tarotCalculationOpts.PriorityFee.Err != nil || gaugeBalance.Err != nil || gaugeTotalSupply.Err != nil {
			log.Error().
				Str("chain", string(tarotOpts.Chain)).
				AnErr("pendingRewardError", tarotCalculationOpts.VaultPendingReward.Err).
				AnErr("baseFeeError", tarotCalculationOpts.BaseFeePerGas.Err).
				AnErr("gasLimitError", tarotCalculationOpts.EstimateGasLimit.Err).
				AnErr("rewardPairError", tarotCalculationOpts.RewardPair.Err).
				AnErr("priorityFeeError", tarotCalculationOpts.PriorityFee.Err).
				AnErr("balanceError", gaugeBalance.Err).
				AnErr("totalSupplyError", gaugeTotalSupply.Err).
				Msg("Failed to calculate transaction parameters")
			iterCancelCtx()
			time.Sleep(utils.RetryErrorSleep)
			continue
		}

		// Set values for direct access to avoid deeply nested references
		tarotCalculationOpts.VaultPendingRewardValue = GetVaultPendingReward(tarotCalculationOpts.VaultPendingReward.Value, rewardRate, blockTime, gaugeBalance.Value, gaugeTotalSupply.Value)
		tarotCalculationOpts.BaseFeeValue = tarotCalculationOpts.BaseFeePerGas.Value
		tarotCalculationOpts.EstimateGasLimitValue = tarotCalculationOpts.EstimateGasLimit.Value
		tarotCalculationOpts.PriorityFeeValue = tarotCalculationOpts.PriorityFee.Value
		tarotCalculationOpts.RewardPairValue = tarotCalculationOpts.RewardPair.Value

		// Add extra priority fees to make it unpredictable
		priorityFeeExtraPercent := utils.RandomNumberInRange(minExtraPriorityFeePercent, maxExtraPriorityFeePercent)
		isL2Worth, l2GasOpts, rewardEth, err := GetL2TransactionGasFees(tarotOpts, tarotCalculationOpts, priorityFeeExtraPercent, gasLimitExtraPercent)
		if err != nil {
			log.Error().Err(err).Str("chain", string(tarotOpts.Chain)).Msg("Error getting gas on Tarot")
			iterCancelCtx()
			time.Sleep(utils.RetryErrorSleep)
			continue
		}

		if !isL2Worth {
			iterCancelCtx()
			time.Sleep(utils.RetryMainSleep)
			continue
		}

		// Estimate L1 gas fee
		isWorth, signedTx, err := getL1TransactionGasFees(iterCtx, ethClient, chainID, callOpts, l2GasOpts, tarotOpts, contractGasPriceOracle, lenderCallData, rewardEth, walletPrivateKey)
		iterCancelCtx()
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
		txCtx, txCancelCtx := context.WithTimeout(rootCtx, time.Second*20)
		err = ethClientWriter.SendTransaction(txCtx, signedTx)

		if err != nil {
			log.Error().Err(err).Str("chain", string(tarotOpts.Chain)).Msg("Failed to send transaction on Tarot")
			txCancelCtx()
			time.Sleep(utils.RetryErrorSleep)
			continue
		}

		waitTransaction(ethClient, txCtx, signedTx, tarotOpts.Chain)

		// free resources
		txCancelCtx()
	}
}

func GetL2TransactionGasFees(
	tarotOpts *models.TarotOpts,
	tarotCalculationOpts *ProtocolCalculationOpts,
	priorityFeeExtraPercent int,
	gasLimitExtraPercent uint64,
) (bool, *web3.GasOpts, *big.Int, error) {
	// Gas limit is too low to be correct
	if tarotCalculationOpts.EstimateGasLimitValue < gasLimitUsedExpectedMin {
		log.Debug().Msgf("Update gas used from %v to %v", tarotCalculationOpts.EstimateGasLimitValue, tarotOpts.GasUsedDefault)
		tarotCalculationOpts.EstimateGasLimitValue = tarotOpts.GasUsedDefault
	}

	//Set new priority fee depending on competitors
	if tarotCalculationOpts.PriorityFeeValue == nil {
		tarotCalculationOpts.PriorityFeeValue = zeroValue
		return false, nil, nil, fmt.Errorf("priority fee is set to 0")
	}

	//newPriorityFee := tarotOpts.PriorityFee
	newPriorityFee_ := tarotCalculationOpts.PriorityFeeValue
	if tarotOpts.PriorityFee.Cmp(tarotCalculationOpts.PriorityFeeValue) == 1 {
		// Use the highest priority fee for the transaction
		newPriorityFee_ = tarotOpts.PriorityFee
	}

	newPriorityFee := utils.IncreaseAmount(newPriorityFee_, priorityFeeExtraPercent)
	rewardToken := ComputeReward(tarotCalculationOpts.VaultPendingRewardValue, tarotOpts.ReinvestBounty)
	rewardEth := utils.ConvertToEth(rewardToken, tarotCalculationOpts.RewardPairValue)

	gasOpts := web3.BuildTransactionFeeArgs(tarotCalculationOpts.BaseFeeValue, newPriorityFee, tarotCalculationOpts.EstimateGasLimitValue)
	diff := utils.ComputeDifference(rewardEth, gasOpts.TransactionFee)
	isWorth := diff > -10

	log.Info().Str("vault pending reward", tarotCalculationOpts.VaultPendingRewardValue.String()).
		Str("reward erc20", rewardToken.String()).
		Str("reward weth", rewardEth.String()).
		Str("l2 transaction fee", gasOpts.TransactionFee.String()).
		Str("l2 base fee", tarotCalculationOpts.BaseFeeValue.String()).
		Str("max fee", gasOpts.GasFeeCap.String()).
		Str("priority fee", gasOpts.GasTipCap.String()).
		Uint64("gas limit", gasOpts.GasLimit).
		Str("reward pair", tarotCalculationOpts.RewardPairValue.String()).
		Float64("l2 diff", diff).
		Msg("")

	// Increase gas limit to ensure the success of the transaction
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
	toContractCallData []byte,
	rewardEth *big.Int,
	walletPrivateKey *ecdsa.PrivateKey,
) (bool, *types.Transaction, error) {
	l1GasFee, signedTx, err := web3.GetL1GasFee(ctx, ethClient, chainId, callOpts, gasOpts, contractGasPriceOracle, &tarotOpts.ContractLender, toContractCallData, walletPrivateKey)
	if err != nil {
		return false, nil, err
	}

	// The gas-price oracle may apply an extra buffer in the final block
	// to account for last‑second basefee volatility.
	scaledL1GasFee := utils.DecreaseAmount(l1GasFee, 12)
	transactionFee := new(big.Int).Add(gasOpts.TransactionFee, scaledL1GasFee)
	diff := utils.ComputeDifference(rewardEth, transactionFee)

	isWorth := diff > tarotOpts.ProfitableThreshold
	log.Info().Str("l1GasFee", l1GasFee.String()).Str("scaledL1GasFee", scaledL1GasFee.String()).Str("transaction fee", transactionFee.String()).Float64("l1 diff", diff).Msg("")

	return isWorth, signedTx, nil
}

func ComputeReward(vaultPendingReward *big.Int, reinvestBounty *big.Int) *big.Int {
	// Calculate bounty = reward * REINVEST_BOUNTY / 1e18
	bounty := new(big.Int).Mul(vaultPendingReward, reinvestBounty) // reward * reinvestBounty
	bounty.Div(bounty, utils.OneE18)                               // divide by 1e18

	return bounty
}

// GetVaultPendingReward predicts the next earned reward based on last mined block values
func GetVaultPendingReward(
	lastEarned *big.Int,   // earned(account) from last mined block
	rewardRate *big.Int,   // tokens emitted per second
	expectedSeconds int64, // estimated seconds until your tx is mined
	balanceOf *big.Int,    // contract LP token balance
	totalSupply *big.Int,  // total LP supply in gauge
) *big.Int {
	log.Debug().Str("earned", lastEarned.String()).Str("rewardRate", rewardRate.String()).Str("totalSupply", totalSupply.String()).Str("balanceOf", balanceOf.String()).Int64("seconds", expectedSeconds).Msg("")
	rewardRateTimesSeconds := new(big.Int).Mul(rewardRate, big.NewInt(expectedSeconds))
	userRewardPortion := new(big.Int).Mul(rewardRateTimesSeconds, balanceOf)
	additionalReward := new(big.Int).Div(userRewardPortion, totalSupply)
	estimateReward := new(big.Int).Add(lastEarned, additionalReward)

	return estimateReward
}

// buildOpts initializes and returns the contract bindings and call options
// for interacting with the Tarot Gauge and Gas Price Oracle contracts.
// Note: The returned CallOpts.Context must be set manually by the caller
//
//	(e.g., using context.WithTimeout or context.WithCancel) before use.
func buildOpts(ethClient *ethclient.Client, tarotOpts *models.TarotOpts) (*bind.BoundContract, *bind.BoundContract, *bind.CallOpts, ethereum.CallMsg, []byte) {
	contractGauge, err := web3.BuildContractInstance(ethClient, tarotOpts.ContractGauge, contract_abi.CONTRACT_ABI_GAUGE)
	if err != nil {
		log.Fatal().Err(err).Str("gauge contract", tarotOpts.ContractGauge.String()).Msg("Error building tarot contract gauge instance")
	}

	contractGasPriceOracle, err := web3.BuildContractInstance(ethClient, tarotOpts.ContractGasPriceOracle, contract_abi.CONTRACT_ABI_GAS_PRICE_ORACLE)
	if err != nil {
		log.Fatal().Err(err).Str("gauge contract", tarotOpts.ContractGasPriceOracle.String()).Msg("Error building tarot contract L1 Block instance")
	}

	lenderAbiJson, err := web3.LoadAbi(contract_abi.CONTRACT_ABI_LENDER)
	if err != nil {
		log.Fatal().Err(err).Str("chain", string(tarotOpts.Chain)).Msg("Error loading Tarot contract_abi")
	}

	lenderData, err := lenderAbiJson.Pack(reinvestFunctionName)
	if err != nil {
		log.Fatal().Err(err).Str("chain", string(tarotOpts.Chain)).Msg("Failed to pack Tarot contract_abi")
	}

	callOpts := &bind.CallOpts{
		Pending:     false, // last block mined
		BlockNumber: nil,
		From:        tarotOpts.Sender,
		// the context must set manually after calling this function
	}

	// Create a message to simulate the transaction
	callMsg := ethereum.CallMsg{
		From:  tarotOpts.Sender,
		To:    &tarotOpts.ContractLender,
		Data:  lenderData, // ABI-encoded function call lenderData
		Value: zeroValue,
	}

	return contractGauge, contractGasPriceOracle, callOpts, callMsg, lenderData
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

// startRateRewardFetcher blocks, polling rateReward every interval
func startRateRewardFetcher(
	ctx context.Context,
	contract *bind.BoundContract,
	contractFunction string,
	callOpts *bind.CallOpts,
	interval time.Duration,
	rateCh chan<- *big.Int,
) {
	// get rate reward now
	if rate, err := web3.EthCall(contract, contractFunction, callOpts); err == nil {
		select {
		case rateCh <- rate:
		default:
		}
	} else {
		log.Error().Err(err).Msg("initial rateReward fetch failed")
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			rate, err := web3.EthCall(contract, contractFunction, callOpts)
			if err != nil {
				log.Error().Err(err).Msg("failed to fetch rateReward")
				continue
			}
			select {
			case rateCh <- rate:
			default:
			}
		}
	}
}
