# Beacon Header Verification Tool

A tool for verifying Ethereum beacon block headers by generating SSZ Merkle proofs for header fields and optionally verifying them on-chain using a smart contract.

---

## Overview

The Beacon Header Verification Tool fetches beacon block header data from configured API endpoints, generates Merkle proofs for selected header fields, and (if configured) performs on-chain verification using an Ethereum node connection. The project is split into two main components:

- **Go Application**: Implements the logic for fetching, processing, and verifying beacon block headers.
- **Solidity Smart Contract**: Contains the `BeaconHeaderVerifier` contract used for on-chain verification of header fields via Merkle proofs.

---

## Prerequisites

- **Go**: Install [Go](https://golang.org/dl/) (version 1.16 or higher recommended).
- **Ethereum Node**: An Ethereum node endpoint is required for on-chain verification (e.g., [Infura](https://infura.io/), [Alchemy](https://alchemy.com/), or your own node).
- **Beacon API Endpoints**: At least one Beacon API endpoint to fetch beacon block header data.
- **Foundry**: Install [Foundry](https://github.com/foundry-rs/foundry) for smart contract development, testing, and deployment.
- **Make**: GNU Make is used for building, testing, and running the application.

---

## Installation

1. **Clone the Repository:**

   ```bash
   git clone https://github.com/Mikelle/beacon-root-verification.git
   cd beacon-root-verification/beacon-verifier
   ```

2. **Download Dependencies:**

    ```bash
    go mod tidy
    ```

3. **Build the Application:**

    Use the Makefile to build the binary:

    ```bash
    make build
    ```

    This will compile the application and place the binary in the bin/ directory.

## Configuration

The application requires configuration details provided either via a configuration file or environment variables (as defined in the project’s `config` package). Key configuration parameters include:

- **Beacon API Endpoints**: At least one URL from which to fetch beacon block header data.
- **Ethereum Node Endpoint**: URL for connecting to an Ethereum node for on-chain verification.
- **Slot**: (Optional) Specific beacon block slot to verify. If omitted, the application fetches the latest header and its predecessor.
- **Verification Fields**: List of header fields (e.g., `slot`, `proposer_index`, `parent_root`, `state_root`, `body_root`) to generate and verify proofs for.
- **Retry Attempts**: Number of retry attempts for fetching beacon block headers if a particular slot is unavailable.

Ensure that your configuration adheres to the expected schema.

---

## Running the Application

### Using the Makefile

To build and run the application with a specific slot, execute:

```bash
make run SLOT=1234567
```

Replace 1234567 with the desired slot number. If no slot is specified, the application will automatically fetch the latest beacon block header and its predecessor.

### Running the Binary Directly

After building, you can run the binary directly:

```bash
./bin/beacon-verifier -slot 1234567
```

or without specifying a slot:

```bash
./bin/beacon-verifier
```

The application logs detailed information about the header data, generated Merkle proofs, and verification results.

## Testing and Code Quality

The provided Makefile includes targets for testing, formatting, and linting:

- **Run Tests:**

  ```bash
  make test
  ```

  This command runs all the unit tests in verbose mode.
- **Format Code:**

  ```bash
  make fmt
  ```

  This command formats the Go code using gofmt.

- **Clean Build Artifacts:**

  ```bash
  make clean
  ```

- **List All Make Targets:**

  ```bash
  make help
  ```

## Smart Contract Verification

The Solidity smart contract `BeaconHeaderVerifier` is used to verify beacon block header fields on-chain. Its key functionalities include:

- **Retrieving the Beacon Block Root**: Uses the Beacon Roots contract (as per EIP-4788) to fetch the root.
- **Field Verification**: Verifies individual header fields (e.g., `slot`, `proposer index`, `parent root`, `state root`, `body root`) using SSZ Merkle proofs.

### Deployed contract

### Deployed contract

The `BeaconHeaderVerifier` contract is deployed on the Holesky Ethereum testnet. You can view the contract on Etherscan using the following link:

[BeaconHeaderVerifier on Holesky Etherscan](https://holesky.etherscan.io/address/0x4D581D208fe2645A97Bee8344c5073c6729a715b#code)

### Run tests

```bash
cd contracts
forge test
```
