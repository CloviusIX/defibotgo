package web3

import (
	"context"
	"defibotgo/internal/abi"
	"defibotgo/internal/config"
	"defibotgo/internal/models"
	"defibotgo/internal/web3"
	"fmt"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/rs/zerolog/log"
	"math/big"
	"testing"
)

var testWalletAddress = common.HexToAddress("0x19719b8d58376F3480Bc98e91eCcA64640f6D520")
var contractAbiTest = `[{"inputs":[],"name":"get","outputs":[{"internalType":"uint256","name":"","type":"uint256"}],"stateMutability":"view","type":"function"},{"inputs":[{"internalType":"uint256","name":"_value","type":"uint256"}],"name":"set","outputs":[],"stateMutability":"nonpayable","type":"function"}]`
var testnetRpcUrl = "https://sepolia.optimism.io"
var contractTestAddress = common.HexToAddress("0x39Df60Fcb52Bf97dFf6Fb5bDa969A131Eb99bB80")
var testWriteFunction = "set"

var chain = models.Optimism
var blockNumber = big.NewInt(124626577)
var callOpts = bind.CallOpts{
	Pending:     false,
	BlockNumber: blockNumber,
	Context:     context.Background(),
}

func TestSendTransaction(t *testing.T) {
	functionParam := big.NewInt(1)
	gasLimit := uint64(1090381)
	priorityFee := big.NewInt(5275)
	walletPrivateKey := config.GetSecret(config.WalletTestPrivateKey)

	if walletPrivateKey == "" {
		log.Fatal().Msg("wallet test private key not found")
	}

	testPrivateKey, err := crypto.HexToECDSA(walletPrivateKey)
	if err != nil {
		t.Fatalf("Failed to load account private key: %v", err)
	}

	ethClient, err := ethclient.Dial(testnetRpcUrl)
	if err != nil {
		t.Fatalf("Failed to build test Web3 client: %v", err)
	}

	baseFee, err := web3.GetBaseFeePerGas(ethClient, nil)
	if err != nil {
		t.Fatalf("Failed to get base fee: %v", err)
	}

	contract, err := web3.BuildContractInstance(ethClient, contractTestAddress, contractAbiTest)
	if err != nil {
		t.Fatalf("Failed to build contract: %v", err)
	}

	gasOpts := web3.BuildTransactionFeeArgs(baseFee, priorityFee, gasLimit)
	tx, err := web3.SendTransaction(ethClient, contract, testWriteFunction, gasOpts, testPrivateKey, functionParam)

	if tx == nil || err != nil {
		t.Fatalf("Failed to send transaction: %v", err)
	}
}

func TestEthCallGetTarotEarnedFn(t *testing.T) {
	expected := new(big.Int)
	expected.SetString("80409552891557643738", 10)

	abiStr := abi.CONTRACT_ABI_GAUGE
	contractAddressLender := common.HexToAddress("0x3b749be6ca33f27e2837138ede69f8c6c53f9207")
	contractAddressGauge := common.HexToAddress("0x1239c54d9fd91e6ecec8eaad80df0fed43c47673")
	functionName := "earned"

	ethClient, err := web3.BuildWeb3Client(chain, true)

	if err != nil {
		t.Fatalf("Failed to build web3 client: %v", err)
	}

	contract, err := web3.BuildContractInstance(ethClient, contractAddressGauge, abiStr)
	if err != nil {
		t.Fatalf("Failed to build contract: %v", err)
	}

	result, err := web3.EthCall(contract, functionName, &callOpts, contractAddressLender)
	if err != nil {
		t.Fatalf("EthCall function %s failed: %v", functionName, err)
	}

	if result.Cmp(expected) != 0 {
		t.Fatalf("EthCall function %s failed, expected %d, got %d", functionName, expected, result)
	}
}

func TestGetBaseFeePerGas(t *testing.T) {
	ethClient, err := web3.BuildWeb3Client(chain, true)
	if err != nil {
		t.Fatalf("Failed to build web3 client: %v", err)
	}

	baseFeePerGasByBlockNumber, err := web3.GetBaseFeePerGas(ethClient, nil)
	if err != nil {
		t.Fatalf("Failed to get base fee per gas: %v", err)
	}

	baseFeePerGasLatest, err := web3.GetBaseFeePerGas(ethClient, blockNumber)
	if err != nil {
		t.Fatalf("Failed to get base fee per gas: %v", err)
	}

	if baseFeePerGasByBlockNumber.Cmp(baseFeePerGasLatest) == 0 && baseFeePerGasByBlockNumber.Cmp(big.NewInt(0)) < 1 && baseFeePerGasLatest.Cmp(big.NewInt(0)) < 1 {
		t.Fatalf("GetBaseFeePerGas failed, got by block number %v and from the latest %v", baseFeePerGasByBlockNumber, baseFeePerGasLatest)
	}
}
func TestEstimateGas(t *testing.T) {
	fromAddress := testWalletAddress
	toAddress := contractTestAddress
	ethClient, err := ethclient.Dial(testnetRpcUrl)
	if err != nil {
		t.Fatalf("Failed to build web3 client: %v", err)
	}

	abiJson, err := web3.LoadAbi(contractAbiTest)
	if err != nil {
		t.Fatalf("Failed to load abi: %v", err)
	}

	data, err := abiJson.Pack(testWriteFunction, big.NewInt(1))
	if err != nil {
		t.Fatalf("Failed to pack abi: %v", err)
	}

	// Create a message to simulate the transaction
	msg := ethereum.CallMsg{
		From:  fromAddress,
		To:    &toAddress,
		Data:  data, // ABI-encoded function call data
		Value: big.NewInt(0),
	}

	estimateGas, err := web3.EstimateGas(ethClient, msg)

	if err != nil {
		t.Fatalf("Failed to estimate gas: %v", err)
	}

	if estimateGas == 0 {
		t.Fatalf("Failed to estimate gas, expected non-zero value")
	}
}

func TestGetPriorityFee(t *testing.T) {
	senderAddressA := common.HexToAddress("0x44017f6b5774Fd1DfAe40D8d743BFdfFdd42A0f0")
	contractAddressA := common.HexToAddress("0x3b749be6ca33f27e2837138ede69f8c6c53f9207")

	senderAddressB := common.HexToAddress("0xb8cc2829d05d12acd85329a738b290c4b76d3211")
	contractAddressB := common.HexToAddress("0x80942A0066F72eFfF5900CF80C235dd32549b75d")

	testCases := []struct {
		name            string
		senderAddress   common.Address
		contractAddress common.Address
		toBlock         *big.Int
		expected        *big.Int
	}{
		{"Transactions are from the sender_Returns 0", senderAddressA, contractAddressA, big.NewInt(124314120), big.NewInt(0)},
		{"Transactions are shared with other wallets_Returns the highest priority fee", senderAddressB, contractAddressB, big.NewInt(124791661), big.NewInt(154275)},
		{"Transactions are shared between the sender address and other wallets_Return the other account highest priority fee", senderAddressB, contractAddressB, big.NewInt(124314024), big.NewInt(70260)},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := getPriorityFee(tc.senderAddress, tc.contractAddress, tc.toBlock)
			if err != nil {
				t.Fatalf("getPriorityFee failed: %v", err)
			} else if result.Cmp(tc.expected) != 0 {
				t.Fatalf("getPriorityFee failed, expected %v, got %v", tc.expected, result)
			}
		})
	}
}

func getPriorityFee(senderAddress common.Address, contractAddress common.Address, toBlock *big.Int) (*big.Int, error) {
	ethClient, err := web3.BuildWeb3Client(chain, true)
	lastBlockN := big.NewInt(50)

	if err != nil {
		return nil, fmt.Errorf("failed to build web3 client: %v", err)
	}

	priorityFee, err := web3.GetPriorityFee(ethClient, senderAddress, contractAddress, lastBlockN, toBlock)

	if err != nil {
		return nil, fmt.Errorf("failed to get priority fee: %v", err)
	}

	return priorityFee, nil
}
