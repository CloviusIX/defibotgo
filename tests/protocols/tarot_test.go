package protocols

import (
	"context"
	"defibotgo/internal/config"
	"defibotgo/internal/contract_abi"
	"defibotgo/internal/models"
	"defibotgo/internal/protocols/tarot"
	"defibotgo/internal/utils"
	"defibotgo/internal/web3"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"math/big"
	"testing"
)

var reinvestBounty = big.NewInt(20000000000000000)

func TestComputeRewardToEth(t *testing.T) {
	expected := big.NewInt(508132436312)

	vaultPendingReward := big.NewInt(795946798735693857)
	pairValue := big.NewInt(31920000000000)

	rewardToken := tarot.ComputeReward(vaultPendingReward, reinvestBounty)
	rewardConverted := utils.ConvertToEth(rewardToken, pairValue)
	if rewardConverted.Cmp(expected) != 0 {
		t.Fatalf("rewardToken is incorrect: expecting %v got %v", expected, rewardConverted)
	}
}

func TestComputeRewardOnChain(t *testing.T) {
	chain := models.Base
	contractLenderAddress := common.HexToAddress("0x042c37762d1d126bc61eac2f5ceb7a96318f5db9")
	contractGaugeAddress := common.HexToAddress("0x4f09bab2f0e15e2a078a227fe1537665f55b8360")
	rewardExpected := big.NewInt(2448679558346277)

	callOpts := &bind.CallOpts{
		Pending:     false,
		BlockNumber: big.NewInt(27248735),
		Context:     context.Background(),
	}

	ethClient, err := web3.BuildWeb3Client(chain, true)
	if err != nil {
		t.Fatalf("failed to build web3 client: %v", err)
	}
	contractGauge, err := web3.BuildContractInstance(ethClient, contractGaugeAddress, contract_abi.CONTRACT_ABI_GAUGE)
	if err != nil {
		t.Fatalf("failed to build contract instance: %v", err)
	}

	vaultPendingReward, err := web3.EthCall(contractGauge, "earned", callOpts, contractLenderAddress)
	if err != nil {
		t.Fatalf("failed to call earned contract: %v", err)
	}

	rewardToken := tarot.ComputeReward(vaultPendingReward, reinvestBounty)
	if rewardExpected.Cmp(rewardToken) != 0 {
		t.Fatalf("the reward token is incorrect: expecting %v got %v", rewardExpected, rewardToken)
	}
}

func TestGetVaultPendingReward(t *testing.T) {
	//https://basescan.org/tx/0x557f720fe6c3e091f62615e6415c7caaef92eef914b2fc44cbca5e7e156bd31c
	//blockNumber := big.NewInt(29525546)
	balance := big.NewInt(10547979589919134)
	totalSupply := big.NewInt(608561762745652518)
	earned := big.NewInt(891792427871174773)
	rewardRate := big.NewInt(1071909015217126497)
	secondInterval := int64(2)

	expectedVaultPendingReward := big.NewInt(928950445699140388)
	expectedReward := big.NewInt(18579008913982807)

	estimateEarned := tarot.GetVaultPendingReward(earned, rewardRate, secondInterval, balance, totalSupply)
	estimateReward := tarot.ComputeReward(estimateEarned, reinvestBounty)
	if estimateEarned.Cmp(expectedVaultPendingReward) != 0 {
		t.Fatalf("the estimation of Earned is inccorect: expecting %v got %v", expectedVaultPendingReward, estimateEarned)
	}

	if estimateReward.Cmp(expectedReward) != 0 {
		t.Fatalf("the estimation of reward is incorrect: expecting %v got %v", expectedReward, estimateReward)
	}
}

func TestGetTransactionGasFees(t *testing.T) {
	chain := models.Base
	priorityFeeIncreasePercent := 0
	gasLimitExtraPercent := uint64(0)

	transactionFeeExpected := big.NewInt(813970887184)
	gasLimitExpected := uint64(426244)
	gasFeeExpected := big.NewInt(1909636)
	gasTipExpected := big.NewInt(5678)

	protocolOpts := &models.TarotOpts{
		ReinvestBounty: reinvestBounty,
		PriorityFee:    big.NewInt(5678),
		Sender:         common.HexToAddress(config.GetSecret(config.WalletTestPrivateKey)),
		Chain:          chain,
		ContractLender: common.HexToAddress("0x042c37762d1d126bc61eac2f5ceb7a96318f5db9"),
		ContractGauge:  common.HexToAddress("0x4f09bab2f0e15e2a078a227fe1537665f55b8360"),
	}

	//https://basescan.org/tx/0x93efd0f572de355f5cd34120af45360cc1d22765df8ae7fe91528ff2801b210b
	tarotCalculationOpts := &tarot.TarotCalculationOpts{}
	tarotCalculationOpts.VaultPendingRewardValue = big.NewInt(276513852697572252)
	tarotCalculationOpts.RewardPairValue = big.NewInt(269300000000000)
	tarotCalculationOpts.BaseFeeValue = big.NewInt(1903958)
	tarotCalculationOpts.EstimateGasLimitValue = 426244
	tarotCalculationOpts.PriorityFeeValue = big.NewInt(5678)

	isWorth, gasOpts, _, err := tarot.GetL2TransactionGasFees(
		protocolOpts,
		tarotCalculationOpts,
		priorityFeeIncreasePercent,
		gasLimitExtraPercent,
	)

	if err != nil {
		t.Fatalf("Error getting protocol gas fees: %v", err)
	}

	if !isWorth {
		t.Fatalf("the transaction is not worthy")
	}

	if gasOpts.TransactionFee.Cmp(transactionFeeExpected) != 0 {
		t.Fatalf("the transaction fee is incorrect: expecting %v got %v", transactionFeeExpected, gasOpts.TransactionFee)
	}

	if gasOpts.GasLimit != gasLimitExpected {
		t.Fatalf("the gas limit is incorrect: expecting %v got %v", gasLimitExpected, gasOpts.GasLimit)
	}

	if gasOpts.GasFeeCap.Cmp(gasFeeExpected) != 0 {
		t.Fatalf("the gas fee is incorrect: expecting %v got %v", gasFeeExpected, gasOpts.GasFeeCap)
	}

	if gasOpts.GasTipCap.Cmp(gasTipExpected) != 0 {
		t.Fatalf("the priority fee is incorrect: expecting %v got %v", big.NewInt(461678), gasOpts.GasTipCap)
	}

}
