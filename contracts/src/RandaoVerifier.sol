// SPDX-License-Identifier: MIT
pragma solidity ^0.8.26;

import {SSZMerkleProof} from "./lib/ssz.sol";

/**
 * @title BeaconRANDAOVerifier
 * @dev Contract to verify RANDAO values from the beacon chain
 */
contract BeaconRANDAOVerifier {
    // Address of the Beacon Roots contract (EIP-4788)
    address public immutable BEACON_ROOTS_ADDRESS;

    // Constants for verification
    uint256 private constant RANDAO_MIXES_INDEX = 13; // Index of randao_mixes in BeaconState

    /**
     * @dev Constructor
     * @param _beaconRootsAddress Address of the Beacon Roots contract
     */
    constructor(address _beaconRootsAddress) {
        BEACON_ROOTS_ADDRESS = _beaconRootsAddress;
    }

    /**
     * @dev Gets the beacon block root for a timestamp from the EIP-4788 contract
     * @param timestamp The timestamp to query
     * @return The beacon block root for the timestamp
     */
    function getBeaconBlockRoot(
        uint256 timestamp
    ) public view returns (bytes32) {
        // Format the timestamp as a 32-byte big-endian value
        bytes32 timestampBytes = bytes32(timestamp);

        // Call the beacon roots contract
        (bool success, bytes memory returnData) = BEACON_ROOTS_ADDRESS
            .staticcall(abi.encodePacked(timestampBytes));

        require(success, "Call to beacon roots contract failed");
        require(
            returnData.length == 32,
            "Invalid response from beacon roots contract"
        );

        return abi.decode(returnData, (bytes32));
    }

    /**
     * @dev Get RANDAO value for a specific slot, verified against beacon roots
     * @param timestamp The timestamp corresponding to the slot
     * @param randaoMix The claimed RANDAO mix value
     * @param merkleProof The Merkle branch proof
     * @param generalizedIndex The generalized index of the RANDAO mix
     * @return The verified RANDAO mix value
     */
    function getVerifiedRANDAO(
        uint256 timestamp,
        bytes32 randaoMix,
        bytes32[] calldata merkleProof,
        uint64 generalizedIndex
    ) external view returns (bytes32) {
        bool isValid = verifyRANDAOMix(
            timestamp,
            randaoMix,
            merkleProof,
            generalizedIndex
        );

        require(isValid, "Invalid RANDAO proof");
        return randaoMix;
    }

    /**
     * @dev Verifies a RANDAO mix value using a Merkle proof against the beacon root
     * @param timestamp The timestamp corresponding to the slot (used to query the beacon roots contract)
     * @param randaoMix The RANDAO mix value to verify
     * @param merkleProof The Merkle branch proof
     * @param generalizedIndex The generalized index of the RANDAO mix in the beacon state
     * @return True if the RANDAO mix is valid, false otherwise
     */
    function verifyRANDAOMix(
        // uint64 slot,
        uint256 timestamp,
        bytes32 randaoMix,
        bytes32[] calldata merkleProof,
        uint64 generalizedIndex
    ) public view returns (bool) {
        // Get the beacon block root from the Beacon Roots contract
        bytes32 beaconRoot = getBeaconBlockRoot(timestamp);

        // Verify the beacon root is not empty
        require(
            beaconRoot != bytes32(0),
            "Beacon root not available for timestamp"
        );

        // Verify the SSZ Merkle proof
        return
            SSZMerkleProof._verify(
                merkleProof,
                beaconRoot,
                randaoMix,
                generalizedIndex
            );
    }

    function debugVerifyRANDAOMix(
        uint256 timestamp,
        bytes32 randaoMix,
        bytes32[] calldata merkleProof,
        uint64 generalizedIndex
    )
        external
        view
        returns (bytes32 beaconRoot, bytes32 computedRoot, bool isValid)
    {
        // Get the beacon block root from the Beacon Roots contract
        beaconRoot = getBeaconBlockRoot(timestamp);

        // Compute root from the proof
        bytes32 node = randaoMix;
        uint64 index = generalizedIndex;

        for (uint256 i = 0; i < merkleProof.length; i++) {
            bool isLeft = (index & 1) == 0;
            index = index / 2;

            if (isLeft) {
                node = sha256(abi.encodePacked(node, merkleProof[i]));
            } else {
                node = sha256(abi.encodePacked(merkleProof[i], node));
            }
        }

        computedRoot = node;
        isValid = (node == beaconRoot);

        return (beaconRoot, computedRoot, isValid);
    }

    /**
     * @dev Verifies an SSZ Merkle proof
     * @param root The Merkle root
     * @param leaf The leaf value
     * @param proof The Merkle branch proof
     * @param index The generalized index of the leaf
     * @return True if the proof is valid, false otherwise
     */
    // function verifySSZMerkleProof(
    //     bytes32 root,
    //     bytes32 leaf,
    //     bytes32[] calldata proof,
    //     uint64 index
    // ) internal pure returns (bool) {
    //     // Start with the leaf node
    //     bytes32 node = leaf;

    //     // Iterate through each proof element
    //     for (uint256 i = 0; i < proof.length; i++) {
    //         // Determine if we're dealing with a left or right node
    //         bool isLeft = (index & 1) == 0;
    //         index = index / 2;

    //         // Combine the node with the proof element according to its position
    //         if (isLeft) {
    //             node = sha256(abi.encodePacked(node, proof[i]));
    //         } else {
    //             node = sha256(abi.encodePacked(proof[i], node));
    //         }
    //     }

    //     // Check if the computed root matches the expected root
    //     return node == root;
    // }

    /**
     * @dev Calculates the slot timestamp
     * @param slot The slot number
     * @param genesisTime The genesis time of the beacon chain
     * @return The timestamp for the slot
     */
    function calculateSlotTimestamp(
        uint64 slot,
        uint64 genesisTime
    ) public pure returns (uint256) {
        uint64 SECONDS_PER_SLOT = 12;
        return genesisTime + (slot * SECONDS_PER_SLOT);
    }

    /**
     * @dev Calculates the epoch for a slot
     * @param slot The slot number
     * @return The epoch number
     */
    function slotToEpoch(uint64 slot) public pure returns (uint64) {
        uint64 SLOTS_PER_EPOCH = 32;
        return slot / SLOTS_PER_EPOCH;
    }

    /**
     * @dev Calculates the index in RANDAO_MIXES for an epoch
     * @param epoch The epoch number
     * @return The index in the RANDAO_MIXES array
     */
    function epochToRandaoIndex(uint64 epoch) public pure returns (uint64) {
        uint64 EPOCHS_PER_HISTORICAL_VECTOR = 65536; // 2^16
        return epoch % EPOCHS_PER_HISTORICAL_VECTOR;
    }

    /**
     * @dev Calculates the generalized index for a RANDAO mix
     * @param randaoIndex The index in the RANDAO_MIXES array
     * @return The generalized index in the beacon state tree
     */
    function calculateGeneralizedIndex(
        uint64 randaoIndex
    ) public pure returns (uint64) {
        // This is a simplified calculation and may need adjustment based on
        // the exact tree structure of the beacon state
        uint64 TREE_DEPTH = 5; // This value depends on the state structure
        return uint64((1 << TREE_DEPTH) * RANDAO_MIXES_INDEX + randaoIndex);
    }
}
