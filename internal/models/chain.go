package models

// Chain represent a blockchain network
type Chain string

// Define supported chains as constants
const (
	Optimism Chain = "OPTIMISM"
)

// ChainToRpcUrlRead maps a Chain to its RPC URL for view functions
var ChainToRpcUrlRead = map[Chain]string{
	Optimism: "https://mainnet.optimism.io",
}

// ChainToRpcUrlWrite maps a Chain to its RPC URL for write functions
var ChainToRpcUrlWrite = map[Chain]string{
	Optimism: "https://mainnet.optimism.io",
}
