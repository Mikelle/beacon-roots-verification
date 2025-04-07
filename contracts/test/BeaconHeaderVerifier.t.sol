// SPDX-License-Identifier: MIT
pragma solidity ^0.8.26;

import "forge-std/Test.sol";
import "../src/BeaconHeaderVerifier.sol";

/**
 * @title MockBeaconRoots
 * @dev Mock implementation of the Beacon Roots contract for testing
 */
contract MockBeaconRoots {
    mapping(uint256 => bytes32) public beaconBlockRoots;
    
    function setBeaconBlockRoot(uint256 timestamp, bytes32 root) external {
        beaconBlockRoots[timestamp] = root;
    }
    
    fallback(bytes calldata input) external returns (bytes memory) {
        uint256 timestamp = abi.decode(input, (uint256));
        return abi.encode(beaconBlockRoots[timestamp]);
    }
}

contract BeaconHeaderVerifierTest is Test {
    // Constant address from the original contract
    address constant BEACON_ROOTS_ADDRESS = 0x000F3df6D732807Ef1319fB7B8bB8522d0Beac02;
    
    BeaconHeaderVerifier private verifier;
    MockBeaconRoots private mockBeaconRoots;
    
    // Test variables
    uint256 constant TEST_TIMESTAMP = 1650000000;
    bytes32 constant TEST_BLOCK_ROOT = 0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef;
    
    // Example field values
    uint64 constant TEST_SLOT = 123456;
    uint64 constant TEST_PROPOSER_INDEX = 789;
    bytes32 constant TEST_PARENT_ROOT = 0xaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa;
    bytes32 constant TEST_STATE_ROOT = 0xbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb;
    bytes32 constant TEST_BODY_ROOT = 0xcccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc;
    
    // Mock merkle proofs 
    bytes32[] mockProof;
    
    function setUp() public {
        // Deploy mock beacon roots contract
        mockBeaconRoots = new MockBeaconRoots();
        
        // Create verifier after setting up the mock contract
        verifier = new BeaconHeaderVerifier();
        
        // Use vm.etch to replace the beacon roots contract at the expected address
        bytes memory mockCode = address(mockBeaconRoots).code;
        vm.etch(BEACON_ROOTS_ADDRESS, mockCode);
        
        // Set up the block root for our test timestamp
        vm.mockCall(
            BEACON_ROOTS_ADDRESS,
            abi.encode(TEST_TIMESTAMP),
            abi.encode(TEST_BLOCK_ROOT)
        );
        
        // Setup mock proofs for testing
        mockProof = new bytes32[](3);
        mockProof[0] = bytes32(uint256(1));
        mockProof[1] = bytes32(uint256(2));
        mockProof[2] = bytes32(uint256(3));
    }
    
    function test_GetBeaconBlockRoot() public {
        bytes32 root = verifier.getBeaconBlockRoot(TEST_TIMESTAMP);
        assertEq(root, TEST_BLOCK_ROOT, "Should return the correct beacon block root");
    }
    
    function test_GetBeaconBlockRoot_NonExistentTimestamp() public {
        uint256 nonExistentTimestamp = TEST_TIMESTAMP + 1;
        
        // Mock call for the non-existent timestamp to return zero
        vm.mockCall(
            BEACON_ROOTS_ADDRESS,
            abi.encode(nonExistentTimestamp),
            abi.encode(bytes32(0))
        );
        
        bytes32 root = verifier.getBeaconBlockRoot(nonExistentTimestamp);
        assertEq(root, bytes32(0), "Should return zero for non-existent timestamp");
    }
    
    function test_VerifySlot() public {
        // Calculate what the expected root would be for this test case
        bytes32 slotBytes = bytes32(uint256(TEST_SLOT));
        bytes32 expectedRoot = _computeExpectedRoot(0, slotBytes, mockProof);
        
        // Mock the beacon roots contract to return our expected root
        vm.mockCall(
            BEACON_ROOTS_ADDRESS,
            abi.encode(TEST_TIMESTAMP),
            abi.encode(expectedRoot)
        );
        
        bool result = verifier.verifySlot(TEST_TIMESTAMP, TEST_SLOT, mockProof);
        assertTrue(result, "Should verify the slot successfully");
    }
    
    function test_VerifyProposerIndex() public {
        bytes32 proposerBytes = bytes32(uint256(TEST_PROPOSER_INDEX));
        bytes32 expectedRoot = _computeExpectedRoot(1, proposerBytes, mockProof);
        
        vm.mockCall(
            BEACON_ROOTS_ADDRESS,
            abi.encode(TEST_TIMESTAMP),
            abi.encode(expectedRoot)
        );
        
        bool result = verifier.verifyProposerIndex(TEST_TIMESTAMP, TEST_PROPOSER_INDEX, mockProof);
        assertTrue(result, "Should verify the proposer index successfully");
    }
    
    function test_VerifyParentRoot() public {
        bytes32 expectedRoot = _computeExpectedRoot(2, TEST_PARENT_ROOT, mockProof);
        
        vm.mockCall(
            BEACON_ROOTS_ADDRESS,
            abi.encode(TEST_TIMESTAMP),
            abi.encode(expectedRoot)
        );
        
        bool result = verifier.verifyParentRoot(TEST_TIMESTAMP, TEST_PARENT_ROOT, mockProof);
        assertTrue(result, "Should verify the parent root successfully");
    }
    
    function test_VerifyStateRoot() public {
        bytes32 expectedRoot = _computeExpectedRoot(3, TEST_STATE_ROOT, mockProof);
        
        vm.mockCall(
            BEACON_ROOTS_ADDRESS,
            abi.encode(TEST_TIMESTAMP),
            abi.encode(expectedRoot)
        );
        
        bool result = verifier.verifyStateRoot(TEST_TIMESTAMP, TEST_STATE_ROOT, mockProof);
        assertTrue(result, "Should verify the state root successfully");
    }
    
    function test_VerifyBodyRoot() public {
        bytes32 expectedRoot = _computeExpectedRoot(4, TEST_BODY_ROOT, mockProof);
        
        vm.mockCall(
            BEACON_ROOTS_ADDRESS,
            abi.encode(TEST_TIMESTAMP),
            abi.encode(expectedRoot)
        );
        
        bool result = verifier.verifyBodyRoot(TEST_TIMESTAMP, TEST_BODY_ROOT, mockProof);
        assertTrue(result, "Should verify the body root successfully");
    }
    
    function test_VerifyHeaderField_InvalidIndex() public {
        vm.expectRevert("Invalid field index");
        verifier.verifyHeaderField(TEST_TIMESTAMP, 5, bytes32(0), mockProof);
    }
    
    function test_VerifyHeaderField_InvalidProof() public {
        // Setup a deliberately incorrect root that won't match what the verification computes
        bytes32 incorrectRoot = bytes32(uint256(999));
        
        vm.mockCall(
            BEACON_ROOTS_ADDRESS,
            abi.encode(TEST_TIMESTAMP),
            abi.encode(incorrectRoot)
        );
        
        bool result = verifier.verifyStateRoot(TEST_TIMESTAMP, TEST_STATE_ROOT, mockProof);
        assertFalse(result, "Should return false for invalid proof");
    }
    
    function testFuzz_VerifySlot(uint64 slot) public {
        // Convert slot to bytes32 as the contract would
        bytes32 slotBytes = bytes32(uint256(slot));
        
        // Generate expected root based on our proof computation
        bytes32 expectedRoot = _computeExpectedRoot(0, slotBytes, mockProof);
        
        // Mock the beacon roots contract to return our expected root
        vm.mockCall(
            BEACON_ROOTS_ADDRESS,
            abi.encode(TEST_TIMESTAMP),
            abi.encode(expectedRoot)
        );
        
        bool result = verifier.verifySlot(TEST_TIMESTAMP, slot, mockProof);
        assertTrue(result, "Should verify the slot in fuzz test");
    }
    
    function testFuzz_VerifyProposerIndex(uint64 proposerIndex) public {
        bytes32 proposerBytes = bytes32(uint256(proposerIndex));
        bytes32 expectedRoot = _computeExpectedRoot(1, proposerBytes, mockProof);
        
        vm.mockCall(
            BEACON_ROOTS_ADDRESS,
            abi.encode(TEST_TIMESTAMP),
            abi.encode(expectedRoot)
        );
        
        bool result = verifier.verifyProposerIndex(TEST_TIMESTAMP, proposerIndex, mockProof);
        assertTrue(result, "Should verify the proposer index in fuzz test");
    }
    
    // Helper function that mirrors the contract's SSZ merkle proof computation
    // This ensures our expected root matches what the contract will calculate
    function _computeExpectedRoot(
        uint8 index, 
        bytes32 leaf, 
        bytes32[] memory proof
    ) internal pure returns (bytes32) {
        bytes32 computedHash = leaf;
        uint256 idx = index;
        
        for (uint256 i = 0; i < proof.length; i++) {
            bytes32 proofElement = proof[i];
            
            if (idx % 2 == 0) {
                computedHash = sha256(abi.encodePacked(computedHash, proofElement));
            } else {
                computedHash = sha256(abi.encodePacked(proofElement, computedHash));
            }
            
            idx = idx / 2;
        }
        
        return computedHash;
    }
}