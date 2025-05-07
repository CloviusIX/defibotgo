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

var TarotBaseWethTarot = models.TarotOpts{
	ReinvestBounty:          big.NewInt(ReinvestBounty),
	PriorityFee:             big.NewInt(56780),
	RewardRate:              big.NewInt(66885542988906833), // from gauge contract
	ProfitableThreshold:     -6,
	GasUsedDefault:          856360,
	ExtraPriorityFeePercent: [2]int{5, 10},
	Chain:                   models.Base,
	Sender:                  common.HexToAddress(config.GetSecret(config.WalletTarotAddressTwo)),
	ContractLender:          common.HexToAddress("0xb556ee2761F5D2887b8f35a7ddA367aBd20503bf"),
	ContractGauge:           common.HexToAddress("0xa81dac2e9caa218Fcd039D7CEdEB7847cf362213"),
	ContractGasPriceOracle:  BaseGasPriceOracleAddress,
}

var TarotBaseAeroTarot = models.TarotOpts{
	ReinvestBounty:          big.NewInt(ReinvestBounty),
	PriorityFee:             big.NewInt(56780),
	RewardRate:              big.NewInt(99537395303616726), // from gauge contract
	ProfitableThreshold:     -6,
	GasUsedDefault:          407294,
	ExtraPriorityFeePercent: [2]int{5, 10},
	Chain:                   models.Base,
	Sender:                  common.HexToAddress(config.GetSecret(config.WalletTarotAddressThree)),
	ContractLender:          common.HexToAddress("0x776236aeAD8A58AC9eC3CF214CDa3c6335f46B2d"),
	ContractGauge:           common.HexToAddress("0x65B4A4b9813E37DA640bbEf8AbDD8E47100bE5f8"),
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
