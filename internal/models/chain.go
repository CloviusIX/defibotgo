package models

// Chain represent a blockchain network
type Chain string
type Protocol string
type Pool string

// Define supported chains as constants
const (
	Optimism Chain = "OPTIMISM"
	Base     Chain = "BASE"
)

// Define supported protocol as constants
const (
	Tarot    Protocol = "TAROT"
	Impermax Protocol = "IMPERMAX"
)

// Define supported pool ID as constants
const (
	FbombCbbtc Pool = "FBOMB_CBBTC"
	UsdcAero   Pool = "USDC_AERO"
	WethTarot  Pool = "WETH_TAROT"
)
