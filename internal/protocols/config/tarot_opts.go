package config

import (
	"defibotgo/internal/config"
	"defibotgo/internal/models"
	"github.com/ethereum/go-ethereum/common"
	"math/big"
)

var TarotBaseUsdcAero = models.TarotOpts{
	ExtraPriorityFeePercent: [2]int{8, 20},
	Chain:                   models.Base,
	ProfitableThreshold:     -6,
	ReinvestBounty:          big.NewInt(ReinvestBounty),
	PriorityFee:             big.NewInt(56780),
	Sender:                  common.HexToAddress(config.GetSecret(config.WalletTarotAddressOne)),
	ContractLender:          common.HexToAddress("0x042c37762d1d126bc61eac2f5ceb7a96318f5db9"),
	ContractGauge:           common.HexToAddress("0x4f09bab2f0e15e2a078a227fe1537665f55b8360"),
	ContractGasPriceOracle:  BaseGasPriceOracleAddress,
}

var TarotOptimismUsdcTarot = models.TarotOpts{
	ExtraPriorityFeePercent: [2]int{8, 20},
	Chain:                   models.Optimism,
	ProfitableThreshold:     -6,
	ReinvestBounty:          big.NewInt(ReinvestBounty),
	PriorityFee:             big.NewInt(56780),
	Sender:                  common.HexToAddress(config.GetSecret(config.WalletTarotAddressOne)),
	ContractLender:          common.HexToAddress("0x80942A0066F72eFfF5900CF80C235dd32549b75d"),
	ContractGauge:           common.HexToAddress("0x73d5C2f4EB0E4EB15B3234f8B880A10c553DA1ea"),
	ContractGasPriceOracle:  common.HexToAddress("TODO ADD ADDRESS"),
}
