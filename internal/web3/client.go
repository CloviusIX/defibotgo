package web3

import (
	"context"
	"crypto/ecdsa"
	"defibotgo/internal/config"
	"defibotgo/internal/models"
	"fmt"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"log"
	"math/big"
)

// BuildWeb3Client initializes a new Web3 client for the specified blockchain network.
//
// Parameters:
//   - chain: The blockchain network for which to build the Web3 client (e.g., models.Optimism).
//
// Returns:
//   - *ethclient.Client: The initialized Ethereum client.
//   - error: An error that occurred during the connection attempt, or nil if successful.
func BuildWeb3Client(chain models.Chain, asReader bool) (*ethclient.Client, error) {
	rpcUrl := config.ChainToRpcUrlWrite[chain]
	if rpcUrl == "" {
		return nil, fmt.Errorf("rpc url is empty")
	}

	if asReader {
		rpcUrl = config.ChainToRpcUrlRead[chain]
	}

	ethClient, err := ethclient.Dial(rpcUrl)

	if err != nil {
		return nil, fmt.Errorf("failed to build test Web3 client: %v", err)
	}

	return ethClient, err
}

// SendTransaction sends a transaction to a smart contract using the specified write function.
//
// Parameters:
//   - ethClient: The Ethereum client used to interact with the Ethereum network.
//   - contract: The bound smart contract to send the transaction to.
//   - functionName: The name of the smart contract write function to invoke.
//   - gasOpts: Options for specifying gas limits and fees (GasLimit, GasFeeCap, GasTipCap).
//   - walletPrivateKey: The private key of the wallet sending the transaction, used for signing.
//
// Returns:
//   - *types.Transaction: The transaction object representing the sent transaction.
//   - error: An error that occurred while sending the transaction, or nil if successful.
func SendTransaction(ethClient *ethclient.Client, contract *bind.BoundContract, functionName string, gasOpts *GasOpts, walletPrivateKey *ecdsa.PrivateKey, params ...interface{}) (*types.Transaction, error) {
	chainID, err := ethClient.ChainID(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to get ChainID: %v", err)
	}

	transactionOpts, err := bind.NewKeyedTransactorWithChainID(walletPrivateKey, chainID)
	if err != nil {
		return nil, fmt.Errorf("failed to create transaction options: %v", err)
	}

	transactionOpts.GasLimit = gasOpts.GasLimit
	transactionOpts.GasFeeCap = gasOpts.GasFeeCap
	transactionOpts.GasTipCap = gasOpts.GasTipCap

	tx, err := contract.Transact(transactionOpts, functionName, params...)
	if err != nil {
		return nil, fmt.Errorf("failed to send transaction %s: %v", functionName, err)
	}

	return tx, err
}

// EthCall calls a view (read-only) function on a smart contract.
//
// Parameters:
//   - contract: The smart contract instance to call the view function on.
//   - functionName: The name of the view function to invoke.
//   - callOpts: Options specifying the block number and context for the contract call.
//   - params: Additional parameters to pass to the view function.
//
// Returns:
//   - big.int: The result of the view function call.
//   - error: An error that occurred during the contract call, or nil if successful.
func EthCall(contract *bind.BoundContract, functionName string, callOpts *bind.CallOpts, params ...interface{}) (*big.Int, error) {
	var results []interface{}
	err := contract.Call(callOpts, &results, functionName, params...)

	if err != nil {
		return nil, fmt.Errorf("failed to call contract function %s: %v", functionName, err)
	}

	for _, result := range results {
		if output, ok := result.(*big.Int); ok {
			return output, nil
		}
		// Handle other potential types as needed
		// We are not interested of having other result than big.int for now
	}

	// Send an error if no valid result was found
	return nil, fmt.Errorf("unexpected result type; expected *big.Int")
}

// GetBaseFeePerGas retrieves the base fee per gas for a specific block.
//
// Parameters:
//   - ethClient: The Ethereum client instance used for blockchain interaction.
//   - blockNumber: The block number for which the base fee per gas is retrieved.
//
// Returns:
//   - *big.Int: The base fee per gas for the specified block.
func GetBaseFeePerGas(ethClient *ethclient.Client, blockNumber *big.Int) (*big.Int, error) {
	header, err := ethClient.HeaderByNumber(context.Background(), blockNumber)

	if err != nil {
		return nil, fmt.Errorf("failed to get latest block header: %v", err)
	}

	return header.BaseFee, nil
}

// EstimateGas estimates the gas required to execute a specific Ethereum transaction.
//
// Parameters:
//   - ethClient: The Ethereum client instance used to connect to the network.
//   - msg: The CallMsg struct defining the transaction details, such as 'From', 'To', 'Gas', 'GasPrice', 'Value', and 'Data'
//
// Returns:
//   - uint64: The estimated gas needed for the transaction execution.
//   - error: An error if the gas estimation fails, or nil if successful.
func EstimateGas(ethClient *ethclient.Client, msg ethereum.CallMsg) (uint64, error) {
	estimateGas, err := ethClient.EstimateGas(context.Background(), msg)
	if err != nil {
		return 0, fmt.Errorf("failed to estimate gas estimation: %v", err)
	}

	return estimateGas, nil
}

// GetPriorityFee calculates the maximum priority fee among recent transactions
// for a specified contract, excluding transactions from a given sender address.
//
// Parameters:
//   - ethClient: The Ethereum client instance for blockchain interaction.
//   - senderAddress: The Ethereum address of the sender whose transactions are to be excluded.
//   - contractAddress: The contract's Ethereum address for which recent transactions are analyzed.
//   - txCount: The number of past transactions to retrieve and analyze.
//   - lastBlockN: Number of blocks before toBlock to start fetching transactions (e.g., 50 means start from 50 blocks before toBlock).
//   - toBlock: The block number up to which transactions are considered.
//
// Returns:
//   - *big.Int: The maximum priority fee found among the transactions, or 0 if none are found.
//   - error: An error if there was an issue fetching transactions or processing them.
func GetPriorityFee(ethClient *ethclient.Client, senderAddress common.Address, contractAddress common.Address, lastBlockN *big.Int, toBlock *big.Int) (*big.Int, error) {
	transactions, err := getPastTransactions(ethClient, contractAddress, lastBlockN, toBlock)
	senderAddressStr := senderAddress.Hex()
	if err != nil {
		return nil, fmt.Errorf("failed to get last n events: %v", err)
	}

	maxPriorityFee := big.NewInt(0)

	for _, transaction := range transactions {
		txSender, errSender := getSender(transaction)
		if errSender != nil {
			log.Printf("failed to get sender: %v", errSender)
			continue
		}
		if txSender.Hex() != senderAddressStr && transaction.GasTipCap().Cmp(maxPriorityFee) == 1 {
			maxPriorityFee = transaction.GasTipCap()
		}
	}

	return maxPriorityFee, nil
}

// getPastTransactions retrieves past transactions for a given contract address within a specified block range.
//
// Parameters:
//   - ethClient: The Ethereum client used to interact with the blockchain.
//   - contractAddress: The address of the contract for which past transactions are retrieved.
//   - txCount: The desired number of transactions to retrieve (e.g., the last 5 or 3 transactions).
//   - lastBlockN: The number of blocks to go back from the toBlock (e.g., 50 means start 50 blocks before toBlock).
//   - toBlock: The block number up to which transactions should be fetched. If nil, the latest block will be used.
//
// Returns:
//   - []*types.Transaction: A slice of transactions matching the criteria, up to the specified txCount.
//   - error: An error if there was an issue retrieving the transactions.
func getPastTransactions(ethClient *ethclient.Client, contractAddress common.Address, lastBlockN *big.Int, toBlock *big.Int) ([]*types.Transaction, error) {
	toBlockResult := toBlock

	if toBlockResult == nil {
		latestBlock, err := ethClient.BlockNumber(context.Background())

		if err != nil {
			return nil, fmt.Errorf("failed to get latest block: %v", err)
		}

		toBlockResult = big.NewInt(int64(latestBlock))
	}

	if lastBlockN == nil {
		lastBlockN = big.NewInt(0) // Default to 0 blocks if lastBlockN is not provided
	}

	filterQuery := ethereum.FilterQuery{
		FromBlock: new(big.Int).Sub(toBlockResult, lastBlockN),
		ToBlock:   toBlockResult,
		Addresses: []common.Address{contractAddress},
	}

	logs, err := ethClient.FilterLogs(context.Background(), filterQuery)

	if err != nil {
		return nil, fmt.Errorf("failed to filter logs: %v", err)
	}

	var transactions []*types.Transaction

	for i, _log := range logs {
		if i-1 >= 0 && _log.TxHash == logs[i-1].TxHash {
			// skip redundant transactions
			continue
		}
		tx, _, err := ethClient.TransactionByHash(context.Background(), _log.TxHash)
		if err != nil {
			log.Printf("failed to get transaction by hash: %v", err)
			continue
		}

		transactions = append(transactions, tx)
	}

	return transactions, nil
}

// getSender retrieves the sender address of a given transaction.
//
// Parameters:
//   - tx: A pointer to the transaction from which to extract the sender address.
//
// Returns:
//   - common.Address: The sender address of the transaction.
//   - error: An error if there was an issue retrieving the sender address.
func getSender(tx *types.Transaction) (common.Address, error) {
	var signer types.Signer
	switch {
	case tx.Type() == types.AccessListTxType:
		signer = types.NewEIP2930Signer(tx.ChainId())
	case tx.Type() == types.DynamicFeeTxType:
		signer = types.NewLondonSigner(tx.ChainId())
	default:
		signer = types.NewEIP155Signer(tx.ChainId())
	}
	sender, err := types.Sender(signer, tx)

	return sender, err
}
