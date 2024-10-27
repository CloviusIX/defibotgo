package web3Async

import (
	"defibotgo/internal/models"
	"defibotgo/internal/utils"
	"defibotgo/internal/web3"
	"fmt"
	"github.com/dgraph-io/ristretto"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"math/big"
	"sync"
)

// EthCallAsync asynchronously calls a view (read-only) function on a smart contract and sends the result
// through a channel for concurrent processing.
//
// Parameters:
//   - contract: The smart contract instance to call the view function on.
//   - functionName: The name of the view function to invoke.
//   - callOpts: Options specifying the block number and context for the contract call.
//   - ch: A channel to send the result as a `models.WeiResult`, which includes the output value and any error encountered.
//   - wg: A WaitGroup used to signal completion of this asynchronous operation to the caller.
//   - params: Additional parameters required by the view function.
//
// Behavior:
//   - Calls the specified view function on the provided contract using the provided parameters.
//   - Sends the function’s result or any error encountered to the channel as a `models.WeiResult`.
//   - Signals the WaitGroup once processing is complete.
//
// Returns:
//   - None directly; the result or error is sent through the `ch` channel as `models.WeiResult`.
func EthCallAsync(contract *bind.BoundContract, functionName string, callOpts *bind.CallOpts, ch chan models.WeiResult, wg *sync.WaitGroup, params ...interface{}) {
	defer wg.Done()
	result, err := web3.EthCall(contract, functionName, callOpts, params...)

	if err != nil {
		ch <- models.WeiResult{Value: nil, Err: fmt.Errorf("failed to call contract function %s: %v", functionName, err)}
		return
	}

	ch <- models.WeiResult{Value: result, Err: nil}
}

// GetBaseFeePerGasAsync asynchronously retrieves the base fee per gas for a specified block and caches the result
// to avoid redundant blockchain queries.
//
// Parameters:
//   - ethClient: The Ethereum client instance used for blockchain interaction.
//   - blockNumber: The block number for which the base fee per gas is retrieved.
//   - cache: A Ristretto cache instance to store the base fee per gas, reducing redundant lookups.
//   - cacheKey: A unique string key used to store and retrieve the base fee in the cache.
//   - ch: A channel for sending the result as a `models.WeiResult`, which includes the base fee and any error encountered.
//   - wg: A WaitGroup to signal completion of this asynchronous function to the calling function.
//
// Behavior:
//   - Checks the cache for a previously stored base fee using the cacheKey. If found, sends the cached value on the channel.
//   - If not cached, retrieves the base fee per gas for the specified block and stores it in the cache with a TTL.
//   - Sends the retrieved base fee or any error encountered to the channel and signals completion to the WaitGroup.
//
// Returns:
//   - None directly; the base fee per gas or error is sent through the `ch` channel as `models.WeiResult`.
func GetBaseFeePerGasAsync(ethClient *ethclient.Client, blockNumber *big.Int, cache *ristretto.Cache, cacheKey string, ch chan models.WeiResult, wg *sync.WaitGroup) {
	defer wg.Done()
	if cacheResult, found := cache.Get(cacheKey); found {
		ch <- models.WeiResult{Value: cacheResult.(*big.Int), Err: nil}
		return
	}

	baseFee, err := web3.GetBaseFeePerGas(ethClient, blockNumber)

	if err != nil {
		ch <- models.WeiResult{Value: nil, Err: err}
		return
	}

	cache.SetWithTTL(cacheKey, baseFee, 1, utils.CacheTime)
	ch <- models.WeiResult{Value: baseFee, Err: nil}
}

// EstimateGasAsync asynchronously estimates the gas required to execute a specific Ethereum transaction
// and caches the result to reduce redundant calculations.
//
// Parameters:
//   - ethClient: The Ethereum client instance used to connect to the network.
//   - msg: The CallMsg struct defining the transaction details, such as 'From', 'To', 'Gas', 'GasPrice', 'Value', and 'Data'.
//   - cache: A Ristretto cache instance for storing gas estimates based on a unique key, reducing repeated calculations.
//   - cacheKey: A unique string key for storing and retrieving the gas estimate in the cache.
//   - ch: A channel through which the function sends the result as a `models.GasLimitResult` containing the gas estimate and any error.
//   - wg: A WaitGroup used to signal completion of this asynchronous operation to the calling function.
//
// Behavior:
//   - Checks if the gas estimate for the provided transaction is available in the cache using cacheKey.
//     If found, sends the cached result on the channel.
//   - If not cached, estimates the gas using the provided transaction details and stores the result in the cache with a TTL.
//   - Sends the estimated gas or any error encountered to the channel and signals completion to the WaitGroup.
//
// Returns:
//   - None directly; the gas estimate or error is sent through the `ch` channel as `models.GasLimitResult`.
func EstimateGasAsync(ethClient *ethclient.Client, msg ethereum.CallMsg, cache *ristretto.Cache, cacheKey string, ch chan models.GasLimitResult, wg *sync.WaitGroup) {
	defer wg.Done()
	if cacheResult, found := cache.Get(cacheKey); found {
		ch <- models.GasLimitResult{Value: cacheResult.(uint64), Err: nil}
		return
	}

	estimateGas, err := web3.EstimateGas(ethClient, msg)
	if err != nil {
		ch <- models.GasLimitResult{Value: 0, Err: err}
		return
	}

	cache.SetWithTTL(cacheKey, estimateGas, 1, utils.CacheTime)
	ch <- models.GasLimitResult{Value: estimateGas, Err: nil}
}

// GetPriorityFeeAsync asynchronously calculates the maximum priority fee among recent transactions
// for a specified contract, excluding transactions from a given sender address.
//
// Parameters:
//   - ethClient: The Ethereum client instance for blockchain interaction.
//   - senderAddress: The Ethereum address of the sender whose transactions are to be excluded.
//   - contractAddress: The contract's Ethereum address for which recent transactions are analyzed.
//   - txCount: The number of past transactions to retrieve and analyze.
//   - lastBlockN: Number of blocks before toBlock to start fetching transactions (e.g., 50 means start from 50 blocks before toBlock).
//   - toBlock: The block number up to which transactions are considered.
//   - cache: A Ristretto cache instance to store results and reduce redundant calculations.
//   - cacheKey: The unique key for storing the calculated priority fee in the cache.
//   - ch: A channel used to send the result as a `models.WeiResult`, which includes the fee value and any error encountered.
//   - wg: A WaitGroup to ensure that the calling function waits for this function to complete.
//
// Behavior:
//   - Checks the cache for an existing priority fee using the provided cacheKey. If found, it sends the cached result on the channel.
//   - If not cached, retrieves recent transactions up to the specified block, calculates the maximum priority fee, and caches it with a time-to-live (TTL).
//   - Sends the result or any error encountered to the channel and signals completion to the WaitGroup.
//
// Returns:
//   - None directly; however, the maximum priority fee or an error is sent through the `ch` channel as `models.WeiResult`.
func GetPriorityFeeAsync(ethClient *ethclient.Client, senderAddress common.Address, contractAddress common.Address, txCount int, lastBlockN *big.Int, toBlock *big.Int, cache *ristretto.Cache, cacheKey string, ch chan models.WeiResult, wg *sync.WaitGroup) {
	defer wg.Done()
	if cacheResult, found := cache.Get(cacheKey); found {
		ch <- models.WeiResult{Value: cacheResult.(*big.Int), Err: nil}
		return
	}

	priorityFee, err := web3.GetPriorityFee(ethClient, senderAddress, contractAddress, txCount, lastBlockN, toBlock)

	if err != nil {
		ch <- models.WeiResult{Value: nil, Err: fmt.Errorf("failed to get last n events: %v", err)}
		return
	}

	cache.SetWithTTL(cacheKey, priorityFee, 1, utils.CacheTime)
	ch <- models.WeiResult{Value: priorityFee, Err: nil}
}
