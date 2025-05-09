package config

import (
	"github.com/joho/godotenv"
	"github.com/rs/zerolog/log"
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
	WalletTarotAddressTwo
	WalletTarotKeyTwo
	WalletTarotAddressThree
	WalletTarotKeyThree
	WalletTestPrivateKey
	WalletImpermaxKeyOne
	WalletImpermaxAddressOne
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
			RpcNodeBaseReadKey:       getEnvOrFatal("RPC_NODE_BASE_READ"),
			RpcNodeBaseWriteKey:      getEnvOrFatal("RPC_NODE_BASE_WRITE"),
			RpcNodeOptimismReadKey:   getEnvOrFatal("RPC_NODE_OPTIMISM_READ"),
			RpcNodeOptimismWriteKey:  getEnvOrFatal("RPC_NODE_OPTIMISM_WRITE"),
			WalletTarotKeyOne:        getEnvOrFatal("ACCOUNT_PRIVATE_KEY_TAROT_ONE"),
			WalletTarotAddressOne:    getEnvOrFatal("ACCOUNT_SENDER_ADDRESS_TAROT_ONE"),
			WalletTarotKeyTwo:        getEnvOrFatal("ACCOUNT_PRIVATE_KEY_TAROT_TWO"),
			WalletTarotAddressTwo:    getEnvOrFatal("ACCOUNT_SENDER_ADDRESS_TAROT_TWO"),
			WalletTarotAddressThree:  getEnvOrFatal("ACCOUNT_SENDER_ADDRESS_TAROT_THREE"),
			WalletTarotKeyThree:      getEnvOrFatal("ACCOUNT_PRIVATE_KEY_TAROT_THREE"),
			WalletImpermaxKeyOne:     getEnvOrFatal("ACCOUNT_PRIVATE_KEY_IMPERMAX_ONE"),
			WalletImpermaxAddressOne: getEnvOrFatal("ACCOUNT_SENDER_ADDRESS_IMPERMAX_ONE"),
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
		log.Info().Msg("Production environment detected; skipping .env file loading.")
		return
	default:
		log.Info().Str("env file", env).Msg("Unknown APP_ENV. Using .env by default.")
		envFile = ".env"
	}

	// Load the selected .env file
	if err := godotenv.Load(envFile); err != nil {
		log.Info().Str("env file", envFile).Msg("env file not found or could not be loaded. Proceeding with environment variables.")
	}
}

// getEnvOrFatal retrieves an environment variable or exits if it’s not set.
func getEnvOrFatal(key string) string {
	value, exists := os.LookupEnv(key)
	if !exists {
		log.Fatal().Str("key", key).Msg("Environment variable is required but not set")
	}
	return value
}

// GetSecret retrieves a secret by key
func GetSecret(key SecretKey) string {
	// Ensure secrets are loaded only once
	secretsStorage.Do(loadSecrets)

	secret, exists := secrets[key]
	if !exists {
		log.Fatal().Int("key", int(key)).Msg("secret not found for key")
	}
	return secret
}
