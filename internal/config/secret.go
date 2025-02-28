package config

import (
	"github.com/joho/godotenv"
	"log"
	"os"
	"sync"
)

// SecretKey is an enum-like type to represent secret keys
type SecretKey int

const (
	RpcNodeBaseReadKey SecretKey = iota
	RpcNodeBaseWriteKey
	RpcNodeOptimismReadKey
	RpcNodeOptimismWriteKey
	WalletTarotKeyOne
	WalletTarotAddressOne
	WalletTestPrivateKey
)

var (
	secrets        map[SecretKey]string
	secretsStorage sync.Once // Ensures secrets are loaded only once
)

// loadSecrets initializes the secrets map with environment variables based on APP_ENV.
func loadSecrets() {
	env := os.Getenv("APP_ENV")
	loadEnvFile(env)

	// Configure secrets based on the environment
	switch env {
	case "test":
		secrets = map[SecretKey]string{
			RpcNodeBaseReadKey:      getEnvOrFatal("RPC_NODE_BASE_READ"),
			RpcNodeBaseWriteKey:     getEnvOrFatal("RPC_NODE_BASE_WRITE"),
			RpcNodeOptimismReadKey:  getEnvOrFatal("RPC_NODE_OPTIMISM_READ"),
			RpcNodeOptimismWriteKey: getEnvOrFatal("RPC_NODE_OPTIMISM_WRITE"),
			WalletTestPrivateKey:    getEnvOrFatal("WALLET_TEST_PRIVATE_KEY"),
		}
	default:
		secrets = map[SecretKey]string{
			RpcNodeBaseReadKey:      getEnvOrFatal("RPC_NODE_BASE_READ"),
			RpcNodeBaseWriteKey:     getEnvOrFatal("RPC_NODE_BASE_WRITE"),
			RpcNodeOptimismReadKey:  getEnvOrFatal("RPC_NODE_OPTIMISM_READ"),
			RpcNodeOptimismWriteKey: getEnvOrFatal("RPC_NODE_OPTIMISM_WRITE"),
			WalletTarotKeyOne:       getEnvOrFatal("ACCOUNT_PRIVATE_KEY_TAROT_ONE_0XB8"),
			WalletTarotAddressOne:   getEnvOrFatal("ACCOUNT_SENDER_ADDRESS_TAROT_ONE_0XB8"),
		}
	}
}

// loadEnvFile loads the appropriate .env file based on the APP_ENV variable.
func loadEnvFile(env string) {
	var envFile string

	switch env {
	case "test":
		envFile = "../../.env.test"
	case "development":
		envFile = ".env"
	case "production":
		log.Println("Production environment detected; skipping .env file loading.")
		return
	default:
		log.Printf("Unknown APP_ENV '%s'. Using .env by default.", env)
		envFile = ".env"
	}

	// Load the selected .env file
	if err := godotenv.Load(envFile); err != nil {
		log.Printf("Info: %s file not found or could not be loaded. Proceeding with environment variables.", envFile)
	}
}

// getEnvOrFatal retrieves an environment variable or exits if it’s not set.
func getEnvOrFatal(key string) string {
	value, exists := os.LookupEnv(key)
	if !exists {
		log.Fatalf("Environment variable %s is required but not set.", key)
	}
	return value
}

// GetSecret retrieves a secret by key
func GetSecret(key SecretKey) string {
	// Ensure secrets are loaded only once
	secretsStorage.Do(loadSecrets)

	secret, exists := secrets[key]
	if !exists {
		log.Fatalf("secret not found for key: %v", key)
	}
	return secret
}
