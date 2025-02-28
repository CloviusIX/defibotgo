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
//   - cache: A cache instance to store the base fee per gas, reducing redundant lookups.
//   - cacheKey: A unique string key used to store and retrieve the base fee in the cache.
//   - ch: A channel for sending the result as a `models.WeiResult`, which includes the base fee and any error encountered.
//   - wg: A WaitGroup to signal completion of this asynchronous function to the calling function.
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
//   - cache: A cache instance for storing gas estimates based on a unique key, reducing repeated calculations.
//   - cacheKey: A unique string key for storing and retrieving the gas estimate in the cache.
//   - ch: A channel through which the function sends the result as a `models.GasLimitResult` containing the gas estimate and any error.
//   - wg: A WaitGroup used to signal completion of this asynchronous operation to the calling function.
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

	if estimateGas <= 100000 {
		ch <- models.GasLimitResult{Value: 0, Err: fmt.Errorf("estimage gas is too low %d", estimateGas)}
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
//   - cache: A cache instance to store results and reduce redundant calculations.
//   - cacheKey: The unique key for storing the calculated priority fee in the cache.
//   - ch: A channel used to send the result as a `models.WeiResult`, which includes the fee value and any error encountered.
//   - wg: A WaitGroup to ensure that the calling function waits for this function to complete. .WeiResult`.
func GetPriorityFeeAsync(ethClient *ethclient.Client, senderAddress common.Address, contractAddress common.Address, lastBlockN *big.Int, toBlock *big.Int, cache *ristretto.Cache, cacheKey string, ch chan models.WeiResult, wg *sync.WaitGroup) {
	defer wg.Done()
	if cacheResult, found := cache.Get(cacheKey); found {
		ch <- models.WeiResult{Value: cacheResult.(*big.Int), Err: nil}
		return
	}

	priorityFee, err := web3.GetPriorityFee(ethClient, senderAddress, contractAddress, lastBlockN, toBlock)

	if err != nil {
		ch <- models.WeiResult{Value: nil, Err: fmt.Errorf("failed to get last n events: %v", err)}
		return
	}

	cache.SetWithTTL(cacheKey, priorityFee, 1, utils.CacheShorterTime)
	ch <- models.WeiResult{Value: priorityFee, Err: nil}
}
