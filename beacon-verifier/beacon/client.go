package beacon

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
)

// APIResponse represents the top-level structure of a Beacon API response
type APIResponse struct {
	Data struct {
		Root   string `json:"root"`
		Header struct {
			Message struct {
				Slot          string `json:"slot"`
				ProposerIndex string `json:"proposer_index"`
				ParentRoot    string `json:"parent_root"`
				StateRoot     string `json:"state_root"`
				BodyRoot      string `json:"body_root"`
			} `json:"message"`
		} `json:"header"`
	} `json:"data"`
}

// BlockResponse represents the response for a beacon block request
type BlockResponse struct {
	Data struct {
		Message struct {
			Body struct {
				ExecutionPayload struct {
					Timestamp string `json:"timestamp"`
				} `json:"execution_payload"`
			} `json:"body"`
		} `json:"message"`
	} `json:"data"`
}

// Client provides methods to interact with the Beacon API
type Client struct {
	BaseURL string
}

// Direction represents the direction to fetch the block header
type Direction int

const (
	Previous Direction = iota
	Next
	Requested
)

// NewClient creates a new Beacon API client
func NewClient(baseURL string) *Client {
	return &Client{
		BaseURL: baseURL,
	}
}

// FetchBlockHeader fetches a beacon block header from the API
func (c *Client) FetchBlockHeader(slot string) (HeaderData, error) {
	return c.fetchBlockData(slot)
}

// fetchBlockData fetches beacon block header and timestamp from API
func (c *Client) fetchBlockData(slot string) (HeaderData, error) {
	var headerData HeaderData

	// Fetch the header data
	apiURL := fmt.Sprintf("%s/eth/v1/beacon/headers/%s", c.BaseURL, slot)
	resp, err := http.Get(apiURL)
	if err != nil {
		return headerData, fmt.Errorf("error fetching header: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return headerData, fmt.Errorf("API returned status code %d", resp.StatusCode)
	}

	var apiResp APIResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return headerData, fmt.Errorf("error decoding API response: %w", err)
	}

	headerData.Slot = apiResp.Data.Header.Message.Slot
	headerData.ProposerIndex = apiResp.Data.Header.Message.ProposerIndex
	headerData.ParentRoot = apiResp.Data.Header.Message.ParentRoot
	headerData.StateRoot = apiResp.Data.Header.Message.StateRoot
	headerData.BodyRoot = apiResp.Data.Header.Message.BodyRoot

	// Add block root if available
	if apiResp.Data.Root != "" {
		headerData.BlockRoot = apiResp.Data.Root
	}

	// Fetch the block to get the timestamp
	blockURL := fmt.Sprintf("%s/eth/v2/beacon/blocks/%s", c.BaseURL, slot)
	blockResp, err := http.Get(blockURL)
	if err != nil {
		return headerData, fmt.Errorf("error fetching block data: %w", err)
	}
	defer blockResp.Body.Close()

	if blockResp.StatusCode == http.StatusOK {
		var blockData BlockResponse
		if err := json.NewDecoder(blockResp.Body).Decode(&blockData); err != nil {
			return HeaderData{}, fmt.Errorf("error decoding block response: %w", err)
		}
		// Extract timestamp
		timestampStr := blockData.Data.Message.Body.ExecutionPayload.Timestamp
		if timestampStr != "" {
			headerData.Timestamp, err = strconv.ParseInt(timestampStr, 10, 64)
			if err != nil {
				return HeaderData{}, fmt.Errorf("error parsing timestamp: %w", err)
			}
			return headerData, nil
		}
	}
	return HeaderData{}, errors.New("block data not found")
}
