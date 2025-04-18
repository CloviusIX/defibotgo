package web3

import (
	"context"
	"crypto/ecdsa"
	"defibotgo/internal/contract_abi"
	"fmt"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"math/big"
	"strings"
)

func GetL1GasFee(
	ctx context.Context,
	ethClient *ethclient.Client,
	chainId *big.Int,
	callOpts *bind.CallOpts,
	gasOpts *GasOpts,
	contractGasPriceOracle *bind.BoundContract,
	lenderAddress *common.Address,
	functionName string,
	walletPrivateKey *ecdsa.PrivateKey,
) (*big.Int, *types.Transaction, error) {
	// TODO set set chain though param ?
	transactionOpts, err := bind.NewKeyedTransactorWithChainID(walletPrivateKey, chainId)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create transaction options: %v", err)
	}

	// 2) Pack the call data for your function
	// TODO: get the abi once and pass it though param
	// TODO:uUse abigen ?
	parsedABI, err := abi.JSON(strings.NewReader(contract_abi.CONTRACT_ABI_LENDER))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse ABI: %v", err)
	}

	callData, err := parsedABI.Pack(functionName)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate call data: %v", err)
	}

	// Fetch nonce
	nonce, err := ethClient.PendingNonceAt(ctx, transactionOpts.From)
	if err != nil {
		return nil, nil, err
	}

	// Build an EIP‑1559 tx (not yet signed)
	tx := types.NewTx(&types.DynamicFeeTx{
		ChainID:   chainId,
		Nonce:     nonce,
		To:        lenderAddress,
		Data:      callData,
		Gas:       gasOpts.GasLimit,
		GasTipCap: gasOpts.GasTipCap,
		GasFeeCap: gasOpts.GasFeeCap,
	})

	// Sign it locally
	signer := types.LatestSignerForChainID(chainId)
	signedTx, err := types.SignTx(tx, signer, walletPrivateKey)
	if err != nil {
		return nil, nil, fmt.Errorf("sign tx: %w", err)
	}

	// RLP‑encode the signed tx and call the oracle
	rlpBytes, err := signedTx.MarshalBinary()
	if err != nil {
		return nil, nil, fmt.Errorf("rlp encode: %w", err)
	}

	l1Fee, err := EthCall(contractGasPriceOracle, "getL1Fee", callOpts, rlpBytes)
	//l1Fee, err := contractGasPriceOracle.GetL1GasFee(&bind.CallOpts{Context: ctx}, rlpBytes)
	if err != nil {
		return nil, nil, fmt.Errorf("oracle.GetL1GasFee: %w", err)
	}

	return l1Fee, signedTx, nil
}
