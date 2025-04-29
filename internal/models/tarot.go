package models

import (
	"github.com/ethereum/go-ethereum/common"
	"math/big"
)

type TarotOpts struct {
	ReinvestBounty          *big.Int // 8 bytes, align 8
	PriorityFee             *big.Int // 8 bytes, align 8
	RewardRate              *big.Int // 8 bytes, align 8
	ProfitableThreshold     float64  // 8 bytes, align 8
	GasUsedDefault          uint64   // 8 bytes, align 8
	ExtraPriorityFeePercent [2]int   // 16 bytes, align 8
	Chain                   Chain    // string (ptr+len=16 B), align 8

	Sender                 common.Address // [20]byte, align 1
	ContractLender         common.Address // [20]byte
	ContractGauge          common.Address // [20]byte
	ContractGasPriceOracle common.Address // [20]byte
}
