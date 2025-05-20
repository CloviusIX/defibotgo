package models

import (
	"github.com/ethereum/go-ethereum/common"
	"math/big"
)

type TarotOpts struct {
	ReinvestBounty          *big.Int
	PriorityFee             *big.Int
	RewardRate              *big.Int
	BlockRange              *big.Int
	ProfitableThreshold     float64
	GasUsedDefault          uint64
	ExtraPriorityFeePercent [2]int
	Chain                   Chain

	Sender                 common.Address
	ContractLender         common.Address
	ContractGauge          common.Address
	ContractGasPriceOracle common.Address
}
