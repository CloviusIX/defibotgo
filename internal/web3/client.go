package web3

import (
	"defibotgo/internal/models"
	"github.com/ethereum/go-ethereum/ethclient"
)

// BuildWeb3Client initializes a new Web3 client for the specified blockchain network (chain).
//
// Parameters:
//   - chain: The blockchain network for which the Web3 client should be built (e.g., models.Optimism).
//
// Returns:
//   - *ethclient.Client: The initialized Ethereum client.
//   - error: An error that occurred during the connection attempt, or nil if successful.
func BuildWeb3Client(chain models.Chain) (*ethclient.Client, error) {
	rpcUrl := models.ChainToRpcUrlRead[chain]
	client, err := ethclient.Dial(rpcUrl)

	if err != nil {
		return nil, err
	}

	return client, err
}
