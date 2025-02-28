package utils

import "math/big"

var OneE18, _ = new(big.Int).SetString("1000000000000000000", 10)

func ConvertToEth(value *big.Int, ratio *big.Int) *big.Int {
	// Create a new big.Int to hold the result
	wethRewardWei := new(big.Int)

	// Multiply reward by the pair value (both in Wei)
	wethRewardWei.Mul(value, ratio)

	// Divide by 1e18 to get the correct WETH value in Wei
	wethRewardWei.Div(wethRewardWei, OneE18)

	// Return the final reward in WETH (Wei)
	return wethRewardWei
}
