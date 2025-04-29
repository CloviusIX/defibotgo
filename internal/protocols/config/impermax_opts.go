package config

import (
	"defibotgo/internal/config"
	"defibotgo/internal/models"
	"github.com/ethereum/go-ethereum/common"
	"math/big"
)

var ImpermaxBaseSTKDUNIV2 = models.TarotOpts{
	ReinvestBounty:          big.NewInt(int64(20000000000000000)),
	PriorityFee:             big.NewInt(56780),
	RewardRate:              big.NewInt(96262220291333273),
	ProfitableThreshold:     -6,
	GasUsedDefault:          770819,
	ExtraPriorityFeePercent: [2]int{15, 25},
	Chain:                   models.Base,
	Sender:                  common.HexToAddress(config.GetSecret(config.WalletImpermaxAddressOne)),
	ContractLender:          common.HexToAddress("0xAa9F575a3fBF36d54FA3270fE25D4bB7Bb3bA3aE"),
	ContractGauge:           common.HexToAddress("0xA95EbEfbCB77Ae1daf0d2123784594F8ccE90274"),
	ContractGasPriceOracle:  BaseGasPriceOracleAddress,
}
