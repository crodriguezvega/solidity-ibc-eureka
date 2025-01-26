// SPDX-License-Identifier: MIT
pragma solidity ^0.8.28;

import { ERC20 } from "@openzeppelin-contracts/token/ERC20/ERC20.sol";
import { IICS20Transfer } from "../interfaces/IICS20Transfer.sol";
import { IIBCERC20 } from "../interfaces/IIBCERC20.sol";
import { IEscrow } from "../interfaces/IEscrow.sol";
import { ICS20Lib } from "../utils/ICS20Lib.sol";

contract IBCERC20 is IIBCERC20, ERC20 {
    /// @notice The full IBC denom path for this token
    ICS20Lib.Denom private _denom;
    /// @notice The escrow contract address
    address public immutable ESCROW;
    /// @notice The ICS20 contract address
    address public immutable ICS20;

    /// @notice Unauthorized function call
    /// @param caller The caller of the function
    error IBCERC20Unauthorized(address caller);

    constructor(
        IICS20Transfer ics20_,
        IEscrow escrow_,
        bytes32 denomID_,
        ICS20Lib.Denom memory denom_
    )
        // TODO: Was there something I was supposed to be using instead of encodePacked?
        ERC20(string(abi.encodePacked(denomID_)), denom_.base)
    {
        _denom = denom_;
        ESCROW = address(escrow_);
        ICS20 = address(ics20_);
    }

    /// @inheritdoc IIBCERC20
    function fullDenom() public view returns (ICS20Lib.Denom memory) {
        return _denom;
    }

    /// @inheritdoc IIBCERC20
    function mint(uint256 amount) external onlyICS20 {
        _mint(ESCROW, amount);
    }

    /// @inheritdoc IIBCERC20
    function burn(uint256 amount) external onlyICS20 {
        _burn(ESCROW, amount);
    }

    modifier onlyICS20() {
        require(_msgSender() == ICS20, IBCERC20Unauthorized(_msgSender()));
        _;
    }
}
