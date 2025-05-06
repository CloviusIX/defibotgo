package config

import (
	"defibotgo/internal/config"
	"defibotgo/internal/models"
	"github.com/ethereum/go-ethereum/common"
	"math/big"
)

var TarotBaseUsdcAero = models.TarotOpts{
	ReinvestBounty:          big.NewInt(ReinvestBounty),
	PriorityFee:             big.NewInt(56780),
	RewardRate:              big.NewInt(1059238100440517689), // from gauge contract
	ProfitableThreshold:     -6,
	GasUsedDefault:          426244,
	ExtraPriorityFeePercent: [2]int{8, 20},
	Chain:                   models.Base,
	Sender:                  common.HexToAddress(config.GetSecret(config.WalletTarotAddressOne)),
	ContractLender:          common.HexToAddress("0x042c37762d1d126bc61eac2f5ceb7a96318f5db9"),
	ContractGauge:           common.HexToAddress("0x4F09bAb2f0E15e2A078A227FE1537665F55b8360"),
	ContractGasPriceOracle:  BaseGasPriceOracleAddress,
}

var TarotOptimismUsdcTarot = models.TarotOpts{
	ReinvestBounty:          big.NewInt(ReinvestBounty),
	PriorityFee:             big.NewInt(56780),
	RewardRate:              big.NewInt(1059238100440517689),
	ProfitableThreshold:     -6,
	GasUsedDefault:          426244,
	ExtraPriorityFeePercent: [2]int{8, 20},
	Chain:                   models.Optimism,
	Sender:                  common.HexToAddress(config.GetSecret(config.WalletTarotAddressOne)),
	ContractLender:          common.HexToAddress("0x80942A0066F72eFfF5900CF80C235dd32549b75d"),
	ContractGauge:           common.HexToAddress("0x73d5C2f4EB0E4EB15B3234f8B880A10c553DA1ea"),
	ContractGasPriceOracle:  common.HexToAddress("TODO ADD ADDRESS"),
}
