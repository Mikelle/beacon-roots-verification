// SPDX-License-Identifier: UNLICENSED
pragma solidity ^0.8.26;

import {Script, console} from "forge-std/Script.sol";
import {BeaconHeaderVerifier} from "../src/BeaconHeaderVerifier.sol";

contract HeaderVerifierScript is Script {
    BeaconHeaderVerifier public beaconHeaderVerifier;

    // Beacon Roots contract address is the same on all networks (per EIP-4788)
    address constant BEACON_ROOTS_ADDRESS = 0x000F3df6D732807Ef1319fB7B8bB8522d0Beac02;

    function run() public {
        uint256 deployerPrivateKey = vm.envUint("PRIVATE_KEY");
        vm.startBroadcast(deployerPrivateKey);

        beaconHeaderVerifier = new BeaconHeaderVerifier();
        
        console.log("BeaconHeaderVerifier deployed at:", address(beaconHeaderVerifier));

        vm.stopBroadcast();
    }
}