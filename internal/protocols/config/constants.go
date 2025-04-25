package config

import "github.com/ethereum/go-ethereum/common"

var ZeroAddress = common.Address{}
var ReinvestBounty = int64(10000000000000000) // 1% of fee
var BaseGasPriceOracleAddress = common.HexToAddress("0x420000000000000000000000000000000000000F")
