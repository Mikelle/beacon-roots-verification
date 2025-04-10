package app

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/ethereum/go-ethereum/ethclient"

	"github.com/Mikelle/beacon-root-verification/beacon-verifier/beacon"
	"github.com/Mikelle/beacon-root-verification/beacon-verifier/config"
	"github.com/Mikelle/beacon-root-verification/beacon-verifier/proof"
)

// Application encapsulates the beacon verification application
type Application struct {
	Config         *config.Config
	BeaconClient   *beacon.Client
	EthereumClient *ethclient.Client
	Web3Connected  bool
}

// NewApplication creates and initializes a new application instance
func NewApplication(cfg *config.Config) (*Application, error) {
	if len(cfg.BeaconAPI.Endpoints) == 0 {
		return nil, fmt.Errorf("no beacon API endpoints configured")
	}

	app := &Application{
		Config:        cfg,
		BeaconClient:  beacon.NewClient(cfg.BeaconAPI.Endpoints[0]),
		Web3Connected: false,
	}

	// Initialize Ethereum client for onchain verification
	client, err := ethclient.Dial(cfg.EthereumNode.Endpoint)
	if err != nil {
		log.Printf("Warning: Error setting up Web3: %v", err)
	} else {
		chainID, err := client.ChainID(context.Background())
		if err != nil {
			log.Printf("Warning: Error getting chain ID: %v", err)
		} else {
			app.EthereumClient = client
			app.Web3Connected = true
			log.Println("Connected to Ethereum node for onchain verification.")
			log.Printf("Network info: Chain ID: %s", chainID.String())
		}
	}

	return app, nil
}

// Run executes the main application logic
func (a *Application) Run() error {
	var (
		err                                  error
		nextFilledSlotHeader, headerToVerify beacon.HeaderData
	)
	if a.Config.Slot != "" {
		log.Printf("Using specified slot %s for verification...", a.Config.Slot)
		headerToVerify, err = a.fetchHeader(beacon.HeaderData{Slot: a.Config.Slot}, beacon.Requested)
		if err != nil {
			return fmt.Errorf("error fetching specified slot %s: %w", a.Config.Slot, err)
		}
		nextFilledSlotHeader, err = a.fetchHeader(headerToVerify, beacon.Next)
		if err != nil {
			return fmt.Errorf("error fetching previous header: %w", err)
		}
	} else {
		nextFilledSlotHeader, err = a.fetchLatestHeader()
		if err != nil {
			return fmt.Errorf("error fetching latest header: %w", err)
		}

		log.Println("No specific slot provided. Attempting to fetch a previous header for verification...")
		headerToVerify, err = a.fetchHeader(nextFilledSlotHeader, beacon.Previous)
		if err != nil {
			return fmt.Errorf("error fetching previous header: %w", err)
		}
	}

	a.displayHeaderInfo(nextFilledSlotHeader, headerToVerify)

	results, err := a.verifyFields(headerToVerify, nextFilledSlotHeader)
	if err != nil {
		return fmt.Errorf("error during verification: %w", err)
	}

	a.displayResults(results)

	return nil
}

// fetchLatestHeader retrieves the latest beacon block header
func (a *Application) fetchLatestHeader() (beacon.HeaderData, error) {
	log.Printf("Fetching latest beacon block header from %s...", a.Config.BeaconAPI.Endpoints[0])

	latestHeaderData, err := a.BeaconClient.FetchBlockHeader("head")
	if err != nil {
		return beacon.HeaderData{}, fmt.Errorf("could not fetch latest beacon block header: %w", err)
	}

	latestSlot, err := strconv.ParseUint(latestHeaderData.Slot, 10, 64)
	if err != nil {
		return beacon.HeaderData{}, fmt.Errorf("error parsing latest slot: %w", err)
	}

	log.Printf("Latest block is at slot %d with timestamp %d", latestSlot, latestHeaderData.Timestamp)
	return latestHeaderData, nil
}

// fetchHeader attempts to fetch an adjacent block header for verification.
// The 'direction' parameter should be either Previous or Next.
// fetchHeader attempts to fetch an adjacent block header for verification.
// The 'direction' parameter should be either beacon.Previous, beacon.Next, or beacon.Requested.
func (a *Application) fetchHeader(header beacon.HeaderData, direction beacon.Direction) (beacon.HeaderData, error) {
	slot, err := strconv.ParseUint(header.Slot, 10, 64)
	if err != nil {
		return beacon.HeaderData{}, fmt.Errorf("error parsing header slot: %w", err)
	}

	maxAttempts := a.Config.BeaconAPI.RetryAttempts
	var targetSlot uint64
	// logic above is to handle missing slots and not take into account, that url could be unreachable
	for i := 1; i <= maxAttempts; i++ {
		switch direction {
		case beacon.Previous:
			// For a previous header, subtract the offset. Check to avoid underflow.
			if slot < uint64(i) {
				// If we can't subtract i from slot, break.
				break
			}
			targetSlot = slot - uint64(i)
		case beacon.Next:
			// For a next header, try the current header on the first attempt,
			// then subsequent slots.
			targetSlot = slot + uint64(i)
		case beacon.Requested:
			// For the requested header, try the provided header.
			targetSlot = slot
		default:
			return beacon.HeaderData{}, fmt.Errorf("invalid direction")
		}

		log.Printf("Fetching beacon block header at slot %d... (attempt %d/%d)", targetSlot, i, maxAttempts)

		currentHeaderData, err := a.BeaconClient.FetchBlockHeader(strconv.FormatUint(targetSlot, 10))
		if err == nil && currentHeaderData.Slot != "" {
			blockTimestamp := currentHeaderData.Timestamp
			log.Printf("Successfully fetched block header at slot %d", targetSlot)
			log.Printf("Block time: %s", time.Unix(blockTimestamp, 0).UTC().Format("2006-01-02 15:04:05 UTC"))
			log.Printf("Block proposer: %s", currentHeaderData.ProposerIndex)
			return currentHeaderData, nil
		} else {
			log.Printf("Error fetching block header at slot %d: %v", targetSlot, err)
		}
	}

	log.Printf("Failed to fetch any valid beacon block header after %d attempts.", maxAttempts)
	return beacon.HeaderData{}, fmt.Errorf("could not fetch any valid beacon block header")
}

// displayHeaderInfo shows information about the header being verified
func (a *Application) displayHeaderInfo(headerData beacon.HeaderData, nextHeader beacon.HeaderData) {
	log.Println("\nBeacon Block Header (to verify):")
	log.Printf("  slot: %s", headerData.Slot)
	log.Printf("  proposer_index: %s", headerData.ProposerIndex)
	log.Printf("  parent_root: %s", headerData.ParentRoot)
	log.Printf("  state_root: %s", headerData.StateRoot)
	log.Printf("  body_root: %s", headerData.BodyRoot)

	nextSlotTimestamp := nextHeader.Timestamp
	blockTimestamp := headerData.Timestamp
	timeDifference := nextSlotTimestamp - blockTimestamp

	log.Printf("\nNext filled slot timestamp: %d", nextSlotTimestamp)
	log.Printf("Time difference between blocks: %d seconds (%.2f slots)",
		timeDifference, float64(timeDifference)/beacon.SecondsPerSlot)
}

// verifyFields generates and verifies proofs for each field
func (a *Application) verifyFields(headerData beacon.HeaderData, nextFilledSlotHeader beacon.HeaderData) (map[string]bool, error) {
	proofResults := make(map[string]bool)

	if !a.Web3Connected {
		log.Println("Warning: No Ethereum connection available. Skipping on-chain verification.")
		return proofResults, nil
	}

	fields := a.Config.Verification.FieldsToVerify
	nextSlotTimestamp := nextFilledSlotHeader.Timestamp

	for _, fieldName := range fields {
		log.Printf("\n=== Generating proof for %s ===", fieldName)
		proofData, err := proof.GenerateHeaderProof(headerData, fieldName, nextSlotTimestamp)
		if err != nil {
			log.Printf("Error generating proof for %s: %v", fieldName, err)
			continue
		}

		log.Printf("Proof generated with %d elements", len(proofData.MerkleProof))

		// Perform onchain verification
		log.Println("\nPerforming onchain verification...")
		result, err := proof.VerifyOnChain(a.EthereumClient, a.Config.Verification.VerifierAddress, proofData)
		if err != nil {
			log.Printf("Error performing onchain verification: %v", err)
		} else {
			proofResults[fieldName] = result
		}
	}

	return proofResults, nil
}

// displayResults shows a summary of verification results
func (a *Application) displayResults(results map[string]bool) {
	if len(results) == 0 {
		log.Println("\nNo verification results to display.")
		return
	}

	log.Println("\n=== Summary of verification results ===")
	allPassed := true

	for field, result := range results {
		status := "✅ Passed"
		if !result {
			status = "❌ Failed"
			allPassed = false
		}
		log.Printf("%s: %s", field, status)
	}

	if allPassed {
		log.Println("\nAll verifications passed successfully! ✅")
	} else {
		log.Println("\nSome verifications failed. Check the results above. ❌")
	}
}
