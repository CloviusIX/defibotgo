package web3

import (
	"context"
	"defibotgo/internal/config"
	"defibotgo/internal/contract_abi"
	"defibotgo/internal/models"
	"defibotgo/internal/utils"
	"defibotgo/internal/web3"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"math/big"
	"testing"
)

func TestGetEstimateL1Fee(t *testing.T) {
	ctx, ctxCancel := context.WithCancel(context.Background())
	defer ctxCancel()

	reinvestFunctionName := "reinvest"
	l1FeeExpected := big.NewInt(16268490247)
	differenceExpected := float64(-11.507087398437047)

	ethClient, err := web3.BuildWeb3Client(models.Base, true)
	if err != nil {
		t.Fatalf("failed to build eth client err")
	}

	gasOracle := common.HexToAddress("0x420000000000000000000000000000000000000F")
	lenderAddress := common.HexToAddress("0x042c37762d1d126bc61eac2f5ceb7a96318f5db9")
	callOpt := &bind.CallOpts{
		Pending:     false,
		BlockNumber: big.NewInt(29004389),
		Context:     ctx,
	}

	gasOpts := &web3.GasOpts{
		GasFeeCap: big.NewInt(3116168),
		GasTipCap: big.NewInt(556962),
		GasLimit:  413043,
	}

	chainId, err := ethClient.ChainID(ctx)
	if err != nil {
		panic(err)
	}

	contractGasOracle, err := web3.BuildContractInstance(ethClient, gasOracle, contract_abi.CONTRACT_ABI_GAS_PRICE_ORACLE)
	if err != nil {
		t.Fatalf("failed to build contract instance: %v", err)
	}

	lenderAbiJson, err := web3.LoadAbi(contract_abi.CONTRACT_ABI_LENDER)
	if err != nil {
		t.Fatalf("failed to load contract abi: %v", err)
	}

	lenderData, err := lenderAbiJson.Pack(reinvestFunctionName)
	if err != nil {
		t.Fatalf("failed to load lender abi: %v", err)
	}

	walletPrivateKey := config.GetSecret(config.WalletTestPrivateKey)
	walletPrivateKeyCiph, errCiph := crypto.HexToECDSA(walletPrivateKey)
	if errCiph != nil {
		t.Fatalf("Failed to build Base contract L1 Fee instance: %v", errCiph)
	}

	l1Fee, _, err := web3.GetL1GasFee(ctx, ethClient, chainId, callOpt, gasOpts, contractGasOracle, &lenderAddress, lenderData, walletPrivateKeyCiph)
	diff := utils.ComputeDifference(l1FeeExpected, l1Fee)

	if err != nil {
		t.Fatalf("failed to get estimate l1 fee: %v", err)
	}

	if diff != differenceExpected {
		t.Fatalf("estimateL1Fee failed, expected %v, got %v with a difference too big %v", l1FeeExpected, l1Fee, diff)
	}
}
