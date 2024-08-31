package abi

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
