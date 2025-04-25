package config

import (
	"defibotgo/internal/config"
	"defibotgo/internal/models"
	"github.com/ethereum/go-ethereum/common"
	"math/big"
)

var ImpermaxBaseSTKDUNIV2 = models.TarotOpts{
	ExtraPriorityFeePercent: [2]int{8, 20},
	Chain:                   models.Base,
	ProfitableThreshold:     -6,
	ReinvestBounty:          big.NewInt(int64(10000000000000000)),
	PriorityFee:             big.NewInt(56780),
	Sender:                  common.HexToAddress(config.GetSecret(config.WalletImpermaxAddressOne)),
	ContractLender:          common.HexToAddress("0xAa9F575a3fBF36d54FA3270fE25D4bB7Bb3bA3aE"),
	ContractGauge:           common.HexToAddress("0xA95EbEfbCB77Ae1daf0d2123784594F8ccE90274"),
	ContractGasPriceOracle:  BaseGasPriceOracleAddress,
}
