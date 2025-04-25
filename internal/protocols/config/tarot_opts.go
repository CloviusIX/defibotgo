package config

import (
	"defibotgo/internal/config"
	"defibotgo/internal/models"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"math/big"
)

var zeroAddress = common.Address{}
var reinvestBounty = int64(10000000000000000) // 1% of fee
var baseGasPriceOracleAddress = common.HexToAddress("0x420000000000000000000000000000000000000F")

func GetTarotBaseUsdcAero() (*models.TarotOpts, error) {
	var opts = models.TarotOpts{
		ExtraPriorityFeePercent: [2]int{8, 20},
		Chain:                   models.Base,
		ProfitableThreshold:     -6,
		ReinvestBounty:          big.NewInt(reinvestBounty),
		PriorityFee:             big.NewInt(56780),
		Sender:                  common.HexToAddress(config.GetSecret(config.WalletTarotAddressOne)),
		ContractLender:          common.HexToAddress("0x042c37762d1d126bc61eac2f5ceb7a96318f5db9"),
		ContractGauge:           common.HexToAddress("0x4f09bab2f0e15e2a078a227fe1537665f55b8360"),
		ContractGasPriceOracle:  baseGasPriceOracleAddress,
	}

	if opts.Sender == zeroAddress {
		return nil, fmt.Errorf("sender for %s is empty", opts.ContractLender)
	}

	return &opts, nil
}

func GetTarotOptimismUsdcTarot() (*models.TarotOpts, error) {
	var opts = models.TarotOpts{
		ExtraPriorityFeePercent: [2]int{8, 20},
		Chain:                   models.Optimism,
		ProfitableThreshold:     -6,
		ReinvestBounty:          big.NewInt(reinvestBounty),
		PriorityFee:             big.NewInt(56780),
		Sender:                  common.HexToAddress(config.GetSecret(config.WalletTarotAddressOne)),
		ContractLender:          common.HexToAddress("0x80942A0066F72eFfF5900CF80C235dd32549b75d"),
		ContractGauge:           common.HexToAddress("0x73d5C2f4EB0E4EB15B3234f8B880A10c553DA1ea"),
		ContractGasPriceOracle:  common.HexToAddress("TODO ADD ADDRESS"),
	}

	if opts.Sender == zeroAddress {
		return nil, fmt.Errorf("sender for %s is empty", opts.ContractLender)
	}

	return &opts, nil
}
