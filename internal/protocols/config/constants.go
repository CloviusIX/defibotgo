package config

import "github.com/ethereum/go-ethereum/common"

var ZeroAddress = common.Address{}
var ReinvestBounty = int64(20000000000000000) // 2% of fee
var BaseGasPriceOracleAddress = common.HexToAddress("0x420000000000000000000000000000000000000F")
