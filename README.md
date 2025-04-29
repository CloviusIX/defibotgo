# DefitBot

DefiBot is a self-project to explore DeFi (Decentralized Finance) opportunities for learning and practical experience. This tool helps you to catch the next DeFi opportunity.

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
ACCOUNT_PRIVATE_KEY_TAROT_ONE_0XB8=<your_wallet_private_key>
ACCOUNT_SENDER_ADDRESS_TAROT_ONE_0XB8=<your_wallet_address>
ACCOUNT_PRIVATE_KEY_IMPERMAX_ONE_0XAE=<your_wallet_private_key>
ACCOUNT_SENDER_ADDRESS_IMPERMAX_ONE_0XAE=<your_wallet_address>
RPC_NODE_OPTIMISM_READ=<your_optimism_reading_rpc_node>
RPC_NODE_OPTIMISM_WRITE=<your_optimism_writing_rpc_node>
RPC_NODE_BASE_READ=<your_base_reading_rpc_node>
RPC_NODE_BASE_WRITE=<your_base_writing_rpc_node>
```

### Setup

Initialize the project and install dependencies:

```
make
```

### Run

Start the application in production mode:

```
make run
```

### Development Mode

Run the application without building for faster development cycles:

```
make rundev
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
make docker/run
```

## 📝 License

This project is for educational purposes. Use at your own risk.