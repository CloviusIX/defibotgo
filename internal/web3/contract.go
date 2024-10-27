package web3

import (
	"fmt"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"strings"
)

// LoadAbi parses a given ABI string and returns the parsed ABI or an error.
//
// Parameters:
//   - abiStr: The ABI string to be parsed.
//
// Returns:
//   - abi.ABI: The parsed ABI structure.
//   - error: An error that occurred during ABI parsing, or nil if successful.
func LoadAbi(abiStr string) (abi.ABI, error) {
	parsedAbi, err := abi.JSON(strings.NewReader(abiStr))

	if err != nil {
		return abi.ABI{}, fmt.Errorf("failed to parse ABI file: %v", err)
	}

	return parsedAbi, err
}

// BuildContractInstance creates and returns a new bound contract instance for a given contract address and ABI string.
//
// Parameters:
//   - client: The client used to interact with the blockchain.
//   - contractAddress: The address of the smart contract on the blockchain.
//   - abiStr: The ABI string representing the contract's interface.
//
// Returns:
//   - *bind.BoundContract: The bound contract instance that allows interaction with the smart contract.
func BuildContractInstance(client *ethclient.Client, contractAddress common.Address, abiStr string) (*bind.BoundContract, error) {
	parsedAbi, err := LoadAbi(abiStr)
	if err != nil {
		return nil, err
	}

	contract := bind.NewBoundContract(contractAddress, parsedAbi, client, client, client)
	return contract, err
}
