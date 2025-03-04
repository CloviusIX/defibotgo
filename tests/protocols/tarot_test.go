package protocols

import (
	"context"
	"defibotgo/internal/abi"
	"defibotgo/internal/config"
	"defibotgo/internal/models"
	"defibotgo/internal/protocols/tarot"
	"defibotgo/internal/utils"
	"defibotgo/internal/web3"
	"github.com/dgraph-io/ristretto"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"log"
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

func TestGetTransactionGasFees(t *testing.T) {
	chain := models.Base
	reinvestFunctionName := "reinvest"
	priorityFeeIncreasePercent := 5
	transactionFeeExpected := big.NewInt(1293073197052)

	vaultPendingRewardChan := make(chan models.WeiResult, 1)
	baseFeePerGasChan := make(chan models.WeiResult, 1)
	estimateGasChan := make(chan models.GasLimitResult, 1)
	rewardPairValueChan := make(chan models.WeiResult, 1)
	priorityFeeChan := make(chan models.WeiResult, 1)

	ethClient, err := web3.BuildWeb3Client(chain, true)
	if err != nil {
		log.Fatalf("Error building eth cient: %s", err)
	}

	protocolOpts := &models.TarotOpts{
		Chain:            chain,
		Sender:           common.HexToAddress(config.GetSecret(config.WalletTestPrivateKey)),
		PriorityFee:      big.NewInt(5678),
		BlockRangeFilter: big.NewInt(20),
		ContractLender:   common.HexToAddress("0x042c37762d1d126bc61eac2f5ceb7a96318f5db9"),
		ContractGauge:    common.HexToAddress("0x4f09bab2f0e15e2a078a227fe1537665f55b8360"),
	}

	//https://basescan.org/tx/0x93efd0f572de355f5cd34120af45360cc1d22765df8ae7fe91528ff2801b210b
	blockOpts := &bind.CallOpts{
		Pending:     false,
		BlockNumber: big.NewInt(27149476),
		Context:     context.Background(),
	}

	contractGauge, err := web3.BuildContractInstance(ethClient, protocolOpts.ContractGauge, abi.CONTRACT_ABI_GAUGE)
	if err != nil {
		log.Fatalf("error building tarot contract gauge instance on %s: %v", protocolOpts.ContractGauge, err)
	}

	abiJson, err := web3.LoadAbi(abi.CONTRACT_ABI_LENDER)
	if err != nil {
		log.Fatalf("error loading Tarot abi on %s: %v", protocolOpts.Chain, err)
	}

	data, err := abiJson.Pack(reinvestFunctionName)
	if err != nil {
		log.Fatalf("failed to pack Tarot abi on %s: %v", protocolOpts.Chain, err)
	}

	// Create a message to simulate the transaction
	callMsg := ethereum.CallMsg{
		From:  protocolOpts.Sender,
		To:    &protocolOpts.ContractLender,
		Data:  data, // ABI-encoded function call data
		Value: big.NewInt(0),
	}
	log.Printf("calling contract lender on %s", protocolOpts.ContractLender.Hex())

	cache, err := ristretto.NewCache(&ristretto.Config{
		NumCounters: 50,  // 10x the number of items we expect to store
		MaxCost:     320, // Approx. 64 bytes per *big.Int, 5 keys in total
		BufferItems: 64,  // Recommended size for eviction buffer
	})
	if err != nil {
		panic(err)
	}

	_, gasOpts, err := tarot.GetTransactionGasFees(
		ethClient,
		contractGauge,
		callMsg,
		protocolOpts,
		blockOpts,
		priorityFeeIncreasePercent,
		cache,
		vaultPendingRewardChan,
		baseFeePerGasChan,
		estimateGasChan,
		rewardPairValueChan,
		priorityFeeChan,
	)

	if err != nil {
		t.Fatalf("Error getting protocol gas fees: %v", err)
	}
	if gasOpts.TransactionFee.Cmp(transactionFeeExpected) != 0 {
		t.Fatalf("the transaction fee is incorrect: expecting %v got %v", transactionFeeExpected, gasOpts.TransactionFee)
	}

	if gasOpts.GasTipCap.Cmp(big.NewInt(461678)) != 0 {
		t.Fatalf("the priority fee is incrorrect: expecting %v got %v", big.NewInt(461678), gasOpts.GasTipCap)
	}
	log.Printf("gas fee is %v", gasOpts)
}
