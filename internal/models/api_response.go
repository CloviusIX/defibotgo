package models

import "math/big"

type WeiResult struct {
	Value *big.Int
	Err   error
}

type GasLimitResult struct {
	Value uint64
	Err   error
}
