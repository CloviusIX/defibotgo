package utils

import (
	"math/big"
	"math/rand"
)

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

// IncreaseAmount adds a given percentage to the original *big.Int value.
// For example, if the value is 1000 and percent is 25, the function returns 1250.
func IncreaseAmount(value *big.Int, percent int) *big.Int {
	// Convert percent to a *big.Int
	percentage := big.NewInt(int64(percent))

	// Calculate the additional percentage amount
	increase := new(big.Int).Mul(value, percentage)
	increase.Div(increase, big.NewInt(100))

	// Add the increase to the original value
	result := new(big.Int).Add(value, increase)

	return result
}

// DecreaseAmount adds a given percentage to the original *big.Int value.
// For example, if the value is 1000 and percent is 25, the function returns 1250.
func DecreaseAmount(value *big.Int, percent int) *big.Int {
	// Convert percent to a *big.Int
	percentage := big.NewInt(int64(percent))

	// Calculate the additional percentage amount
	decrease := new(big.Int).Mul(value, percentage)
	decrease.Div(decrease, big.NewInt(100))

	// Add the increase to the original value
	result := new(big.Int).Sub(value, decrease)

	return result
}

// RandomNumberInRange generates a random int64 number between min and max (inclusive).
func RandomNumberInRange(min int, max int) int {
	if min >= max {
		panic("min should be less than max")
	}

	// Generate a random number between min and max
	return min + rand.Intn(max-min+1)
}
