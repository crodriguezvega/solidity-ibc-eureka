// SPDX-License-Identifier: UNLICENSED
pragma solidity ^0.8.28;

// solhint-disable custom-errors,max-line-length

import { Test } from "forge-std/Test.sol";

import { IICS20Transfer } from "../../contracts/interfaces/IICS20Transfer.sol";
import { IICS20TransferMsgs } from "../../contracts/msgs/IICS20TransferMsgs.sol";
import { IERC20Errors } from "@openzeppelin-contracts/interfaces/draft-IERC6093.sol";
import { IBCERC20 } from "../../contracts/utils/IBCERC20.sol";
import { Escrow } from "../../contracts/utils/Escrow.sol";
import { ICS20Lib } from "../../contracts/utils/ICS20Lib.sol";
import { DummyICS20Transfer } from "./mocks/DummyICS20Transfer.sol";

contract IBCERC20Test is Test, DummyICS20Transfer {
    IBCERC20 public ibcERC20;
    Escrow public _escrow;
    ICS20Lib.Denom public denom;

    function setUp() public {
        _escrow = new Escrow(address(this));
        denom.base = "test";
        denom.trace.push(IICS20TransferMsgs.Hop({ portId: "testport", clientId: "channel-42" }));
        ibcERC20 = new IBCERC20(IICS20Transfer(this), _escrow, denom);
    }

    function test_IBCERC20Validation() public {
        IICS20TransferMsgs.Hop[] memory trace = new IICS20TransferMsgs.Hop[](1);
        trace[0] = IICS20TransferMsgs.Hop({ portId: "testport", clientId: "client-42" });
        ICS20Lib.Denom memory testDenom = ICS20Lib.Denom("test", trace);

        // success: valid denom
        new IBCERC20(IICS20Transfer(this), _escrow, testDenom);

        // failure: empty base denom
        testDenom.base = "";
        vm.expectRevert(abi.encodeWithSelector(IBCERC20.IBCERC20InvalidDenom.selector, testDenom));
        new IBCERC20(IICS20Transfer(this), _escrow, testDenom);
        // reset
        testDenom.base = "test";

        // failure: empty trace
        testDenom.trace = new IICS20TransferMsgs.Hop[](0);
        vm.expectRevert(abi.encodeWithSelector(IBCERC20.IBCERC20InvalidDenom.selector, testDenom));
        new IBCERC20(IICS20Transfer(this), _escrow, testDenom);
    }

    function test_ERC20Metadata() public view {
        ICS20Lib.Denom memory fullDenom = ibcERC20.fullDenom();
        string memory path = ICS20Lib.getPath(fullDenom);
        assertEq(path, "testport/channel-42/test");

        assertEq(ibcERC20.ICS20(), address(this));
        assertEq(ibcERC20.ESCROW(), address(_escrow));
        assertEq(ibcERC20.name(), path);
        assertEq(ibcERC20.symbol(), denom.base);
        assertEq(0, ibcERC20.totalSupply());
    }

    function testFuzz_success_Mint(uint256 amount) public {
        ibcERC20.mint(amount);
        assertEq(ibcERC20.balanceOf(address(_escrow)), amount);
        assertEq(ibcERC20.totalSupply(), amount);
    }

    // Just to document the behaviour
    function test_MintZero() public {
        ibcERC20.mint(0);
        assertEq(ibcERC20.balanceOf(address(_escrow)), 0);
        assertEq(ibcERC20.totalSupply(), 0);
    }

    function testFuzz_unauthorized_Mint(uint256 amount) public {
        address notICS20Transfer = makeAddr("notICS20Transfer");
        vm.expectRevert(abi.encodeWithSelector(IBCERC20.IBCERC20Unauthorized.selector, notICS20Transfer));
        vm.prank(notICS20Transfer);
        ibcERC20.mint(amount);
        assertEq(ibcERC20.balanceOf(notICS20Transfer), 0);
        assertEq(ibcERC20.balanceOf(address(_escrow)), 0);
        assertEq(ibcERC20.totalSupply(), 0);
    }

    function testFuzz_success_Burn(uint256 startingAmount, uint256 burnAmount) public {
        burnAmount = bound(burnAmount, 0, startingAmount);
        ibcERC20.mint(startingAmount);
        assertEq(ibcERC20.balanceOf(address(_escrow)), startingAmount);

        ibcERC20.burn(burnAmount);
        uint256 leftOver = startingAmount - burnAmount;
        assertEq(ibcERC20.balanceOf(address(_escrow)), leftOver);
        assertEq(ibcERC20.totalSupply(), leftOver);

        if (leftOver != 0) {
            ibcERC20.burn(leftOver);
            assertEq(ibcERC20.balanceOf(address(_escrow)), 0);
            assertEq(ibcERC20.totalSupply(), 0);
        }
    }

    function testFuzz_unauthorized_Burn(uint256 startingAmount, uint256 burnAmount) public {
        burnAmount = bound(burnAmount, 0, startingAmount);
        ibcERC20.mint(startingAmount);
        assertEq(ibcERC20.balanceOf(address(_escrow)), startingAmount);

        address notICS20Transfer = makeAddr("notICS20Transfer");
        vm.expectRevert(abi.encodeWithSelector(IBCERC20.IBCERC20Unauthorized.selector, notICS20Transfer));
        vm.prank(notICS20Transfer);
        ibcERC20.burn(burnAmount);
        assertEq(ibcERC20.balanceOf(notICS20Transfer), 0);
        assertEq(ibcERC20.balanceOf(address(_escrow)), startingAmount);
        assertEq(ibcERC20.totalSupply(), startingAmount);
    }

    // Just to document the behaviour
    function test_BurnZero() public {
        ibcERC20.burn(0);
        assertEq(ibcERC20.balanceOf(address(_escrow)), 0);
        assertEq(ibcERC20.totalSupply(), 0);

        ibcERC20.mint(1000);
        ibcERC20.burn(0);
        assertEq(ibcERC20.balanceOf(address(_escrow)), 1000);
        assertEq(ibcERC20.totalSupply(), 1000);
    }

    function test_failure_Burn() public {
        // test burn with zero balance
        vm.expectRevert(abi.encodeWithSelector(IERC20Errors.ERC20InsufficientBalance.selector, address(_escrow), 0, 1));
        ibcERC20.burn(1);

        // mint some to test other cases
        ibcERC20.mint(1000);

        // test burn with insufficient balance
        vm.expectRevert(
            abi.encodeWithSelector(IERC20Errors.ERC20InsufficientBalance.selector, address(_escrow), 1000, 1001)
        );
        ibcERC20.burn(1001);
    }
}
