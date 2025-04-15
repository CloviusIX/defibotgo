package models

import (
	"github.com/ethereum/go-ethereum/common"
	"math/big"
)

type TarotOpts struct {
	ReinvestBounty   *big.Int
	PriorityFee      *big.Int
	BlockRangeFilter *big.Int
	Chain            Chain
	Sender           common.Address
	ContractLender   common.Address
	ContractGauge    common.Address
}
