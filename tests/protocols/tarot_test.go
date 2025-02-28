package protocols

import (
	"defibotgo/internal/protocols/tarot"
	"defibotgo/internal/utils"
	"math/big"
	"testing"
)

func TestComputeReward(t *testing.T) {
	expected := big.NewInt(508132436312)

	vaultPendingReward := big.NewInt(795946798735693857)
	pairValue := big.NewInt(31920000000000)

	rewardToken := tarot.ComputeReward(vaultPendingReward)
	rewardConverted := utils.ConvertToEth(rewardToken, pairValue)
	if rewardConverted.Cmp(expected) != 0 {
		t.Fatalf("rewardToken is incorrect: expecting %v got %v", expected, rewardConverted)
	}
}
