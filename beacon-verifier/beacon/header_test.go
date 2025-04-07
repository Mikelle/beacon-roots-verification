package beacon

import (
	"bytes"
	"encoding/hex"
	"strconv"
	"testing"
)

func TestBlockHeaderFromAPIResponse(t *testing.T) {
	tests := []struct {
		name    string
		data    HeaderData
		wantErr bool
	}{
		{
			name: "Valid header data",
			data: HeaderData{
				Slot:          "123456",
				ProposerIndex: "42",
				ParentRoot:    "0x4a81947b35bdc11471fc7b42350427a3b9d2b92bf21d423ded6dcc5c66caad0e",
				StateRoot:     "0x5bc9a4ef3cf09a315ffbc12872de6cc412a7abb55a5228cc21fbdb5fb797d7a8",
				BodyRoot:      "0x67df26e0c9f5de4fe7b3f66f3591f84a9cf6e8cda7f5b3f23db5c3967a505c31",
			},
			wantErr: false,
		},
		{
			name: "Empty roots - should fill with zeros",
			data: HeaderData{
				Slot:          "123456",
				ProposerIndex: "42",
				ParentRoot:    "",
				StateRoot:     "",
				BodyRoot:      "",
			},
			wantErr: false,
		},
		{
			name: "Invalid slot - not a number",
			data: HeaderData{
				Slot:          "not-a-number",
				ProposerIndex: "42",
				ParentRoot:    "0x4a81947b35bdc11471fc7b42350427a3b9d2b92bf21d423ded6dcc5c66caad0e",
				StateRoot:     "0x5bc9a4ef3cf09a315ffbc12872de6cc412a7abb55a5228cc21fbdb5fb797d7a8",
				BodyRoot:      "0x67df26e0c9f5de4fe7b3f66f3591f84a9cf6e8cda7f5b3f23db5c3967a505c31",
			},
			wantErr: true,
		},
		{
			name: "Invalid proposer index - not a number",
			data: HeaderData{
				Slot:          "123456",
				ProposerIndex: "not-a-number",
				ParentRoot:    "0x4a81947b35bdc11471fc7b42350427a3b9d2b92bf21d423ded6dcc5c66caad0e",
				StateRoot:     "0x5bc9a4ef3cf09a315ffbc12872de6cc412a7abb55a5228cc21fbdb5fb797d7a8",
				BodyRoot:      "0x67df26e0c9f5de4fe7b3f66f3591f84a9cf6e8cda7f5b3f23db5c3967a505c31",
			},
			wantErr: true,
		},
		{
			name: "Invalid parent root - not hex",
			data: HeaderData{
				Slot:          "123456",
				ProposerIndex: "42",
				ParentRoot:    "0xNOT-HEX",
				StateRoot:     "0x5bc9a4ef3cf09a315ffbc12872de6cc412a7abb55a5228cc21fbdb5fb797d7a8",
				BodyRoot:      "0x67df26e0c9f5de4fe7b3f66f3591f84a9cf6e8cda7f5b3f23db5c3967a505c31",
			},
			wantErr: true,
		},
		{
			name: "Invalid state root - not hex",
			data: HeaderData{
				Slot:          "123456",
				ProposerIndex: "42",
				ParentRoot:    "0x4a81947b35bdc11471fc7b42350427a3b9d2b92bf21d423ded6dcc5c66caad0e",
				StateRoot:     "0xNOT-HEX",
				BodyRoot:      "0x67df26e0c9f5de4fe7b3f66f3591f84a9cf6e8cda7f5b3f23db5c3967a505c31",
			},
			wantErr: true,
		},
		{
			name: "Invalid body root - not hex",
			data: HeaderData{
				Slot:          "123456",
				ProposerIndex: "42",
				ParentRoot:    "0x4a81947b35bdc11471fc7b42350427a3b9d2b92bf21d423ded6dcc5c66caad0e",
				StateRoot:     "0x5bc9a4ef3cf09a315ffbc12872de6cc412a7abb55a5228cc21fbdb5fb797d7a8",
				BodyRoot:      "0xNOT-HEX",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var header BlockHeader
			err := header.FromAPIResponse(tt.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("BlockHeader.FromAPIResponse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				// Verify slot value
				slotWant, _ := strconv.ParseUint(tt.data.Slot, 10, 64)
				if header.Slot != slotWant {
					t.Errorf("BlockHeader.Slot = %d, want %d", header.Slot, slotWant)
				}

				// Verify proposer index
				proposerWant, _ := strconv.ParseUint(tt.data.ProposerIndex, 10, 64)
				if header.ProposerIndex != proposerWant {
					t.Errorf("BlockHeader.ProposerIndex = %d, want %d", header.ProposerIndex, proposerWant)
				}

				// Verify parent root
				if tt.data.ParentRoot == "" {
					if !bytes.Equal(header.ParentRoot, make([]byte, 32)) {
						t.Errorf("BlockHeader.ParentRoot not correctly zero-filled for empty input")
					}
				} else {
					parentRootWant, _ := hex.DecodeString(trimHexPrefix(tt.data.ParentRoot))
					if !bytes.Equal(header.ParentRoot, parentRootWant) {
						t.Errorf("BlockHeader.ParentRoot = %x, want %x", header.ParentRoot, parentRootWant)
					}
				}

				// Similar checks for state root and body root could be added
			}
		})
	}
}

func TestSerializeForMerkleization(t *testing.T) {
	header := BlockHeader{
		Slot:          123456,
		ProposerIndex: 42,
		ParentRoot:    bytes.Repeat([]byte{0x01}, 32),
		StateRoot:     bytes.Repeat([]byte{0x02}, 32),
		BodyRoot:      bytes.Repeat([]byte{0x03}, 32),
	}

	serialized := header.SerializeForMerkleization()

	// Check number of fields
	if len(serialized) != 5 {
		t.Errorf("SerializeForMerkleization() returned %d fields, want 5", len(serialized))
	}

	// Check each field has correct length
	for i, field := range serialized {
		if len(field) != 32 {
			t.Errorf("Field %d has length %d, want 32", i, len(field))
		}
	}

	// Verify slot serialization (little-endian)
	slotBytes := serialized[0]
	slotValue := uint64(0)
	for i := 0; i < 8; i++ {
		slotValue |= uint64(slotBytes[i]) << (i * 8)
	}
	if slotValue != header.Slot {
		t.Errorf("Deserialized slot = %d, want %d", slotValue, header.Slot)
	}

	// Verify proposer index serialization
	proposerBytes := serialized[1]
	proposerValue := uint64(0)
	for i := 0; i < 8; i++ {
		proposerValue |= uint64(proposerBytes[i]) << (i * 8)
	}
	if proposerValue != header.ProposerIndex {
		t.Errorf("Deserialized proposer index = %d, want %d", proposerValue, header.ProposerIndex)
	}

	// Verify root fields
	if !bytes.Equal(serialized[2], header.ParentRoot) {
		t.Errorf("Serialized parent root doesn't match original")
	}
	if !bytes.Equal(serialized[3], header.StateRoot) {
		t.Errorf("Serialized state root doesn't match original")
	}
	if !bytes.Equal(serialized[4], header.BodyRoot) {
		t.Errorf("Serialized body root doesn't match original")
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

func TestWriteUint64LittleEndian(t *testing.T) {
	tests := []struct {
		value uint64
		want  []byte
	}{
		{0, []byte{0, 0, 0, 0, 0, 0, 0, 0}},
		{1, []byte{1, 0, 0, 0, 0, 0, 0, 0}},
		{256, []byte{0, 1, 0, 0, 0, 0, 0, 0}},
		{0xdeadbeef, []byte{0xef, 0xbe, 0xad, 0xde, 0, 0, 0, 0}},
		{0x123456789abcdef0, []byte{0xf0, 0xde, 0xbc, 0x9a, 0x78, 0x56, 0x34, 0x12}},
	}

	for _, tt := range tests {
		t.Run("uint64="+strconv.FormatUint(tt.value, 10), func(t *testing.T) {
			buf := make([]byte, 8)
			writeUint64LittleEndian(buf, tt.value)
			if !bytes.Equal(buf, tt.want) {
				t.Errorf("writeUint64LittleEndian(%d) = %v, want %v", tt.value, buf, tt.want)
			}
		})
	}
}
