pragma solidity 0.8.15;

import { Test } from "forge-std/Test.sol";
import { SystemConfig } from "../../L1/SystemConfig.sol";
import { OptimismPortal } from "../../L1/OptimismPortal.sol";
import { L2OutputOracle } from "../../L1/L2OutputOracle.sol";

contract SystemConfig_GasLimitLowerBound_Invariant is Test {
    SystemConfig public config;
    OptimismPortal public portal;

    function setUp() public {
        portal = new OptimismPortal({
            _l2Oracle: L2OutputOracle(address(0)),
            _guardian: address(0),
            _paused: true
        });

        config = new SystemConfig({
            _owner: address(0xbeef),
            _overhead: 2100,
            _scalar: 1000000,
            _batcherHash: bytes32(hex"abcd"),
            _gasLimit: 8_000_000,
            _unsafeBlockSigner: address(1),
            _portal: address(portal)
        });

        // Set the target contract to the `config`
        targetContract(address(config));
        // Set the target sender to the `config`'s owner (0xbeef)
        targetSender(address(0xbeef));
        // Set the target selector for `setGasLimit`
        // `setGasLimit` is the only function we care about, as it is the only function
        // that can modify the gas limit within the SystemConfig.
        bytes4[] memory selectors = new bytes4[](1);
        selectors[0] = config.setGasLimit.selector;
        FuzzSelector memory selector = FuzzSelector({
            addr: address(config),
            selectors: selectors
        });
        targetSelector(selector);
    }

    /**
     * @custom:invariant The gas limit of the `SystemConfig` contract can never be lower
     * than the hard-coded lower bound.
     */
    function invariant_gasLimitLowerBound() external {
        uint256 maxResourceLimit = uint256(portal.MAX_RESOURCE_LIMIT());
        assertTrue(config.gasLimit() >= maxResourceLimit);
    }
}
