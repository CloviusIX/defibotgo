package protocols

import (
	"defibotgo/internal/protocols/tarot"
	"log"
	"math/big"
	"testing"
)

func TestComputeReward(t *testing.T) {
	expected := big.NewInt(508132436312)

	vaultPendingReward := big.NewInt(795946798735693857)
	pairValue := big.NewInt(31920000000000)

	reward := tarot.ComputeReward(vaultPendingReward, pairValue)
	if reward.Cmp(expected) != 0 {
		t.Fatalf("reward is incorrect: expecting %v got %v", expected, reward)
	}
	log.Println(reward)
}
