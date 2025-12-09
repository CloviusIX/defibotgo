# DefitBot

DefiBot is a self-initiated project to explore MEV (Maximal Extractable Value) opportunities on DeFi for both learning and practical experience.

## 🔍 Opportunities
- Interacting with Tarot in order to harvest the fee.
- Interacting with Impermax in order to harvest the fee.

## 📋 Prerequisites

- Go (1.23+)
- Make
- Ethereum wallet with private key
- RPC node for used chains
- Docker (optional, for containerized deployment)

## 🚀 Getting Started

### Configuration

Create a file .env in the root directory with the following content:

```
# POOL=USDC_AERO
ACCOUNT_PRIVATE_KEY_TAROT_ONE=<your_wallet_private_key>
ACCOUNT_SENDER_ADDRESS_TAROT_ONE=<your_wallet_address>

# WETH_TAROT
ACCOUNT_PRIVATE_KEY_TAROT_TWO=<your_wallet_private_key>
ACCOUNT_SENDER_ADDRESS_TAROT_TWO=<your_wallet_address>

# POOL=AERO_TAROT
ACCOUNT_PRIVATE_KEY_TAROT_THREE=<your_wallet_private_key>
ACCOUNT_SENDER_ADDRESS_TAROT_THREE=<your_wallet_address>

# POOL=FBOMB_CBBTC
ACCOUNT_PRIVATE_KEY_IMPERMAX_ONE=<your_wallet_private_key>
ACCOUNT_SENDER_ADDRESS_IMPERMAX_ONE=<your_wallet_address>

RPC_NODE_BASE_READ=<your_base_reading_rpc_node>
RPC_NODE_BASE_WRITE=<your_base_writing_rpc_node>
RPC_NODE_OPTIMISM_READ=<your_optimism_reading_rpc_node>
RPC_NODE_OPTIMISM_WRITE=<your_optimism_writing_rpc_node>
```

Each wallet manages a specific pool on a specific chain. Leave any wallet empty if you don’t want to use it. It is recommended to use separate wallets to avoid overlap when two runs are executed simultaneously.

### Setup

Initialize the project and install dependencies:

```
make
```

### Run

Start the application in production mode for the desired pool:

```
make run CHAIN=<chain_name> PROTOCOL=<protocol_name> POOL=<pool_name>
```

Examples:
```

make run CHAIN=base PROTOCOL=tarot POOL=USDC_AERO
make run CHAIN=base PROTOCOL=tarot POOL=WETH_TAROT
make run CHAIN=base PROTOCOL=tarot POOL=AERO_TAROT

make run CHAIN=base PROTOCOL=impermax POOL=FBOMB_CBBTC
```

### Development Mode

Run the application without building for faster development cycles:

```
make rundev CHAIN=<chain_name> PROTOCOL=<protocol_name> POOL=<pool_name>
```

### Code Quality

Run linter and clean up code:

```bash
make lint
```

## 🧪 Testing

### Configuration

Create a file `.env.test` in the root directory with the following content:

```
WALLET_TEST_PRIVATE_KEY=<your_wallet_private_key>
RPC_NODE_BASE_READ=<your_base_reading_rpc_node>
RPC_NODE_BASE_WRITE=<your_base_writing_rpc_node>
RPC_NODE_OPTIMISM_READ=<your_optimism_reading_rpc_node>
RPC_NODE_OPTIMISM_WRITE=<your_optimism_writing_rpc_node>
```

### Run Tests

Execute the test suite:

```
make test
```

## 🐳 Docker

Build and run the application using Docker:

```bash
# Build the Docker image
make docker/build

# Run the container
make docker/run CHAIN=<chain_name> PROTOCOL=<protocol_name> POOL=<pool_name>
```

## 💡 Suggestions

- Run your own RPC node to cut latency and give yourself a better chance in transaction races.
- Prefer WebSockets (WSS) over HTTP for real-time updates and faster block notifications.
- Watch the mempool to spot profitable opportunities early and frontrun competing transactions when needed.
- Send transactions through private relays (e.g. Flashbots) to stay out of the public mempool and avoid being copied or frontrun.


## 📝 Disclaimer

This project is for educational purposes. Use at your own risk.