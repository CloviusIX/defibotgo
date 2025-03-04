package tarot

import (
	"context"
	"crypto/ecdsa"
	"defibotgo/internal/abi"
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
	"log"
	"math/big"
	"strings"
	"sync"
	"time"
)

type WeiApiFunc func() (*big.Int, error)

type TarotCalculationOpts struct {
	VaultPendingReward models.WeiResult
	BaseFeePerGas      models.WeiResult
	EstimateGasLimit   models.GasLimitResult
	RewardPair         models.WeiResult
	PriorityFee        models.WeiResult

	// Cached fields for direct access (populated after error checking)
	VaultPendingRewardValue *big.Int
	BaseFeeValue            *big.Int
	EstimateGasLimitValue   uint64
	RewardPairValue         *big.Int
	PriorityFeeValue        *big.Int
}

var (
	reinvestFunctionName    = "reinvest"
	zeroValue               = big.NewInt(0)
	transactionsBlockRange  = big.NewInt(50)
	gasLimitExtraPercent    = uint64(30)
	gasLimitUsedExpectedMin = uint64(100000)
	reinvestFee, _          = new(big.Int).SetString("20000000000000000", 10) // 0.02 * 1e18 (which is 2e16)
)

func Run(ethClient *ethclient.Client, ethClientWriter *ethclient.Client, tarotOtps *models.TarotOpts, walletPrivateKey *ecdsa.PrivateKey) {
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

	contractLender, err := web3.BuildContractInstance(ethClient, tarotOtps.ContractLender, abi.CONTRACT_ABI_LENDER)
	if err != nil {
		log.Fatalf("error building tarot contract lender instance on %s: %v", tarotOtps.Chain, err)
	}

	contractGauge, err := web3.BuildContractInstance(ethClient, tarotOtps.ContractGauge, abi.CONTRACT_ABI_GAUGE)
	if err != nil {
		log.Fatalf("error building tarot contract gauge instance on %s: %v", tarotOtps.ContractGauge, err)
	}

	abiJson, err := web3.LoadAbi(abi.CONTRACT_ABI_LENDER)
	if err != nil {
		log.Fatalf("error loading Tarot abi on %s: %v", tarotOtps.Chain, err)
	}

	data, err := abiJson.Pack(reinvestFunctionName)
	if err != nil {
		log.Fatalf("failed to pack Tarot abi on %s: %v", tarotOtps.Chain, err)
	}

	callOpts := &bind.CallOpts{
		Pending:     true,
		BlockNumber: nil,
		Context:     context.Background(),
	}

	// Create a message to simulate the transaction
	callMsg := ethereum.CallMsg{
		From:  tarotOtps.Sender,
		To:    &tarotOtps.ContractLender,
		Data:  data, // ABI-encoded function call data
		Value: zeroValue,
	}
	log.Printf("calling contract lender on %s", tarotOtps.ContractLender.Hex())

	for {
		// Keep the code in a block to avoid overhead from additional function calls (optimizing execution time)
		tarotCalculationOpts := &TarotCalculationOpts{}
		var wg sync.WaitGroup
		wg.Add(5)

		// Call web3 api asynchronously
		go web3Async.EthCallAsync(contractGauge, "earned", callOpts, vaultPendingRewardChan, &wg, tarotOtps.ContractLender)
		go web3Async.GetBaseFeePerGasAsync(ethClient, callOpts.BlockNumber, cache, "1", baseFeePerGasChan, &wg)
		go web3Async.EstimateGasAsync(ethClient, callMsg, cache, "2", estimateGasChan, &wg)
		go web3Async.GetPriorityFeeAsync(ethClient, tarotOtps.Sender, tarotOtps.ContractLender, transactionsBlockRange, callOpts.BlockNumber, cache, "3", priorityFeeChan, &wg)
		go asyncservices.GetPoolPriceAsync(tarotOtps.Chain, cache, "4", rewardPairValueChan, &wg)

		// Wait for goroutines
		wg.Wait()

		// Get the channels result
		tarotCalculationOpts.VaultPendingReward = <-vaultPendingRewardChan
		tarotCalculationOpts.BaseFeePerGas = <-baseFeePerGasChan
		tarotCalculationOpts.EstimateGasLimit = <-estimateGasChan
		tarotCalculationOpts.RewardPair = <-rewardPairValueChan
		tarotCalculationOpts.PriorityFee = <-priorityFeeChan

		if tarotCalculationOpts.VaultPendingReward.Err != nil || tarotCalculationOpts.BaseFeePerGas.Err != nil || tarotCalculationOpts.EstimateGasLimit.Err != nil || tarotCalculationOpts.RewardPair.Err != nil || tarotCalculationOpts.PriorityFee.Err != nil {
			log.Printf("error getting gas limit or reward pair value\nPending reward: %s\n:Base Fee: %s\nEstimate gas: %s\nReward Pair value: %s\nPriority fee: %s", tarotCalculationOpts.VaultPendingReward.Err, tarotCalculationOpts.BaseFeePerGas.Err, tarotCalculationOpts.EstimateGasLimit.Err, tarotCalculationOpts.RewardPair.Err, tarotCalculationOpts.PriorityFee.Err)
		}

		tarotCalculationOpts.VaultPendingRewardValue = tarotCalculationOpts.VaultPendingReward.Value
		tarotCalculationOpts.BaseFeeValue = tarotCalculationOpts.BaseFeePerGas.Value
		tarotCalculationOpts.EstimateGasLimitValue = tarotCalculationOpts.EstimateGasLimit.Value
		tarotCalculationOpts.PriorityFeeValue = tarotCalculationOpts.PriorityFee.Value
		tarotCalculationOpts.RewardPairValue = tarotCalculationOpts.RewardPair.Value

		priorityFeeExtraPercent := utils.RandomNumberInRange(20, 25)
		isWorth, gasOpts, err := GetTransactionGasFees(tarotOtps, tarotCalculationOpts, priorityFeeExtraPercent, gasLimitExtraPercent)
		if err != nil {
			log.Printf("error getting gas on Tarot %s: %v", tarotOtps.Chain, err)
			time.Sleep(utils.RetryErrorSleep)
			continue
		}

		if isWorth {
			tx, err := web3.SendTransaction(ethClientWriter, contractLender, reinvestFunctionName, gasOpts, walletPrivateKey)
			if err != nil {
				log.Printf("failed to send transaction on Tarot %s: %v", tarotOtps.Chain, err)
				time.Sleep(utils.RetryErrorSleep)

				if strings.Contains(err.Error(), "context deadline exceeded") {
					time.Sleep(utils.RetryExpiredContextSleep)
					continue
				}

				if strings.Contains(err.Error(), "replacement transaction underpriced") {
					//TODO: send eth to the wallet with higher priority fee.
				}

				time.Sleep(utils.RetryErrorSleep)
				continue
			}

			log.Println("Sent transaction on Tarot ", tx.Hash().Hex())

			ctx, cancel := context.WithTimeout(context.Background(), time.Second*20)
			receipt, err := bind.WaitMined(ctx, ethClient, tx)
			cancel() // Immediately call cancel to free up resources

			if err != nil {
				log.Printf("failed to wait for receipt on Tarot %s: %v", tarotOtps.Chain, err)
				continue
			}

			if receipt.Status == types.ReceiptStatusSuccessful {
				log.Println("Successfully sent transaction on Tarot ", tx.Hash().Hex())
			} else {
				log.Printf("failed to send transaction on Tarot %s: %v", tarotOtps.Chain, err)
			}
		}
		time.Sleep(utils.RetryMainSleep)
	}
}

func GetTransactionGasFees(
	tarotOpts *models.TarotOpts,
	tarotCalculationOpts *TarotCalculationOpts,
	priorityFeeExtraPercent int,
	gasLimitExtraPercent uint64,
) (bool, *web3.GasOpts, error) {
	// Gas limit is too low to be correct
	if tarotCalculationOpts.EstimateGasLimitValue < gasLimitUsedExpectedMin {
		return false, nil, fmt.Errorf("estimated gas limit is too low => skip")
	}

	// Set new priority fee depending on competitors
	if tarotCalculationOpts.PriorityFeeValue == nil {
		//TODO zero log
		tarotCalculationOpts.PriorityFeeValue = zeroValue
		return false, nil, fmt.Errorf("priority fee is set to 0")
	}

	newPriorityFee_ := tarotCalculationOpts.PriorityFeeValue
	if tarotOpts.PriorityFee.Cmp(tarotCalculationOpts.PriorityFeeValue) == 1 {
		// Use the highest priority fee for the transaction
		newPriorityFee_ = tarotOpts.PriorityFee
	}

	newPriorityFee := utils.IncreaseAmount(newPriorityFee_, priorityFeeExtraPercent)
	rewardToken := ComputeReward(tarotCalculationOpts.VaultPendingRewardValue)
	rewardEth := utils.ConvertToEth(rewardToken, tarotCalculationOpts.RewardPairValue)

	gasOpts := web3.BuildTransactionFeeArgs(tarotCalculationOpts.BaseFeeValue, newPriorityFee, tarotCalculationOpts.EstimateGasLimitValue)
	gasOpts.GasLimit = tarotCalculationOpts.EstimateGasLimitValue + (tarotCalculationOpts.EstimateGasLimitValue*gasLimitExtraPercent)/100

	diff := utils.ComputeDifference(rewardEth, gasOpts.TransactionFee)
	isWorth := diff > -10
	log.Printf("reward erc20: %v; reward weth: %v; transaction fee: %v; priorityFee: %v, reward pair: %v; difference: %v", rewardToken, rewardEth, gasOpts.TransactionFee, newPriorityFee, tarotCalculationOpts.RewardPairValue, diff)
	return isWorth, gasOpts, nil
}

func ComputeReward(vaultPendingReward *big.Int) *big.Int {
	// Calculate fee = reward * REINVEST_FEE / 1e18
	reward := new(big.Int)
	reward.Mul(vaultPendingReward, reinvestFee) // reward * reinvestFee
	reward.Div(reward, utils.OneE18)            // divide by 1e18

	return reward
}
