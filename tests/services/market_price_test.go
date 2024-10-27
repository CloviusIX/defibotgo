package services

import (
	"defibotgo/internal/models"
	"defibotgo/internal/services"
	"math/big"
	"testing"
)

func TestPoolPriceApi(t *testing.T) {
	chain := models.Optimism
	pairPrice, err := services.GetPoolPrice(chain)

	if err != nil {
		t.Fatalf("Error getting pool value: %v", err)
	}

	if pairPrice.Cmp(big.NewInt(0)) != 1 {
		t.Fatalf("GetPoolPrice should be greater than 0: got %v", pairPrice)
	}
}
