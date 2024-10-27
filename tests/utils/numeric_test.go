package utils

import (
	"defibotgo/internal/utils"
	"math/big"
	"testing"
)

func TestParseWeiString(t *testing.T) {
	valueStr := "0.00003182"
	expected := big.NewInt(31820000000000)

	resultWei, err := utils.ParseWeiString(valueStr)
	if err != nil {
		t.Fatalf("failed to parse wei string: %v", err)
	}

	if resultWei.Cmp(expected) != 0 {
		t.Fatalf("failed to parse wei string. expected: %v, got: %v", expected, resultWei)
	}
}
