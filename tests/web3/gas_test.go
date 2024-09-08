package web3

import (
	"defibotgo/internal/web3"
	"math/big"
	"testing"
)

func TestBuildTransactionFeeArgs(t *testing.T) {
	baseFee := big.NewInt(57143102)
	gasLimit := uint64(1090381)
	priorityFee := big.NewInt(323762)

	transactionFeeExpected := big.NewInt(62660776635184)

	gasOpts := web3.BuildTransactionFeeArgs(baseFee, priorityFee, gasLimit)

	if gasOpts.TransactionFee.Cmp(transactionFeeExpected) != 0 {
		t.Fatalf("gasOpts.TransactionFee: expected %v, got %v", transactionFeeExpected, gasOpts.TransactionFee)
	}

	if gasOpts.GasLimit != gasLimit && gasOpts.GasFeeCap.Cmp(baseFee) != 0 && gasOpts.GasTipCap.Cmp(priorityFee) != 0 {
		t.Fatal("gasOpts is different from expected")
	}
}
