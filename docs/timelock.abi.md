# timelock abi

- [timelock abi](#timelock-abi)
  - [Events](#events)
  - [Functions](#functions)
    - [Role](#role)
    - [EIP Support](#eip-support)
    - [Operations methods](#operations-methods)
    - [Operations Management](#operations-management)
    - [MinDelay Management](#mindelay-management)

## Events

```text
event FunctionSelectorBlocked(bytes4 indexed selector)
event FunctionSelectorUnblocked(bytes4 indexed selector)
event MinDelayChange(uint256 oldDuration, uint256 newDuration)
event RoleGranted(bytes32 indexed role, address indexed account, address indexed sender)
event RoleRevoked(bytes32 indexed role, address indexed account, address indexed sender)
event CallExecuted(bytes32 indexed id, uint256 indexed index, address target, uint256 value, bytes data)
event CallScheduled(bytes32 indexed id, uint256 indexed index, address target, uint256 value, bytes data, bytes32 predecessor, bytes32 salt, uint256 delay)
event Cancelled(bytes32 indexed id)
event RoleAdminChanged(bytes32 indexed role, bytes32 indexed previousAdminRole, bytes32 indexed newAdminRole)
event BypasserCallExecuted(uint256 indexed index, address target, uint256 value, bytes data)
```

## Functions

```text
function DEFAULT_ADMIN_ROLE() view returns(bytes32)
function getRoleMember(bytes32 role, uint256 index) view returns(address)
function renounceRole(bytes32 role, address account) returns()
function isOperationPending(bytes32 id) view returns(bool pending)
function getBlockedFunctionSelectorCount() view returns(uint256)
function getMinDelay() view returns(uint256 duration)
function getRoleAdmin(bytes32 role) view returns(bytes32)
function getRoleMemberCount(bytes32 role) view returns(uint256)
function grantRole(bytes32 role, address account) returns()
function hashOperationBatch((address,uint256,bytes)[] calls, bytes32 predecessor, bytes32 salt) pure returns(bytes32 hash)
function isOperationDone(bytes32 id) view returns(bool done)
function updateDelay(uint256 newDelay) returns()
function EXECUTOR_ROLE() view returns(bytes32)
function executeBatch((address,uint256,bytes)[] calls, bytes32 predecessor, bytes32 salt) payable returns()
function getBlockedFunctionSelectorAt(uint256 index) view returns(bytes4)
function scheduleBatch((address,uint256,bytes)[] calls, bytes32 predecessor, bytes32 salt, uint256 delay) returns()
function cancel(bytes32 id) returns()
function hasRole(bytes32 role, address account) view returns(bool)
function onERC721Received(address , address , uint256 , bytes ) returns(bytes4)
function PROPOSER_ROLE() view returns(bytes32)
function blockFunctionSelector(bytes4 selector) returns()
function bypasserExecuteBatch((address,uint256,bytes)[] calls) payable returns()
function onERC1155Received(address , address , uint256 , uint256 , bytes ) returns(bytes4)
function revokeRole(bytes32 role, address account) returns()
function supportsInterface(bytes4 interfaceId) view returns(bool)
function ADMIN_ROLE() view returns(bytes32)
function isOperation(bytes32 id) view returns(bool registered)
function isOperationReady(bytes32 id) view returns(bool ready)
function unblockFunctionSelector(bytes4 selector) returns()
function BYPASSER_ROLE() view returns(bytes32)
function CANCELLER_ROLE() view returns(bytes32)
function getTimestamp(bytes32 id) view returns(uint256 timestamp)
```
