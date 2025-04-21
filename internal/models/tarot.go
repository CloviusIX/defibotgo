package models

import (
	"github.com/ethereum/go-ethereum/common"
	"math/big"
)

type TarotOpts struct {
	ExtraPriorityFeePercent [2]int // [min %, max %]
	Chain                   Chain
	ProfitableThreshold     float64
	ReinvestBounty          *big.Int
	PriorityFee             *big.Int
	BlockRangeFilter        *big.Int
	Sender                  common.Address
	ContractLender          common.Address
	ContractGauge           common.Address
	ContractGasPriceOracle  common.Address
}
