package utils

import (
	"defibotgo/internal/utils"
	"math/big"
	"testing"
)

func TestComputeDifference(t *testing.T) {
	value1 := big.NewInt(49886086613922)
	value2 := big.NewInt(62313943885180)
	expected := -19.943942713941578

	result := utils.ComputeDifference(value1, value2)
	if result != expected {
		t.Errorf("ComputeDifference(%v, %v) returned %v, expected %v", value1, value2, result, expected)
	}
}
