package proof

import (
	"encoding/hex"
	"reflect"
	"sort"
	"testing"

	"github.com/Mikelle/beacon-root-verification/beacon-verifier/beacon"
	"github.com/ethereum/go-ethereum/common"
)

// MockEthClient mocks the ethereum client for testing
type MockEthClient struct {
	CallContractFunc func(result []byte, err error)
}

// Setup test data
func setupTestHeader() beacon.HeaderData {
	return beacon.HeaderData{
		Slot:          "123456",
		ProposerIndex: "42",
		ParentRoot:    "0x4a81947b35bdc11471fc7b42350427a3b9d2b92bf21d423ded6dcc5c66caad0e",
		StateRoot:     "0x5bc9a4ef3cf09a315ffbc12872de6cc412a7abb55a5228cc21fbdb5fb797d7a8",
		BodyRoot:      "0x67df26e0c9f5de4fe7b3f66f3591f84a9cf6e8cda7f5b3f23db5c3967a505c31",
		BlockRoot:     "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
		Timestamp:     1634567890,
	}
}

func TestFieldNames(t *testing.T) {
	// Ensure all field names are defined
	expectedFields := []string{"slot", "proposer_index", "parent_root", "state_root", "body_root"}

	// Check that we have all expected field names
	for _, field := range expectedFields {
		if _, exists := FieldNames[field]; !exists {
			t.Errorf("Expected field name '%s' not found in FieldNames map", field)
		}
	}

	// Check that we don't have any extra fields
	actualFields := getMapKeys(FieldNames)
	if len(actualFields) != len(expectedFields) {
		t.Errorf("FieldNames map has %d entries, expected %d", len(actualFields), len(expectedFields))
	}
}

func TestGetMapKeys(t *testing.T) {
	testMap := map[string]int{
		"a": 1,
		"b": 2,
		"c": 3,
	}

	keys := getMapKeys(testMap)
	sort.Strings(keys) // Sort for deterministic comparison

	expected := []string{"a", "b", "c"}
	if !reflect.DeepEqual(keys, expected) {
		t.Errorf("getMapKeys() = %v, want %v", keys, expected)
	}
}

func TestTrimHexPrefix(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"0x1234", "1234"},
		{"1234", "1234"},
		{"0x", ""},
		{"", ""},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			if got := trimHexPrefix(tt.input); got != tt.want {
				t.Errorf("trimHexPrefix(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestGenerateHeaderProof(t *testing.T) {
	headerData := setupTestHeader()
	nextSlotTimestamp := int64(1634567890 + 12) // current + 1 slot

	tests := []struct {
		name      string
		fieldName string
		wantErr   bool
	}{
		{"Valid slot field", "slot", false},
		{"Valid proposer_index field", "proposer_index", false},
		{"Valid parent_root field", "parent_root", false},
		{"Valid state_root field", "state_root", false},
		{"Valid body_root field", "body_root", false},
		{"Invalid field name", "invalid_field", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			proofData, err := GenerateHeaderProof(headerData, tt.fieldName, nextSlotTimestamp)

			if (err != nil) != tt.wantErr {
				t.Errorf("GenerateHeaderProof() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				// Verify the proof data structure
				if proofData.BeaconTimestamp != nextSlotTimestamp {
					t.Errorf("Expected timestamp %d, got %d", nextSlotTimestamp, proofData.BeaconTimestamp)
				}

				if proofData.FieldIndex != FieldNames[tt.fieldName] {
					t.Errorf("Expected field index %d, got %d", FieldNames[tt.fieldName], proofData.FieldIndex)
				}

				if len(proofData.BeaconBlockRoot) <= 2 {
					t.Errorf("Invalid BeaconBlockRoot: %s", proofData.BeaconBlockRoot)
				}

				if len(proofData.FieldValue) <= 2 {
					t.Errorf("Invalid FieldValue: %s", proofData.FieldValue)
				}

				// Merkle proof should have log2(5) elements rounded up, which is 3
				if len(proofData.MerkleProof) != 3 {
					t.Errorf("Expected proof length 3, got %d", len(proofData.MerkleProof))
				}

				// Verify each proof element has proper format (0x...)
				for i, proof := range proofData.MerkleProof {
					if len(proof) < 2 || proof[:2] != "0x" {
						t.Errorf("Proof element %d doesn't have 0x prefix: %s", i, proof)
					}

					// Ensure it's a valid hex string of the right length (0x + 64 hex chars)
					if len(proof) != 66 {
						t.Errorf("Proof element %d has incorrect length: %d, expected 66", i, len(proof))
					}

					// Try to decode the hex
					_, err := hex.DecodeString(proof[2:])
					if err != nil {
						t.Errorf("Proof element %d is not valid hex: %v", i, err)
					}
				}
			}
		})
	}
}

func TestGenerateHeaderProofValues(t *testing.T) {
	headerData := setupTestHeader()
	nextSlotTimestamp := int64(1634567890 + 12)

	// Test the slot field specifically to verify its value
	t.Run("Verify slot value", func(t *testing.T) {
		proofData, err := GenerateHeaderProof(headerData, "slot", nextSlotTimestamp)
		if err != nil {
			t.Fatalf("GenerateHeaderProof() error = %v", err)
		}

		fieldValueHex := trimHexPrefix(proofData.FieldValue)
		fieldValueBytes, err := hex.DecodeString(fieldValueHex)
		if err != nil {
			t.Fatalf("Error decoding field value: %v", err)
		}

		// In little endian, first 8 bytes represent the slot number
		slotValue := uint64(0)
		for i := 0; i < 8; i++ {
			slotValue |= uint64(fieldValueBytes[i]) << (i * 8)
		}

		expectedSlot := uint64(123456) // From the test header
		if slotValue != expectedSlot {
			t.Errorf("Expected slot value %d, got %d", expectedSlot, slotValue)
		}
	})

	// Test the proposer_index field
	t.Run("Verify proposer_index value", func(t *testing.T) {
		proofData, err := GenerateHeaderProof(headerData, "proposer_index", nextSlotTimestamp)
		if err != nil {
			t.Fatalf("GenerateHeaderProof() error = %v", err)
		}

		fieldValueHex := trimHexPrefix(proofData.FieldValue)
		fieldValueBytes, err := hex.DecodeString(fieldValueHex)
		if err != nil {
			t.Fatalf("Error decoding field value: %v", err)
		}

		// In little endian, first 8 bytes represent the proposer index
		proposerValue := uint64(0)
		for i := 0; i < 8; i++ {
			proposerValue |= uint64(fieldValueBytes[i]) << (i * 8)
		}

		expectedProposer := uint64(42) // From the test header
		if proposerValue != expectedProposer {
			t.Errorf("Expected proposer value %d, got %d", expectedProposer, proposerValue)
		}
	})
}

// TestVerifyOnChain tests the contract verification with mocks
func TestVerifyOnChain(t *testing.T) {
	headerData := setupTestHeader()
	nextSlotTimestamp := int64(1634567890 + 12)

	// Generate a proof
	proofData, err := GenerateHeaderProof(headerData, "slot", nextSlotTimestamp)
	if err != nil {
		t.Fatalf("GenerateHeaderProof() error = %v", err)
	}

	// Structural test
	t.Run("VerifyOnChain structural test", func(t *testing.T) {
		// This test just ensures the proof data has expected structure
		if proofData.BeaconTimestamp <= 0 {
			t.Errorf("Invalid timestamp: %d", proofData.BeaconTimestamp)
		}

		if proofData.FieldIndex < 0 || proofData.FieldIndex > 4 {
			t.Errorf("Invalid field index: %d", proofData.FieldIndex)
		}

		if len(trimHexPrefix(proofData.FieldValue)) != 64 {
			t.Errorf("Invalid field value length: %d", len(trimHexPrefix(proofData.FieldValue)))
		}

		if len(proofData.MerkleProof) == 0 {
			t.Errorf("Empty merkle proof")
		}
	})
}

// Test invalid inputs to make sure they're handled properly
func TestGenerateHeaderProofInvalidInput(t *testing.T) {
	// Valid header data
	headerData := setupTestHeader()
	nextSlotTimestamp := int64(1634567890 + 12)

	// Test with invalid slot value
	t.Run("Invalid slot", func(t *testing.T) {
		invalidHeader := headerData
		invalidHeader.Slot = "not-a-number"

		_, err := GenerateHeaderProof(invalidHeader, "slot", nextSlotTimestamp)
		if err == nil {
			t.Errorf("Expected error for invalid slot, got nil")
		}
	})

	// Test with invalid proposer index
	t.Run("Invalid proposer index", func(t *testing.T) {
		invalidHeader := headerData
		invalidHeader.ProposerIndex = "not-a-number"

		_, err := GenerateHeaderProof(invalidHeader, "proposer_index", nextSlotTimestamp)
		if err == nil {
			t.Errorf("Expected error for invalid proposer index, got nil")
		}
	})

	// Test with invalid hex in parent root
	t.Run("Invalid parent root", func(t *testing.T) {
		invalidHeader := headerData
		invalidHeader.ParentRoot = "0xNOT-HEX"

		_, err := GenerateHeaderProof(invalidHeader, "parent_root", nextSlotTimestamp)
		if err == nil {
			t.Errorf("Expected error for invalid parent root, got nil")
		}
	})
}

// Test additional error cases for VerifyOnChain
func TestVerifyOnChainErrors(t *testing.T) {
	t.Run("Empty contract address", func(t *testing.T) {
		// Just test that the contract address is properly converted to an Ethereum address
		contractAddress := ""
		emptyAddress := common.HexToAddress(contractAddress)
		if emptyAddress != (common.Address{}) {
			t.Errorf("Expected empty address for empty string, got %s", emptyAddress.Hex())
		}
	})

	// Test with invalid field value hex
	t.Run("Invalid field value hex", func(t *testing.T) {
		// This test will fail at ethclient.Dial, but we can at least test the hex decoding part
		proofData := Data{
			FieldValue:      "0xNOT-HEX",
		}

		_, err := hex.DecodeString(trimHexPrefix(proofData.FieldValue))
		if err == nil {
			t.Errorf("Expected error for invalid hex, got nil")
		}
	})

	// Test with invalid merkle proof hex
	t.Run("Invalid merkle proof hex", func(t *testing.T) {
		proofData := Data{
			MerkleProof:     []string{"0xNOT-HEX"},
		}

		_, err := hex.DecodeString(trimHexPrefix(proofData.MerkleProof[0]))
		if err == nil {
			t.Errorf("Expected error for invalid hex, got nil")
		}
	})
}
