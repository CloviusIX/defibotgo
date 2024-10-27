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
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"log"
	"math/big"
	"sync"
	"time"
)

type WeiApiFunc func() (*big.Int, error)

type TarotOpts struct {
	Sender           common.Address
	PriorityFee      *big.Int
	BlockRangeFilter *big.Int
	ContractLender   common.Address
	ContractGauge    common.Address
}

var callOpts = bind.CallOpts{
	Pending:     true,
	BlockNumber: nil,
	Context:     context.Background(),
}

var reinvestFunctionName = "reinvest"

func Run(ethClient *ethclient.Client, ethClientWriter *ethclient.Client, chain models.Chain, opts *TarotOpts, walletPrivateKey *ecdsa.PrivateKey) {
	cache, err := ristretto.NewCache(&ristretto.Config{
		NumCounters: 50,  // 10x the number of items we expect to store
		MaxCost:     320, // Approx. 64 bytes per *big.Int, 5 keys in total
		BufferItems: 64,  // Recommended size for eviction buffer
	})
	if err != nil {
		panic(err)
	}

	contractLender, err := web3.BuildContractInstance(ethClient, opts.ContractLender, abi.CONTRACT_ABI_LENDER)
	if err != nil {
		//return nil, fmt.Errorf("error building tarot contract lender instance on %s: %v", chain, err)
		log.Fatalf("error building tarot contract lender instance on %s: %v", chain, err)
	}

	contractGauge, err := web3.BuildContractInstance(ethClient, opts.ContractGauge, abi.CONTRACT_ABI_GAUGE)
	if err != nil {
		//return nil, fmt.Errorf("error building tarot contract instance on %s: %v", chain, err)
		log.Fatalf("error building tarot contract gauge instance on %s: %v", opts.ContractGauge, err)
	}

	abiJson, err := web3.LoadAbi(abi.CONTRACT_ABI_LENDER)
	if err != nil {
		log.Fatalf("error loading Tarot abi on %s: %v", chain, err)
	}

	data, err := abiJson.Pack(reinvestFunctionName)
	if err != nil {
		log.Fatalf("failed to pack Tarot abi on %s: %v", chain, err)
	}

	// Create a message to simulate the transaction
	callMsg := ethereum.CallMsg{
		From:  opts.Sender,
		To:    &opts.ContractLender,
		Data:  data, // ABI-encoded function call data
		Value: big.NewInt(0),
	}
	log.Printf("calling contract lender on %s", opts.ContractLender.Hex())

	for {
		isWorth, gasOpts, err := getTransactionGasFees(ethClient, chain, contractGauge, callMsg, opts, cache)
		if err != nil {
			log.Printf("error getting gas on Tarot %s: %v", chain, err)
			time.Sleep(utils.RetryErrorSleep)
			continue
		}

		if isWorth {
			tx, err := web3.SendTransaction(ethClientWriter, contractLender, reinvestFunctionName, gasOpts, walletPrivateKey)
			if err != nil {
				log.Printf("failed to send transaction on Tarot %s: %v", chain, err)
				time.Sleep(utils.RetryErrorSleep)
				continue
			}

			log.Println("Sent transaction on Tarot ", tx.Hash().Hex())

			ctx, cancel := context.WithTimeout(context.Background(), time.Second*20)
			receipt, err := bind.WaitMined(ctx, ethClient, tx)
			cancel() // Immediately call cancel to free up resources

			if err != nil {
				log.Printf("failed to wait for receipt on Tarot %s: %v", chain, err)
			}

			if receipt.Status == types.ReceiptStatusSuccessful {
				log.Println("Successfully sent transaction on Tarot ", tx.Hash().Hex())
			} else {
				log.Printf("failed to send transaction on Tarot %s: %v", chain, err)
			}
		}
		time.Sleep(utils.RetryMainSleep)
	}
}

func getTransactionGasFees(ethClient *ethclient.Client, chain models.Chain, contractGauge *bind.BoundContract, msg ethereum.CallMsg, opts *TarotOpts, cache *ristretto.Cache) (bool, *web3.GasOpts, error) {
	var wg sync.WaitGroup
	wg.Add(5)

	vaultPendingRewardChan := make(chan models.WeiResult, 1)
	baseFeePerGasChan := make(chan models.WeiResult, 1)
	estimateGasChan := make(chan models.GasLimitResult, 1)
	rewardPairValueChan := make(chan models.WeiResult, 1)
	priorityFeeChan := make(chan models.WeiResult, 1)

	// Call web3 api asynchronously
	go web3Async.EthCallAsync(contractGauge, "earned", &callOpts, vaultPendingRewardChan, &wg, opts.ContractLender)
	go web3Async.GetBaseFeePerGasAsync(ethClient, callOpts.BlockNumber, cache, "1", baseFeePerGasChan, &wg)
	go web3Async.EstimateGasAsync(ethClient, msg, cache, "2", estimateGasChan, &wg)
	go web3Async.GetPriorityFeeAsync(ethClient, opts.Sender, opts.ContractLender, big.NewInt(50), callOpts.BlockNumber, cache, "3", priorityFeeChan, &wg)
	go asyncservices.GetPoolPriceAsync(chain, cache, "4", rewardPairValueChan, &wg)

	// Wait for goroutines and close the channel
	go func() {
		wg.Wait()
		close(vaultPendingRewardChan)
		close(baseFeePerGasChan)
		close(estimateGasChan)
		close(rewardPairValueChan)
		close(priorityFeeChan)
	}()

	// Get the channels result
	vaultPendingReward := <-vaultPendingRewardChan
	baseFeePerGas := <-baseFeePerGasChan
	estimateGasLimit := <-estimateGasChan
	rewardPairValue := <-rewardPairValueChan
	priorityFee := <-priorityFeeChan

	if vaultPendingReward.Err != nil || baseFeePerGas.Err != nil || estimateGasLimit.Err != nil || rewardPairValue.Err != nil || priorityFee.Err != nil {
		return false, nil, fmt.Errorf("error getting gas limit or reward pair value\nPending reward: %s\n:Base Fee: %s\nEstimate gas: %s\nReward Pair value: %s\nPriority fee: %s", vaultPendingReward.Err, baseFeePerGas.Err, estimateGasLimit.Err, rewardPairValue.Err, priorityFee.Err)
	}

	// Gas limit is too low to be correct
	if estimateGasLimit.Value < 100000 {
		return false, nil, fmt.Errorf("estimated gas limit is too low => skip")
	}

	//The gas limit function could be called from the parent ?
	saveGasLimit := estimateGasLimit.Value + (estimateGasLimit.Value*30)/100

	// Set new priority fee depending on competitors
	if priorityFee.Value == nil {
		print("Priority is set to 0")
		priorityFee.Value = big.NewInt(0)
	}

	newPriorityFee_ := priorityFee.Value
	if opts.PriorityFee.Cmp(priorityFee.Value) == 1 {
		newPriorityFee_ = opts.PriorityFee
	}

	newPriorityFee := utils.ApplyPercentage(newPriorityFee_, utils.RandomNumberInRange(20, 25))
	rewardToken := ComputeReward(vaultPendingReward.Value)
	rewardEth := utils.ConvertToEth(rewardToken, rewardPairValue.Value)

	gasOpts := web3.BuildTransactionFeeArgs(baseFeePerGas.Value, newPriorityFee, estimateGasLimit.Value)
	gasOpts.GasLimit = saveGasLimit

	diff := utils.ComputeDifference(rewardEth, gasOpts.TransactionFee)
	isWorth := diff > -10
	log.Printf("reward erc20: %v; reward weth: %v; transaction fee: %v; priorityFee: %v, reward pair: %v; difference: %v", rewardToken, rewardEth, gasOpts.TransactionFee, newPriorityFee, rewardPairValue.Value, diff)
	return isWorth, gasOpts, nil
}

func ComputeReward(vaultPendingReward *big.Int) *big.Int {
	reinvestFee := new(big.Int)
	reinvestFee.SetString("20000000000000000", 10) // 0.02 * 1e18 (which is 2e16)

	// Calculate fee = reward * REINVEST_FEE / 1e18
	reward := new(big.Int)
	reward.Mul(vaultPendingReward, reinvestFee) // reward * reinvestFee
	reward.Div(reward, utils.OneE18)            // divide by 1e18

	return reward
}
