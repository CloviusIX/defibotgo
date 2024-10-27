package utils

import "math/big"

func ComputeDifference(value1 *big.Int, value2 *big.Int) float64 {
	// Calculate the difference
	difference := new(big.Int).Sub(value1, value2)

	// Convert big.Int values to float64
	diffFloat := new(big.Float).SetInt(difference)
	value2Float := new(big.Float).SetInt(value2)

	// Perform the division and multiply by 100 to get the percentage increase
	percentageIncrease, _ := new(big.Float).Quo(diffFloat, value2Float).Float64()

	// Multiply by 100 to convert to percentage
	return percentageIncrease * 100
}
