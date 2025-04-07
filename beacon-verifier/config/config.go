// Package config provides configuration structures and loading functionality
package config

import (
	"flag"
)

// Config represents the application configuration
type Config struct {
	BeaconAPI    BeaconAPIConfig    `json:"beacon_api"`
	Verification VerificationConfig `json:"verification"`
	EthereumNode EthereumNodeConfig `json:"ethereum_node"`
	Slot         string             `json:"slot"`
}

// BeaconAPIConfig contains beacon chain API configuration
type BeaconAPIConfig struct {
	Endpoints        []string `json:"endpoints"`
	RetryAttempts    int      `json:"retry_attempts"`
	RequestTimeoutMs int      `json:"request_timeout_ms"`
}

// VerificationConfig contains verification-related settings
type VerificationConfig struct {
	VerifierAddress      string   `json:"verifier_address"`
	FieldsToVerify       []string `json:"fields_to_verify"`
	MaxVerificationSlots int      `json:"max_verification_slots"`
}

// EthereumNodeConfig contains Ethereum node configuration
type EthereumNodeConfig struct {
	Endpoint string `json:"endpoint"`
	ChainID  int    `json:"chain_id"`
}

// DefaultConfig returns a configuration with sensible defaults
func DefaultConfig() *Config {
	return &Config{
		BeaconAPI: BeaconAPIConfig{
			Endpoints: []string{
				"https://responsive-spring-glitter.ethereum-holesky.quiknode.pro/cb7d99ff3abcf116b14f8255e084e0a0121ad9e1",
			},
			RetryAttempts:    5,
			RequestTimeoutMs: 5000,
		},
		Verification: VerificationConfig{
			VerifierAddress: "0x4D581D208fe2645A97Bee8344c5073c6729a715b",
			FieldsToVerify: []string{
				"slot",
				"proposer_index",
				"parent_root",
				"state_root",
				"body_root",
			},
			MaxVerificationSlots: 5,
		},
		EthereumNode: EthereumNodeConfig{
			Endpoint: "",    // Default to using the same endpoint as Beacon API
			ChainID:  17000, // Holesky testnet
		},
	}
}

// LoadConfig loads configuration from command line flags
func LoadConfig() (*Config, error) {
	config := DefaultConfig()

	// Define command line flags
	beaconEndpoint := flag.String("beacon", "", "Beacon chain API endpoint")
	verifierAddr := flag.String("verifier", "", "Beacon header verifier contract address")
	ethEndpoint := flag.String("eth", "", "Ethereum node endpoint")
	maxRetries := flag.Int("retries", 0, "Maximum number of retry attempts")
	slotToVerify := flag.String("slot", "", "Specific slot to verify (defaults to auto-detecting a recent slot)")
	flag.Parse()

	// Override with command line parameters if provided
	if *beaconEndpoint != "" {
		config.BeaconAPI.Endpoints = []string{*beaconEndpoint}
	}

	if *verifierAddr != "" {
		config.Verification.VerifierAddress = *verifierAddr
	}

	if *ethEndpoint != "" {
		config.EthereumNode.Endpoint = *ethEndpoint
	}

	if *maxRetries > 0 {
		config.BeaconAPI.RetryAttempts = *maxRetries
	}

	if *slotToVerify != "" {
		config.Slot = *slotToVerify
	}

	// If Ethereum endpoint not specified, use the first beacon API endpoint
	if config.EthereumNode.Endpoint == "" && len(config.BeaconAPI.Endpoints) > 0 {
		config.EthereumNode.Endpoint = config.BeaconAPI.Endpoints[0]
	}

	return config, nil
}
