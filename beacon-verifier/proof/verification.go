// Package proof provides functionality for generating and verifying Merkle proofs
package proof

import (
	"bytes"
	"context"
	"encoding/hex"
	"fmt"
	"log"
	"math/big"

	"github.com/Mikelle/beacon-root-verification/beacon-verifier/beacon"
	"github.com/Mikelle/beacon-root-verification/beacon-verifier/merkle"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

// Data represents the data for a Merkle proof
type Data struct {
	BeaconTimestamp int64    `json:"beaconTimestamp"`
	BeaconBlockRoot string   `json:"beaconBlockRoot"`
	FieldIndex      int      `json:"fieldIndex"`
	FieldValue      string   `json:"fieldValue"`
	MerkleProof     []string `json:"merkleProof"`
}

// FieldNames maps field names to their indices
var FieldNames = map[string]int{
	"slot":           0,
	"proposer_index": 1,
	"parent_root":    2,
	"state_root":     3,
	"body_root":      4,
}

// BeaconHeaderVerifierABI contains the minimal ABI for the verifyHeaderField function
const BeaconHeaderVerifierABI = `[
  {
    "inputs": [
      {"internalType": "uint256", "name": "beaconTimestamp", "type": "uint256"},
      {"internalType": "uint8", "name": "fieldIndex", "type": "uint8"},
      {"internalType": "bytes32", "name": "expectedValue", "type": "bytes32"},
      {"internalType": "bytes32[]", "name": "merkleProof", "type": "bytes32[]"}
    ],
    "name": "verifyHeaderField",
    "outputs": [
      {"internalType": "bool", "name": "", "type": "bool"}
    ],
    "stateMutability": "view",
    "type": "function"
  }
]`

// GenerateHeaderProof generates a Merkle proof for a specific field in a beacon block header
func GenerateHeaderProof(headerData beacon.HeaderData, fieldName string, nextSlotTimestamp int64) (Data, error) {
	var header beacon.BlockHeader
	if err := header.FromAPIResponse(headerData); err != nil {
		return Data{}, fmt.Errorf("error processing header data: %w", err)
	}

	fieldIndex, exists := FieldNames[fieldName]
	if !exists {
		return Data{}, fmt.Errorf("unknown field name: %s. Must be one of %v", fieldName, getMapKeys(FieldNames))
	}

	serializedFields := header.SerializeForMerkleization()

	// Create Merkle tree
	tree, err := merkle.NewTree(serializedFields)
	if err != nil {
		return Data{}, fmt.Errorf("error creating Merkle tree: %w", err)
	}

	merkleProof, err := tree.ComputeProof(fieldIndex)
	if err != nil {
		return Data{}, fmt.Errorf("error computing Merkle proof: %w", err)
	}

	// Get the field value
	var fieldValueBytes []byte
	if fieldName == "slot" || fieldName == "proposer_index" {
		fieldValueBytes = serializedFields[fieldIndex]
		// For numeric fields, also show the decoded value
		if fieldName == "slot" {
			value := uint64(0)
			for i := 0; i < 8; i++ {
				value |= uint64(fieldValueBytes[i]) << (i * 8)
			}
			log.Printf("Slot value (decoded): %d", value)
		} else if fieldName == "proposer_index" {
			value := uint64(0)
			for i := 0; i < 8; i++ {
				value |= uint64(fieldValueBytes[i]) << (i * 8)
			}
			log.Printf("Proposer index (decoded): %d", value)
		}
	} else {
		switch fieldName {
		case "parent_root":
			fieldValueBytes = header.ParentRoot
		case "state_root":
			fieldValueBytes = header.StateRoot
		case "body_root":
			fieldValueBytes = header.BodyRoot
		}
	}

	// Convert proof nodes to hex strings
	proofHexStrings := make([]string, len(merkleProof))
	for i, node := range merkleProof {
		proofHexStrings[i] = "0x" + hex.EncodeToString(node)
	}

	proofData := Data{
		BeaconTimestamp: nextSlotTimestamp,
		BeaconBlockRoot: "0x" + hex.EncodeToString(tree.Root()),
		FieldIndex:      fieldIndex,
		FieldValue:      "0x" + hex.EncodeToString(fieldValueBytes),
		MerkleProof:     proofHexStrings,
	}

	log.Printf("Generated proof for field '%s' (index %d)", fieldName, fieldIndex)
	truncatedValue := proofData.FieldValue
	if len(truncatedValue) > 20 {
		truncatedValue = truncatedValue[:20] + "..."
	}
	log.Printf("Field value: %s", truncatedValue)

	truncatedRoot := proofData.BeaconBlockRoot
	if len(truncatedRoot) > 20 {
		truncatedRoot = truncatedRoot[:20] + "..."
	}
	log.Printf("Header root: %s", truncatedRoot)

	return proofData, nil
}

// VerifyOnChain uses Web3 to call the onchain BeaconHeaderVerifier contract
func VerifyOnChain(client *ethclient.Client, contractAddress string, proofData Data) (bool, error) {
	parsedABI, err := abi.JSON(bytes.NewReader([]byte(BeaconHeaderVerifierABI)))
	if err != nil {
		return false, fmt.Errorf("error parsing ABI: %w", err)
	}

	address := common.HexToAddress(contractAddress)

	// Prepare call parameters
	beaconTimestamp := big.NewInt(proofData.BeaconTimestamp)
	fieldIndex := uint8(proofData.FieldIndex)

	// Convert field value from hex string to bytes32
	fieldValueHex := trimHexPrefix(proofData.FieldValue)
	fieldValueBytes, err := hex.DecodeString(fieldValueHex)
	if err != nil {
		return false, fmt.Errorf("error decoding field value: %w", err)
	}
	var fieldValue [32]byte
	copy(fieldValue[:], fieldValueBytes)

	// Convert merkle proof from hex strings to bytes32 array
	merkleProofBytes := make([][32]byte, len(proofData.MerkleProof))
	for i, proofHex := range proofData.MerkleProof {
		proofBytes, err := hex.DecodeString(trimHexPrefix(proofHex))
		if err != nil {
			return false, fmt.Errorf("error decoding proof element %d: %w", i, err)
		}
		copy(merkleProofBytes[i][:], proofBytes)
	}

	log.Printf("Verifying field index %d with value %s...", fieldIndex, proofData.FieldValue[:10])
	log.Printf("Using timestamp: %d", beaconTimestamp)
	log.Printf("Merkle proof length: %d", len(merkleProofBytes))

	input, err := parsedABI.Pack("verifyHeaderField", beaconTimestamp, fieldIndex, fieldValue, merkleProofBytes)
	if err != nil {
		return false, fmt.Errorf("error packing input data: %w", err)
	}

	msg := ethereum.CallMsg{
		To:   &address,
		Data: input,
	}

	result, err := client.CallContract(context.Background(), msg, nil)
	if err != nil {
		return false, fmt.Errorf("error calling contract: %w", err)
	}

	var verificationResult bool
	if err := parsedABI.UnpackIntoInterface(&verificationResult, "verifyHeaderField", result); err != nil {
		return false, fmt.Errorf("error unpacking result: %w", err)
	}

	if verificationResult {
		log.Println("On-chain verification successful! ✅")
	} else {
		log.Println("On-chain verification failed: Proof is invalid. ❌")
	}

	return verificationResult, nil
}

// Helper function to get map keys as a slice
func getMapKeys(m map[string]int) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

// Helper function to trim "0x" prefix from hex strings
func trimHexPrefix(hexStr string) string {
	if len(hexStr) >= 2 && hexStr[0:2] == "0x" {
		return hexStr[2:]
	}
	return hexStr
}
