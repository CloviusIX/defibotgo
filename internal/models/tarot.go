package models

import (
	"github.com/ethereum/go-ethereum/common"
	"math/big"
)

type TarotOpts struct {
	Chain            Chain
	Sender           common.Address
	PriorityFee      *big.Int
	BlockRangeFilter *big.Int
	ContractLender   common.Address
	ContractGauge    common.Address
}
