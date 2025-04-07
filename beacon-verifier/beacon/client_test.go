package beacon

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// setupTestServer creates a test HTTP server that returns predefined responses
func setupTestServer(t *testing.T, headerHandler func(w http.ResponseWriter, r *http.Request),
	blockHandler func(w http.ResponseWriter, r *http.Request)) *httptest.Server {

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/eth/v1/beacon/headers/head" || r.URL.Path == "/eth/v1/beacon/headers/123456":
			headerHandler(w, r)
		case r.URL.Path == "/eth/v2/beacon/blocks/head" || r.URL.Path == "/eth/v2/beacon/blocks/123456":
			blockHandler(w, r)
		default:
			http.NotFound(w, r)
		}
	}))

	t.Cleanup(func() {
		server.Close()
	})

	return server
}

// Helper function to create a valid header response
func createValidHeaderResponse() APIResponse {
	var resp APIResponse
	resp.Data.Root = "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"
	resp.Data.Header.Message.Slot = "123456"
	resp.Data.Header.Message.ProposerIndex = "42"
	resp.Data.Header.Message.ParentRoot = "0x4a81947b35bdc11471fc7b42350427a3b9d2b92bf21d423ded6dcc5c66caad0e"
	resp.Data.Header.Message.StateRoot = "0x5bc9a4ef3cf09a315ffbc12872de6cc412a7abb55a5228cc21fbdb5fb797d7a8"
	resp.Data.Header.Message.BodyRoot = "0x67df26e0c9f5de4fe7b3f66f3591f84a9cf6e8cda7f5b3f23db5c3967a505c31"
	return resp
}

// Helper function to create a valid block response
func createValidBlockResponse() BlockResponse {
	var resp BlockResponse
	resp.Data.Message.Body.ExecutionPayload.Timestamp = "1651234567"
	return resp
}

func TestNewClient(t *testing.T) {
	baseURL := "https://example.com/api"
	client := NewClient(baseURL)

	if client.BaseURL != baseURL {
		t.Errorf("NewClient().BaseURL = %s, want %s", client.BaseURL, baseURL)
	}
}

func TestFetchBlockHeader_Success(t *testing.T) {
	// Setup test server with successful responses
	server := setupTestServer(t,
		func(w http.ResponseWriter, r *http.Request) {
			resp := createValidHeaderResponse()
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(resp)
		},
		func(w http.ResponseWriter, r *http.Request) {
			resp := createValidBlockResponse()
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(resp)
		},
	)

	client := NewClient(server.URL)
	headerData, err := client.FetchBlockHeader("123456")

	// Verify no error occurred
	if err != nil {
		t.Fatalf("FetchBlockHeader() error = %v", err)
	}

	// Verify the data was parsed correctly
	expectedSlot := "123456"
	if headerData.Slot != expectedSlot {
		t.Errorf("headerData.Slot = %s, want %s", headerData.Slot, expectedSlot)
	}

	expectedProposer := "42"
	if headerData.ProposerIndex != expectedProposer {
		t.Errorf("headerData.ProposerIndex = %s, want %s", headerData.ProposerIndex, expectedProposer)
	}

	expectedTimestamp := int64(1651234567)
	if headerData.Timestamp != expectedTimestamp {
		t.Errorf("headerData.Timestamp = %d, want %d", headerData.Timestamp, expectedTimestamp)
	}
}

func TestFetchBlockHeader_HeaderRequestFails(t *testing.T) {
	// Setup test server that returns error for header request
	server := setupTestServer(t,
		func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		},
		func(w http.ResponseWriter, r *http.Request) {
			resp := createValidBlockResponse()
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(resp)
		},
	)

	client := NewClient(server.URL)
	_, err := client.FetchBlockHeader("123456")

	// Verify an error was returned
	if err == nil {
		t.Fatalf("FetchBlockHeader() expected error, got nil")
	}
}

func TestFetchBlockHeader_InvalidHeaderJson(t *testing.T) {
	// Setup test server that returns invalid JSON
	server := setupTestServer(t,
		func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte("not valid json"))
		},
		func(w http.ResponseWriter, r *http.Request) {
			resp := createValidBlockResponse()
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(resp)
		},
	)

	client := NewClient(server.URL)
	_, err := client.FetchBlockHeader("123456")

	// Verify an error was returned
	if err == nil {
		t.Fatalf("FetchBlockHeader() expected error, got nil")
	}
}

func TestFetchBlockHeader_BlockRequestFails(t *testing.T) {
	// Setup test server with header success but block failure
	server := setupTestServer(t,
		func(w http.ResponseWriter, r *http.Request) {
			resp := createValidHeaderResponse()
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(resp)
		},
		func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		},
	)

	client := NewClient(server.URL)
	headerData, err := client.FetchBlockHeader("123456")

	// Test should still pass because block request failure uses fallback timestamp
	if err != nil {
		t.Fatalf("FetchBlockHeader() error = %v", err)
	}

	// Should be using current time as fallback
	now := time.Now().Unix()
	if headerData.Timestamp == 0 || headerData.Timestamp > now+1 || headerData.Timestamp < now-5 {
		t.Errorf("headerData.Timestamp = %d, expected approximately %d", headerData.Timestamp, now)
	}
}

func TestFetchBlockHeader_InvalidBlockJson(t *testing.T) {
	// Setup test server with header success but invalid block JSON
	server := setupTestServer(t,
		func(w http.ResponseWriter, r *http.Request) {
			resp := createValidHeaderResponse()
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(resp)
		},
		func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte("not valid json"))
		},
	)

	client := NewClient(server.URL)
	headerData, err := client.FetchBlockHeader("123456")

	// Should still pass with fallback timestamp
	if err != nil {
		t.Fatalf("FetchBlockHeader() error = %v", err)
	}

	// Should be using current time as fallback
	now := time.Now().Unix()
	if headerData.Timestamp == 0 || headerData.Timestamp > now+1 || headerData.Timestamp < now-5 {
		t.Errorf("headerData.Timestamp = %d, expected approximately %d", headerData.Timestamp, now)
	}
}

func TestFetchBlockHeader_InvalidTimestamp(t *testing.T) {
	// Setup test server that returns invalid timestamp in block response
	server := setupTestServer(t,
		func(w http.ResponseWriter, r *http.Request) {
			resp := createValidHeaderResponse()
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(resp)
		},
		func(w http.ResponseWriter, r *http.Request) {
			var resp BlockResponse
			resp.Data.Message.Body.ExecutionPayload.Timestamp = "not-a-number"
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(resp)
		},
	)

	client := NewClient(server.URL)
	headerData, err := client.FetchBlockHeader("123456")

	// Should still pass with fallback timestamp
	if err != nil {
		t.Fatalf("FetchBlockHeader() error = %v", err)
	}

	// Should be using current time as fallback
	now := time.Now().Unix()
	if headerData.Timestamp == 0 || headerData.Timestamp > now+1 || headerData.Timestamp < now-5 {
		t.Errorf("headerData.Timestamp = %d, expected approximately %d", headerData.Timestamp, now)
	}
}
