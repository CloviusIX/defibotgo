package web3

import (
	"math/big"
)

type GasOpts struct {
	TransactionFee *big.Int // Transaction fee to pay
	GasFeeCap      *big.Int // Gas fee cap to use for the 1559 transaction execution (nil = gas price oracle)
	GasTipCap      *big.Int // Gas priority fee cap to use for the 1559 transaction execution (nil = gas price oracle)
	GasLimit       uint64   // Gas limit to set for the transaction execution (0 = estimate)
}

// BuildTransactionFeeArgs constructs the transaction fee options needed for a EIP-1559 transaction.
//
// Parameters:
//   - baseFee: The base fee per gas in the network, which represents the minimum gas price for a transaction.
//   - priorityFee: The maximum priority fee per gas to incentivize miners to prioritize this transaction.
//   - gasLimit: The maximum amount of gas units that can be consumed by the transaction.
//
// Returns:
//   - *GasOpts: A struct containing the calculated transaction fee, gas fee cap, gas tip cap, and gas limit.
func BuildTransactionFeeArgs(baseFee *big.Int, priorityFee *big.Int, gasLimit uint64) *GasOpts {
	maxFee := ComputeMaxFee(baseFee, priorityFee)
	return &GasOpts{
		TransactionFee: computeTransactionFee(gasLimit, maxFee),
		GasFeeCap:      maxFee,
		GasTipCap:      priorityFee,
		GasLimit:       gasLimit,
	}
}

// ComputeMaxFee calculates the maximum fee per gas for a transaction.
//
// Parameters:
//   - baseFee: The base fee per gas in the network.
//   - priorityFee: The maximum priority fee per gas set by the sender to incentivize faster transaction processing.
//
// Returns:
//   - *big.Int: The maximum fee per gas, computed as the sum of the base fee and priority fee.
func ComputeMaxFee(baseFee *big.Int, priorityFee *big.Int) *big.Int {
	// Max Fee Per Gas = Base Fee + Max Priority Fee
	return new(big.Int).Add(baseFee, priorityFee)
}

// computeTransactionFee calculates the total transaction fee based on gas limit and the maximum fee per gas.
//
// Parameters:
//   - gasLimit: The maximum amount of gas units allowed for the transaction.
//   - maxFee: The maximum fee per gas, calculated as the sum of the base fee and priority fee.
//
// Returns:
//   - *big.Int: The total transaction fee, computed by multiplying the gas limit by the maximum fee per gas.
func computeTransactionFee(gasLimit uint64, maxFee *big.Int) *big.Int {
	// Max Fee Per Gas = Base Fee + Max Priority Fee
	// Gas Fee = Gas Used × (Base Fee + Max Priority Fee)
	gasLimitBigInt := new(big.Int).SetUint64(gasLimit)
	return new(big.Int).Mul(gasLimitBigInt, maxFee)
}
