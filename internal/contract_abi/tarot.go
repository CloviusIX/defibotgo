package contract_abi

// CONTRACT_ABI_LENDER is the ABI definition for the lender contract
const CONTRACT_ABI_LENDER = `[
    {
        "constant": false,
        "inputs": [],
        "name": "reinvest",
        "outputs": [],
        "payable": false,
        "stateMutability": "nonpayable",
        "type": "function"
    }
]`

// CONTRACT_ABI_GAUGE is the ABI definition for the gauge contract
const CONTRACT_ABI_GAUGE = `[
    {
        "inputs": [{"internalType": "address", "name": "_account", "type": "address"}],
        "name": "earned",
        "outputs": [{"internalType": "uint256", "name": "", "type": "uint256"}],
        "stateMutability": "view",
        "type": "function"
    }
]`

// CONTRACT_ABI_GAS_PRICE_ORACLE is the ABI definition for the gas price oracle contract
const CONTRACT_ABI_GAS_PRICE_ORACLE = `[
	{
	  "inputs": [
		{ "internalType": "bytes", "name": "_data", "type": "bytes" }
	  ],
	  "name": "getL1Fee",
	  "outputs": [
		{ "internalType": "uint256", "name": "fee", "type": "uint256" }
	  ],
	  "stateMutability": "view",
	  "type": "function"
	}
]`
