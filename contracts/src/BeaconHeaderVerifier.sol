// SPDX-License-Identifier: MIT
pragma solidity ^0.8.26;

/**
 * @title BeaconHeaderVerifier
 * @dev Contract to verify BeaconBlockHeader fields using the beacon roots contract
 */
contract BeaconHeaderVerifier {
    // Address of the Beacon Roots contract as specified in EIP-4788
    address constant BEACON_ROOTS_ADDRESS = 0x000F3df6D732807Ef1319fB7B8bB8522d0Beac02;
    
    // BeaconBlockHeader field indices
    uint8 constant SLOT_INDEX = 0;
    uint8 constant PROPOSER_INDEX = 1;
    uint8 constant PARENT_ROOT_INDEX = 2;
    uint8 constant STATE_ROOT_INDEX = 3;
    uint8 constant BODY_ROOT_INDEX = 4;
        
    /**
     * @dev Get the beacon block root from the EIP-4788 beacon roots contract
     * @param timestamp The timestamp of the beacon block
     * @return The beacon block root
     */
    function getBeaconBlockRoot(
        uint256 timestamp
    ) public view returns (bytes32) {
        (bool success, bytes memory returnData) = BEACON_ROOTS_ADDRESS
            .staticcall(abi.encode(timestamp));

        require(success, "Call to beacon roots contract failed");
        require(
            returnData.length == 32,
            "Invalid response from beacon roots contract"
        );

        return abi.decode(returnData, (bytes32));
    }
    
    /**
     * @dev Verifies a field in a BeaconBlockHeader
     * @param beaconTimestamp The timestamp of the beacon block
     * @param fieldIndex The index of the field in the BeaconBlockHeader
     * @param fieldValue The value of the field to verify
     * @param merkleProof The merkle proof from the field to the block root
     * @return True if the proof is valid
     */
    function verifyHeaderField(
        uint256 beaconTimestamp,
        uint8 fieldIndex,
        bytes32 fieldValue,
        bytes32[] calldata merkleProof
    ) public view returns (bool) {
        require(fieldIndex <= BODY_ROOT_INDEX, "Invalid field index");
        
        // Get the beacon block root from the EIP-4788 beacon roots contract
        bytes32 beaconBlockRoot = getBeaconBlockRoot(beaconTimestamp);
        
        // Verify the merkle proof
        return verifySSZMerkleProof(beaconBlockRoot, fieldIndex, fieldValue, merkleProof);
    }
    
    /**
     * @dev Verifies the slot of a beacon block header
     * @param beaconTimestamp The timestamp of the beacon block
     * @param slot The slot value to verify
     * @param merkleProof The merkle proof
     * @return True if the proof is valid
     */
    function verifySlot(
        uint256 beaconTimestamp,
        uint64 slot,
        bytes32[] calldata merkleProof
    ) public view returns (bool) {
        // Convert slot to bytes32, SSZ uses little-endian encoding for integers
        bytes32 slotBytes = bytes32(uint256(slot));
        
        return verifyHeaderField(beaconTimestamp, SLOT_INDEX, slotBytes, merkleProof);
    }
    
    /**
     * @dev Verifies the proposer index of a beacon block header
     * @param beaconTimestamp The timestamp of the beacon block
     * @param proposerIndex The proposer index value to verify
     * @param merkleProof The merkle proof
     * @return True if the proof is valid
     */
    function verifyProposerIndex(
        uint256 beaconTimestamp,
        uint64 proposerIndex,
        bytes32[] calldata merkleProof
    ) public view returns (bool) {
        // Convert proposer index to bytes32, SSZ uses little-endian encoding for integers
        bytes32 proposerIndexBytes = bytes32(uint256(proposerIndex));
        
        return verifyHeaderField(beaconTimestamp, PROPOSER_INDEX, proposerIndexBytes, merkleProof);
    }
    
    /**
     * @dev Verifies the parent root of a beacon block header
     * @param beaconTimestamp The timestamp of the beacon block
     * @param parentRoot The parent root value to verify
     * @param merkleProof The merkle proof
     * @return True if the proof is valid
     */
    function verifyParentRoot(
        uint256 beaconTimestamp,
        bytes32 parentRoot,
        bytes32[] calldata merkleProof
    ) public view returns (bool) {
        return verifyHeaderField(beaconTimestamp, PARENT_ROOT_INDEX, parentRoot, merkleProof);
    }
    
    /**
     * @dev Verifies the state root of a beacon block header
     * @param beaconTimestamp The timestamp of the beacon block
     * @param stateRoot The state root value to verify
     * @param merkleProof The merkle proof
     * @return True if the proof is valid
     */
    function verifyStateRoot(
        uint256 beaconTimestamp,
        bytes32 stateRoot,
        bytes32[] calldata merkleProof
    ) public view returns (bool) {
        return verifyHeaderField(beaconTimestamp, STATE_ROOT_INDEX, stateRoot, merkleProof);
    }
    
    /**
     * @dev Verifies the body root of a beacon block header
     * @param beaconTimestamp The timestamp of the beacon block
     * @param bodyRoot The body root value to verify
     * @param merkleProof The merkle proof
     * @return True if the proof is valid
     */
    function verifyBodyRoot(
        uint256 beaconTimestamp,
        bytes32 bodyRoot,
        bytes32[] calldata merkleProof
    ) public view returns (bool) {
        return verifyHeaderField(beaconTimestamp, BODY_ROOT_INDEX, bodyRoot, merkleProof);
    }
    
    /**
     * @dev Compute the hash of two nodes in a merkle tree
     * @param left The left node
     * @param right The right node
     * @return The parent hash
     */
    function hashPair(bytes32 left, bytes32 right) internal pure returns (bytes32) {
        return sha256(abi.encodePacked(left, right));
    }
    
    /**
     * @dev Verifies an SSZ merkle proof against a known root
     * @param root The merkle root to verify against
     * @param index The index of the leaf in the tree
     * @param leaf The leaf value
     * @param proof The merkle proof (sibling nodes from leaf to root)
     * @return True if the proof is valid
     */
    function verifySSZMerkleProof(
        bytes32 root,
        uint256 index,
        bytes32 leaf,
        bytes32[] calldata proof
    ) internal pure returns (bool) {
        bytes32 computedHash = leaf;
        
        for (uint256 i = 0; i < proof.length; i++) {
            bytes32 proofElement = proof[i];
            
            if (index % 2 == 0) {
                computedHash = hashPair(computedHash, proofElement);
            } else {
                computedHash = hashPair(proofElement, computedHash);
            }
            
            index = index / 2;
        }
        
        return computedHash == root;
    }
}