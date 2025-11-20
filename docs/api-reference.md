# Zindex API Reference

Base URL: `http://localhost:8080`

All endpoints return JSON responses in the following format:
```json
{
  "result": "success",
  "data": { ... }
}
```

Error responses:
```json
{
  "result": "error",
  "error": "error message"
}
```

## Recent Updates

### Enhanced Transaction Data
- **Transaction responses** now include `input_count` and `output_count` fields showing the number of inputs and outputs for each transaction.
- **Account transaction responses** now include `balance_change` field indicating the amount by which the account balance changed (positive for receiving, negative for sending).

### New Features
- **Multiple transaction types**: The `GET /api/v1/tx-graph/transactions/by-type` endpoint now supports comma-separated transaction types (e.g., `?type=tze,t2t,t2z`).
- **Count endpoints**: Added count endpoints for all modules with optional filters to get total counts of transactions, outputs, inputs, accounts, verifiers, proofs, and facts.
- **Proof size aggregation**: Added `GET /api/v1/starks/verifier/sum-proof-sizes` endpoint to get the total proof size for a verifier.

## Table of Contents

1. [Blocks Module](#blocks-module)
2. [Transaction Graph Module](#transaction-graph-module)
3. [Accounts Module](#accounts-module)
4. [TZE Graph Module](#tze-graph-module)
5. [STARKS Module](#starks-module)

---

## Blocks Module

### Get All Blocks

`GET /api/v1/blocks`

Retrieves blocks with pagination.

**Query Parameters:**
- `limit` ![optional](https://img.shields.io/badge/-optional-blue) - Number of blocks to return (default: configured pagination limit)
- `offset` ![optional](https://img.shields.io/badge/-optional-blue) - Number of blocks to skip (default: 0)

**Examples:**
```
http://localhost:8080/api/v1/blocks?limit=10&offset=0
http://localhost:8080/api/v1/blocks?limit=50
```

### Get Block by Height

`GET /api/v1/blocks/block`

Retrieves a single block by height.

**Query Parameters:**
- `height` - Block height (required)

**Examples:**
```
http://localhost:8080/api/v1/blocks/block?height=100
http://localhost:8080/api/v1/blocks/block?height=1000
```

### Get Block by Hash

`GET /api/v1/blocks/by-hash`

Retrieves a single block by hash.

**Query Parameters:**
- `hash` - Block hash (required)

**Examples:**
```
http://localhost:8080/api/v1/blocks/by-hash?hash=00000000000000000002d6cca6761c99b3c2e936f9a0e304b7c7651a993f461b
```

### Get Blocks by Height Range

`GET /api/v1/blocks/range`

Retrieves blocks within a height range.

**Query Parameters:**
- `from_height` - Starting block height (required)
- `to_height` - Ending block height (required)
- `limit` ![optional](https://img.shields.io/badge/-optional-blue) - Number of blocks to return
- `offset` ![optional](https://img.shields.io/badge/-optional-blue) - Number of blocks to skip

**Examples:**
```
http://localhost:8080/api/v1/blocks/range?from_height=100&to_height=200
http://localhost:8080/api/v1/blocks/range?from_height=1000&to_height=1500&limit=50
```

### Get Blocks by Timestamp Range

`GET /api/v1/blocks/timestamp-range`

Retrieves blocks within a timestamp range.

**Query Parameters:**
- `from_timestamp` - Starting Unix timestamp (required)
- `to_timestamp` - Ending Unix timestamp (required)
- `limit` ![optional](https://img.shields.io/badge/-optional-blue) - Number of blocks to return
- `offset` ![optional](https://img.shields.io/badge/-optional-blue) - Number of blocks to skip

**Examples:**
```
http://localhost:8080/api/v1/blocks/timestamp-range?from_timestamp=1609459200&to_timestamp=1640995200
http://localhost:8080/api/v1/blocks/timestamp-range?from_timestamp=1609459200&to_timestamp=1640995200&limit=20
```

### Get Recent Blocks

`GET /api/v1/blocks/recent`

Retrieves the most recent blocks.

**Query Parameters:**
- `limit` ![optional](https://img.shields.io/badge/-optional-blue) - Number of blocks to return (default: configured pagination limit)

**Examples:**
```
http://localhost:8080/api/v1/blocks/recent?limit=10
http://localhost:8080/api/v1/blocks/recent
```

### Get Block Count

`GET /api/v1/blocks/count`

Returns the total number of blocks.

**Query Parameters:** None

**Examples:**
```
http://localhost:8080/api/v1/blocks/count
```

### Get Latest Block

`GET /api/v1/blocks/latest`

Retrieves the most recent block.

**Query Parameters:** None

**Examples:**
```
http://localhost:8080/api/v1/blocks/latest
```

---

## Transaction Graph Module

> **Note:** This module must be enabled in configuration to use these endpoints.

### Transactions

#### Get Transaction

`GET /api/v1/tx-graph/transaction`

Retrieves a single transaction by txid.

**Query Parameters:**
- `txid` - Transaction ID (required)

**Examples:**
```
http://localhost:8080/api/v1/tx-graph/transaction?txid=abc123def456
```

#### Get Transactions by Block

`GET /api/v1/tx-graph/transactions/by-block`

Retrieves all transactions in a specific block.

**Query Parameters:**
- `block_height` - Block height (required)

**Examples:**
```
http://localhost:8080/api/v1/tx-graph/transactions/by-block?block_height=100
http://localhost:8080/api/v1/tx-graph/transactions/by-block?block_height=1500
```

#### Get Transactions by Type

`GET /api/v1/tx-graph/transactions/by-type`

Retrieves transactions filtered by type(s) with pagination. Supports multiple comma-separated types.

**Query Parameters:**
- `type` - Transaction type(s): `coinbase`, `tze`, `t2t`, `t2z`, `z2t`, `z2z` (required). Multiple types can be specified as comma-separated values.
- `limit` ![optional](https://img.shields.io/badge/-optional-blue) - Number of transactions to return
- `offset` ![optional](https://img.shields.io/badge/-optional-blue) - Number of transactions to skip

**Examples:**
```
# Single type
http://localhost:8080/api/v1/tx-graph/transactions/by-type?type=coinbase&limit=10

# Multiple types (comma-separated)
http://localhost:8080/api/v1/tx-graph/transactions/by-type?type=tze,t2t,t2z&limit=50&offset=0
http://localhost:8080/api/v1/tx-graph/transactions/by-type?type=coinbase,tze
```

#### Get Recent Transactions

`GET /api/v1/tx-graph/transactions/recent`

Retrieves the most recent transactions with pagination.

**Query Parameters:**
- `limit` ![optional](https://img.shields.io/badge/-optional-blue) - Number of transactions to return
- `offset` ![optional](https://img.shields.io/badge/-optional-blue) - Number of transactions to skip

**Examples:**
```
http://localhost:8080/api/v1/tx-graph/transactions/recent?limit=10
http://localhost:8080/api/v1/tx-graph/transactions/recent?limit=20&offset=10
```

### Outputs

#### Get Transaction Outputs

`GET /api/v1/tx-graph/outputs`

Retrieves all outputs for a transaction.

**Query Parameters:**
- `txid` - Transaction ID (required)

**Examples:**
```
http://localhost:8080/api/v1/tx-graph/outputs?txid=abc123def456
```

#### Get Transaction Output

`GET /api/v1/tx-graph/outputs/output`

Retrieves a specific output by txid and vout.

**Query Parameters:**
- `txid` - Transaction ID (required)
- `vout` - Output index (required)

**Examples:**
```
http://localhost:8080/api/v1/tx-graph/outputs/output?txid=abc123def456&vout=0
http://localhost:8080/api/v1/tx-graph/outputs/output?txid=abc123def456&vout=2
```

#### Get Unspent Outputs

`GET /api/v1/tx-graph/outputs/unspent`

Retrieves all unspent outputs for a transaction.

**Query Parameters:**
- `txid` - Transaction ID (required)

**Examples:**
```
http://localhost:8080/api/v1/tx-graph/outputs/unspent?txid=abc123def456
```

#### Get Output Spenders

`GET /api/v1/tx-graph/outputs/spenders`

Retrieves all transactions that spent outputs from a given transaction.

**Query Parameters:**
- `txid` - Transaction ID (required)

**Examples:**
```
http://localhost:8080/api/v1/tx-graph/outputs/spenders?txid=abc123def456
```

### Inputs

#### Get Transaction Inputs

`GET /api/v1/tx-graph/inputs`

Retrieves all inputs for a transaction.

**Query Parameters:**
- `txid` - Transaction ID (required)

**Examples:**
```
http://localhost:8080/api/v1/tx-graph/inputs?txid=abc123def456
```

#### Get Transaction Input

`GET /api/v1/tx-graph/inputs/input`

Retrieves a specific input by txid and vin.

**Query Parameters:**
- `txid` - Transaction ID (required)
- `vin` - Input index (required)

**Examples:**
```
http://localhost:8080/api/v1/tx-graph/inputs/input?txid=abc123def456&vin=0
http://localhost:8080/api/v1/tx-graph/inputs/input?txid=abc123def456&vin=1
```

#### Get Input Sources

`GET /api/v1/tx-graph/inputs/sources`

Retrieves all transactions that provided inputs to a given transaction.

**Query Parameters:**
- `txid` - Transaction ID (required)

**Examples:**
```
http://localhost:8080/api/v1/tx-graph/inputs/sources?txid=abc123def456
```

### Count

#### Count Transactions

`GET /api/v1/tx-graph/transactions/count`

Returns the total count of transactions with optional filters.

**Query Parameters:**
- `type` ![optional](https://img.shields.io/badge/-optional-blue) - Filter by transaction type: `coinbase`, `tze`, `t2t`, `t2z`, `z2t`, `z2z`
- `block_height` ![optional](https://img.shields.io/badge/-optional-blue) - Filter by block height

**Response:**
```json
{
  "result": "success",
  "data": {
    "count": 1234
  }
}
```

**Examples:**
```
# Total count of all transactions
http://localhost:8080/api/v1/tx-graph/transactions/count

# Count by type
http://localhost:8080/api/v1/tx-graph/transactions/count?type=coinbase

# Count by block height
http://localhost:8080/api/v1/tx-graph/transactions/count?block_height=1000

# Count by type and block height
http://localhost:8080/api/v1/tx-graph/transactions/count?type=t2z&block_height=1500
```

#### Count Transaction Outputs

`GET /api/v1/tx-graph/outputs/count`

Returns the total count of transaction outputs with optional filters.

**Query Parameters:**
- `txid` ![optional](https://img.shields.io/badge/-optional-blue) - Filter by transaction ID
- `spent` ![optional](https://img.shields.io/badge/-optional-blue) - Filter by spent status (`true` for spent outputs only)

**Response:**
```json
{
  "result": "success",
  "data": {
    "count": 5678
  }
}
```

**Examples:**
```
# Total count of all outputs
http://localhost:8080/api/v1/tx-graph/outputs/count

# Count outputs for a specific transaction
http://localhost:8080/api/v1/tx-graph/outputs/count?txid=abc123def456

# Count spent outputs only
http://localhost:8080/api/v1/tx-graph/outputs/count?spent=true

# Count spent outputs for a transaction
http://localhost:8080/api/v1/tx-graph/outputs/count?txid=abc123def456&spent=true
```

#### Count Transaction Inputs

`GET /api/v1/tx-graph/inputs/count`

Returns the total count of transaction inputs with optional filters.

**Query Parameters:**
- `txid` ![optional](https://img.shields.io/badge/-optional-blue) - Filter by transaction ID

**Response:**
```json
{
  "result": "success",
  "data": {
    "count": 9012
  }
}
```

**Examples:**
```
# Total count of all inputs
http://localhost:8080/api/v1/tx-graph/inputs/count

# Count inputs for a specific transaction
http://localhost:8080/api/v1/tx-graph/inputs/count?txid=abc123def456
```

### Graph

#### Get Transaction Graph

`GET /api/v1/tx-graph/graph`

Builds a graph of connected transactions up to a specified depth.

**Query Parameters:**
- `txid` - Transaction ID (required)
- `depth` ![optional](https://img.shields.io/badge/-optional-blue) - Recursion depth (default: 3, capped at configured max_graph_depth)

**Examples:**
```
http://localhost:8080/api/v1/tx-graph/graph?txid=abc123def456
http://localhost:8080/api/v1/tx-graph/graph?txid=abc123def456&depth=5
```

---

## Accounts Module

> **Note:** This module must be enabled in configuration to use these endpoints.

### Accounts

#### Get All Accounts

`GET /api/v1/accounts`

Retrieves all accounts with pagination.

**Query Parameters:**
- `limit` ![optional](https://img.shields.io/badge/-optional-blue) - Number of accounts to return
- `offset` ![optional](https://img.shields.io/badge/-optional-blue) - Number of accounts to skip

**Examples:**
```
http://localhost:8080/api/v1/accounts?limit=10
http://localhost:8080/api/v1/accounts?limit=50&offset=100
```

#### Get Account

`GET /api/v1/accounts/account`

Retrieves a single account by address.

**Query Parameters:**
- `address` - Account address (required)

**Examples:**
```
http://localhost:8080/api/v1/accounts/account?address=t1abc123def456
```

#### Get Accounts by Balance Range

`GET /api/v1/accounts/balance-range`

Retrieves accounts within a specified balance range.

**Query Parameters:**
- `min_balance` ![optional](https://img.shields.io/badge/-optional-blue) - Minimum balance (default: 0)
- `max_balance` - Maximum balance (required)
- `limit` ![optional](https://img.shields.io/badge/-optional-blue) - Number of accounts to return
- `offset` ![optional](https://img.shields.io/badge/-optional-blue) - Number of accounts to skip

**Examples:**
```
http://localhost:8080/api/v1/accounts/balance-range?min_balance=1000&max_balance=10000
http://localhost:8080/api/v1/accounts/balance-range?max_balance=5000&limit=20
```

#### Get Top Accounts by Balance

`GET /api/v1/accounts/top-balances`

Retrieves accounts with the highest balances.

**Query Parameters:**
- `limit` ![optional](https://img.shields.io/badge/-optional-blue) - Number of accounts to return

**Examples:**
```
http://localhost:8080/api/v1/accounts/top-balances?limit=10
http://localhost:8080/api/v1/accounts/top-balances?limit=100
```

#### Get Recent Active Accounts

`GET /api/v1/accounts/recent-active`

Retrieves accounts with recent transaction activity.

**Query Parameters:**
- `limit` ![optional](https://img.shields.io/badge/-optional-blue) - Number of accounts to return

**Examples:**
```
http://localhost:8080/api/v1/accounts/recent-active?limit=20
http://localhost:8080/api/v1/accounts/recent-active
```

### Account Transactions

#### Get Account Transactions

`GET /api/v1/accounts/transactions`

Retrieves all transactions for a specific account.

**Query Parameters:**
- `address` - Account address (required)
- `limit` ![optional](https://img.shields.io/badge/-optional-blue) - Number of transactions to return
- `offset` ![optional](https://img.shields.io/badge/-optional-blue) - Number of transactions to skip

**Examples:**
```
http://localhost:8080/api/v1/accounts/transactions?address=t1abc123def456&limit=10
http://localhost:8080/api/v1/accounts/transactions?address=t1abc123def456&limit=50&offset=20
```

#### Get Account Transactions by Type

`GET /api/v1/accounts/transactions/type`

Retrieves transactions for an account filtered by type.

**Query Parameters:**
- `address` - Account address (required)
- `type` - Transaction type: `receive`, `send` (required)
- `limit` ![optional](https://img.shields.io/badge/-optional-blue) - Number of transactions to return
- `offset` ![optional](https://img.shields.io/badge/-optional-blue) - Number of transactions to skip

**Examples:**
```
http://localhost:8080/api/v1/accounts/transactions/type?address=t1abc123def456&type=receive&limit=10
http://localhost:8080/api/v1/accounts/transactions/type?address=t1abc123def456&type=send
```

#### Get Account Receiving Transactions

`GET /api/v1/accounts/transactions/receiving`

Retrieves receiving transactions for an account.

**Query Parameters:**
- `address` - Account address (required)
- `limit` ![optional](https://img.shields.io/badge/-optional-blue) - Number of transactions to return
- `offset` ![optional](https://img.shields.io/badge/-optional-blue) - Number of transactions to skip

**Examples:**
```
http://localhost:8080/api/v1/accounts/transactions/receiving?address=t1abc123def456&limit=10
```

#### Get Account Sending Transactions

`GET /api/v1/accounts/transactions/sending`

Retrieves sending transactions for an account.

**Query Parameters:**
- `address` - Account address (required)
- `limit` ![optional](https://img.shields.io/badge/-optional-blue) - Number of transactions to return
- `offset` ![optional](https://img.shields.io/badge/-optional-blue) - Number of transactions to skip

**Examples:**
```
http://localhost:8080/api/v1/accounts/transactions/sending?address=t1abc123def456&limit=10
```

#### Get Account Transactions by Block Range

`GET /api/v1/accounts/transactions/block-range`

Retrieves transactions for an account within a block range.

**Query Parameters:**
- `address` - Account address (required)
- `from_block` - Starting block height (required)
- `to_block` - Ending block height (required)
- `limit` ![optional](https://img.shields.io/badge/-optional-blue) - Number of transactions to return
- `offset` ![optional](https://img.shields.io/badge/-optional-blue) - Number of transactions to skip

**Examples:**
```
http://localhost:8080/api/v1/accounts/transactions/block-range?address=t1abc123def456&from_block=100&to_block=200
http://localhost:8080/api/v1/accounts/transactions/block-range?address=t1abc123def456&from_block=1000&to_block=1500&limit=20
```

#### Get Account Transaction Count

`GET /api/v1/accounts/transactions/count`

Returns the total number of transactions for an account.

**Query Parameters:**
- `address` - Account address (required)

**Examples:**
```
http://localhost:8080/api/v1/accounts/transactions/count?address=t1abc123def456
```

#### Get Account Transaction

`GET /api/v1/accounts/transactions/transaction`

Retrieves a specific transaction for an account.

**Query Parameters:**
- `address` - Account address (required)
- `txid` - Transaction ID (required)

**Examples:**
```
http://localhost:8080/api/v1/accounts/transactions/transaction?address=t1abc123def456&txid=abc123def456
```

#### Get Transaction Accounts

`GET /api/v1/accounts/transactions/by-txid`

Retrieves all accounts associated with a transaction.

**Query Parameters:**
- `txid` - Transaction ID (required)

**Examples:**
```
http://localhost:8080/api/v1/accounts/transactions/by-txid?txid=abc123def456
```

### Count

#### Count Accounts

`GET /api/v1/accounts/count`

Returns the total count of accounts.

**Response:**
```json
{
  "result": "success",
  "data": {
    "count": 12345
  }
}
```

**Examples:**
```
http://localhost:8080/api/v1/accounts/count
```

#### Count Account Transactions

`GET /api/v1/accounts/transactions/total-count`

Returns the total count of account transactions with optional filters.

**Query Parameters:**
- `address` ![optional](https://img.shields.io/badge/-optional-blue) - Filter by account address
- `type` ![optional](https://img.shields.io/badge/-optional-blue) - Filter by transaction type: `send` or `receive`

**Response:**
```json
{
  "result": "success",
  "data": {
    "count": 6789
  }
}
```

**Examples:**
```
# Total count of all account transactions
http://localhost:8080/api/v1/accounts/transactions/total-count

# Count transactions for a specific address
http://localhost:8080/api/v1/accounts/transactions/total-count?address=t1abc123def456

# Count by transaction type
http://localhost:8080/api/v1/accounts/transactions/total-count?type=receive

# Count by address and type
http://localhost:8080/api/v1/accounts/transactions/total-count?address=t1abc123def456&type=send
```

---

## TZE Graph Module

> **Note:** This module must be enabled in configuration to use these endpoints.

### TZE Inputs

#### Get TZE Inputs

`GET /api/v1/tze-graph/inputs`

Retrieves all TZE inputs for a transaction.

**Query Parameters:**
- `txid` - Transaction ID (required)

**Examples:**
```
http://localhost:8080/api/v1/tze-graph/inputs?txid=abc123def456
```

#### Get TZE Input

`GET /api/v1/tze-graph/inputs/input`

Retrieves a specific TZE input by txid and vin.

**Query Parameters:**
- `txid` - Transaction ID (required)
- `vin` - Input index (required)

**Examples:**
```
http://localhost:8080/api/v1/tze-graph/inputs/input?txid=abc123def456&vin=0
```

#### Get TZE Inputs by Type

`GET /api/v1/tze-graph/inputs/by-type`

Retrieves all inputs of a specific TZE type with pagination.

**Query Parameters:**
- `type` - TZE type: `demo`, `stark_verify` (required)
- `limit` ![optional](https://img.shields.io/badge/-optional-blue) - Number of inputs to return
- `offset` ![optional](https://img.shields.io/badge/-optional-blue) - Number of inputs to skip

**Examples:**
```
http://localhost:8080/api/v1/tze-graph/inputs/by-type?type=demo&limit=10
http://localhost:8080/api/v1/tze-graph/inputs/by-type?type=stark_verify&limit=50
```

#### Get TZE Inputs by Mode

`GET /api/v1/tze-graph/inputs/by-mode`

Retrieves all inputs of a specific TZE mode with pagination.

**Query Parameters:**
- `mode` - TZE mode: `0` or `1` (required)
- `limit` ![optional](https://img.shields.io/badge/-optional-blue) - Number of inputs to return
- `offset` ![optional](https://img.shields.io/badge/-optional-blue) - Number of inputs to skip

**Examples:**
```
http://localhost:8080/api/v1/tze-graph/inputs/by-mode?mode=0&limit=10
http://localhost:8080/api/v1/tze-graph/inputs/by-mode?mode=1
```

#### Get TZE Inputs by Type and Mode

`GET /api/v1/tze-graph/inputs/by-type-mode`

Retrieves all inputs matching both type and mode with pagination.

**Query Parameters:**
- `type` - TZE type: `demo`, `stark_verify` (required)
- `mode` - TZE mode string (required, depends on type):
  - For `demo`: `open`, `close`
  - For `stark_verify`: `initialize`, `verify`
- `limit` ![optional](https://img.shields.io/badge/-optional-blue) - Number of inputs to return
- `offset` ![optional](https://img.shields.io/badge/-optional-blue) - Number of inputs to skip

**Examples:**
```
http://localhost:8080/api/v1/tze-graph/inputs/by-type-mode?type=demo&mode=open&limit=10
http://localhost:8080/api/v1/tze-graph/inputs/by-type-mode?type=stark_verify&mode=verify
```

#### Get TZE Inputs by Previous Output

`GET /api/v1/tze-graph/inputs/by-prev-output`

Retrieves all inputs spending a specific previous output.

**Query Parameters:**
- `prev_txid` - Previous transaction ID (required)
- `prev_vout` - Previous output index (required)

**Examples:**
```
http://localhost:8080/api/v1/tze-graph/inputs/by-prev-output?prev_txid=abc123def456&prev_vout=0
```

### TZE Outputs

#### Get TZE Outputs

`GET /api/v1/tze-graph/outputs`

Retrieves all TZE outputs for a transaction.

**Query Parameters:**
- `txid` - Transaction ID (required)

**Examples:**
```
http://localhost:8080/api/v1/tze-graph/outputs?txid=abc123def456
```

#### Get TZE Output

`GET /api/v1/tze-graph/outputs/output`

Retrieves a specific TZE output by txid and vout.

**Query Parameters:**
- `txid` - Transaction ID (required)
- `vout` - Output index (required)

**Examples:**
```
http://localhost:8080/api/v1/tze-graph/outputs/output?txid=abc123def456&vout=0
```

#### Get Unspent TZE Outputs

`GET /api/v1/tze-graph/outputs/unspent`

Retrieves all unspent TZE outputs for a transaction.

**Query Parameters:**
- `txid` - Transaction ID (required)

**Examples:**
```
http://localhost:8080/api/v1/tze-graph/outputs/unspent?txid=abc123def456
```

#### Get All Unspent TZE Outputs

`GET /api/v1/tze-graph/outputs/all-unspent`

Retrieves all unspent TZE outputs with pagination.

**Query Parameters:**
- `limit` ![optional](https://img.shields.io/badge/-optional-blue) - Number of outputs to return
- `offset` ![optional](https://img.shields.io/badge/-optional-blue) - Number of outputs to skip

**Examples:**
```
http://localhost:8080/api/v1/tze-graph/outputs/all-unspent?limit=10
http://localhost:8080/api/v1/tze-graph/outputs/all-unspent?limit=50&offset=20
```

#### Get TZE Outputs by Type

`GET /api/v1/tze-graph/outputs/by-type`

Retrieves all outputs of a specific TZE type with pagination.

**Query Parameters:**
- `type` - TZE type: `demo`, `stark_verify` (required)
- `limit` ![optional](https://img.shields.io/badge/-optional-blue) - Number of outputs to return
- `offset` ![optional](https://img.shields.io/badge/-optional-blue) - Number of outputs to skip

**Examples:**
```
http://localhost:8080/api/v1/tze-graph/outputs/by-type?type=demo&limit=10
http://localhost:8080/api/v1/tze-graph/outputs/by-type?type=stark_verify
```

#### Get TZE Outputs by Mode

`GET /api/v1/tze-graph/outputs/by-mode`

Retrieves all outputs of a specific TZE mode with pagination.

**Query Parameters:**
- `mode` - TZE mode: `0` or `1` (required)
- `limit` ![optional](https://img.shields.io/badge/-optional-blue) - Number of outputs to return
- `offset` ![optional](https://img.shields.io/badge/-optional-blue) - Number of outputs to skip

**Examples:**
```
http://localhost:8080/api/v1/tze-graph/outputs/by-mode?mode=0&limit=10
http://localhost:8080/api/v1/tze-graph/outputs/by-mode?mode=1
```

#### Get TZE Outputs by Type and Mode

`GET /api/v1/tze-graph/outputs/by-type-mode`

Retrieves all outputs matching both type and mode with pagination.

**Query Parameters:**
- `type` - TZE type: `demo`, `stark_verify` (required)
- `mode` - TZE mode string (required, depends on type):
  - For `demo`: `open`, `close`
  - For `stark_verify`: `initialize`, `verify`
- `limit` ![optional](https://img.shields.io/badge/-optional-blue) - Number of outputs to return
- `offset` ![optional](https://img.shields.io/badge/-optional-blue) - Number of outputs to skip

**Examples:**
```
http://localhost:8080/api/v1/tze-graph/outputs/by-type-mode?type=demo&mode=open&limit=10
http://localhost:8080/api/v1/tze-graph/outputs/by-type-mode?type=stark_verify&mode=verify
```

#### Get Unspent TZE Outputs by Type

`GET /api/v1/tze-graph/outputs/unspent-by-type`

Retrieves all unspent outputs of a specific type with pagination.

**Query Parameters:**
- `type` - TZE type: `demo`, `stark_verify` (required)
- `limit` ![optional](https://img.shields.io/badge/-optional-blue) - Number of outputs to return
- `offset` ![optional](https://img.shields.io/badge/-optional-blue) - Number of outputs to skip

**Examples:**
```
http://localhost:8080/api/v1/tze-graph/outputs/unspent-by-type?type=demo&limit=10
http://localhost:8080/api/v1/tze-graph/outputs/unspent-by-type?type=stark_verify
```

#### Get Unspent TZE Outputs by Type and Mode

`GET /api/v1/tze-graph/outputs/unspent-by-type-mode`

Retrieves all unspent outputs matching type and mode.

**Query Parameters:**
- `type` - TZE type: `demo`, `stark_verify` (required)
- `mode` - TZE mode string (required, depends on type):
  - For `demo`: `open`, `close`
  - For `stark_verify`: `initialize`, `verify`
- `limit` ![optional](https://img.shields.io/badge/-optional-blue) - Number of outputs to return
- `offset` ![optional](https://img.shields.io/badge/-optional-blue) - Number of outputs to skip

**Examples:**
```
http://localhost:8080/api/v1/tze-graph/outputs/unspent-by-type-mode?type=demo&mode=open&limit=10
http://localhost:8080/api/v1/tze-graph/outputs/unspent-by-type-mode?type=stark_verify&mode=verify
```

#### Get Spent TZE Outputs

`GET /api/v1/tze-graph/outputs/spent`

Retrieves all spent TZE outputs with pagination.

**Query Parameters:**
- `limit` ![optional](https://img.shields.io/badge/-optional-blue) - Number of outputs to return
- `offset` ![optional](https://img.shields.io/badge/-optional-blue) - Number of outputs to skip

**Examples:**
```
http://localhost:8080/api/v1/tze-graph/outputs/spent?limit=10
http://localhost:8080/api/v1/tze-graph/outputs/spent?limit=50&offset=20
```

#### Get TZE Outputs by Value

`GET /api/v1/tze-graph/outputs/by-value`

Retrieves TZE outputs with value greater than or equal to minimum value.

**Query Parameters:**
- `min_value` ![optional](https://img.shields.io/badge/-optional-blue) - Minimum value (default: 0)
- `limit` ![optional](https://img.shields.io/badge/-optional-blue) - Number of outputs to return
- `offset` ![optional](https://img.shields.io/badge/-optional-blue) - Number of outputs to skip

**Examples:**
```
http://localhost:8080/api/v1/tze-graph/outputs/by-value?min_value=1000&limit=10
http://localhost:8080/api/v1/tze-graph/outputs/by-value?min_value=5000
```

---

## STARKS Module

> **Note:** This module must be enabled in configuration to use these endpoints.

### Verifiers

#### Get Verifier

`GET /api/v1/starks/verifiers/verifier`

Retrieves a single verifier by its ID.

**Query Parameters:**
- `verifier_id` - Verifier ID (required)

**Examples:**
```
http://localhost:8080/api/v1/starks/verifiers/verifier?verifier_id=verifier123
```

#### Get Verifier by Name

`GET /api/v1/starks/verifiers/by-name`

Retrieves a verifier by its name.

**Query Parameters:**
- `verifier_name` - Verifier name (required)

**Examples:**
```
http://localhost:8080/api/v1/starks/verifiers/by-name?verifier_name=StarkVerifier
```

#### Get All Verifiers

`GET /api/v1/starks/verifiers`

Retrieves all verifiers with pagination.

**Query Parameters:**
- `limit` ![optional](https://img.shields.io/badge/-optional-blue) - Number of verifiers to return
- `offset` ![optional](https://img.shields.io/badge/-optional-blue) - Number of verifiers to skip

**Examples:**
```
http://localhost:8080/api/v1/starks/verifiers?limit=10
http://localhost:8080/api/v1/starks/verifiers?limit=20&offset=10
```

#### Get Verifiers by Balance

`GET /api/v1/starks/verifiers/by-balance`

Retrieves verifiers sorted by balance with pagination.

**Query Parameters:**
- `limit` ![optional](https://img.shields.io/badge/-optional-blue) - Number of verifiers to return
- `offset` ![optional](https://img.shields.io/badge/-optional-blue) - Number of verifiers to skip

**Examples:**
```
http://localhost:8080/api/v1/starks/verifiers/by-balance?limit=10
http://localhost:8080/api/v1/starks/verifiers/by-balance
```

### STARK Proofs

#### Get STARK Proof

`GET /api/v1/starks/proofs/proof`

Retrieves a STARK proof by verifier ID and transaction ID.

**Query Parameters:**
- `verifier_id` - Verifier ID (required)
- `txid` - Transaction ID (required)

**Examples:**
```
http://localhost:8080/api/v1/starks/proofs/proof?verifier_id=verifier123&txid=abc123def456
```

#### Get STARK Proofs by Verifier

`GET /api/v1/starks/proofs/by-verifier`

Retrieves all STARK proofs for a verifier with pagination.

**Query Parameters:**
- `verifier_id` - Verifier ID (required)
- `limit` ![optional](https://img.shields.io/badge/-optional-blue) - Number of proofs to return
- `offset` ![optional](https://img.shields.io/badge/-optional-blue) - Number of proofs to skip

**Examples:**
```
http://localhost:8080/api/v1/starks/proofs/by-verifier?verifier_id=verifier123&limit=10
http://localhost:8080/api/v1/starks/proofs/by-verifier?verifier_id=verifier123
```

#### Get STARK Proofs by Transaction

`GET /api/v1/starks/proofs/by-transaction`

Retrieves all STARK proofs for a transaction.

**Query Parameters:**
- `txid` - Transaction ID (required)

**Examples:**
```
http://localhost:8080/api/v1/starks/proofs/by-transaction?txid=abc123def456
```

#### Get STARK Proofs by Block

`GET /api/v1/starks/proofs/by-block`

Retrieves all STARK proofs for a specific block.

**Query Parameters:**
- `block_height` - Block height (required)

**Examples:**
```
http://localhost:8080/api/v1/starks/proofs/by-block?block_height=100
http://localhost:8080/api/v1/starks/proofs/by-block?block_height=1000
```

#### Get Recent STARK Proofs

`GET /api/v1/starks/proofs/recent`

Retrieves the most recent STARK proofs with pagination.

**Query Parameters:**
- `limit` ![optional](https://img.shields.io/badge/-optional-blue) - Number of proofs to return
- `offset` ![optional](https://img.shields.io/badge/-optional-blue) - Number of proofs to skip

**Examples:**
```
http://localhost:8080/api/v1/starks/proofs/recent?limit=10
http://localhost:8080/api/v1/starks/proofs/recent?limit=20&offset=10
```

#### Get STARK Proofs by Size

`GET /api/v1/starks/proofs/by-size`

Retrieves STARK proofs filtered by size range with pagination.

**Query Parameters:**
- `min_size` ![optional](https://img.shields.io/badge/-optional-blue) - Minimum proof size in bytes (default: 0)
- `max_size` - Maximum proof size in bytes (required)
- `limit` ![optional](https://img.shields.io/badge/-optional-blue) - Number of proofs to return
- `offset` ![optional](https://img.shields.io/badge/-optional-blue) - Number of proofs to skip

**Examples:**
```
http://localhost:8080/api/v1/starks/proofs/by-size?min_size=1000&max_size=10000
http://localhost:8080/api/v1/starks/proofs/by-size?max_size=5000&limit=20
```

### Ztarknet Facts

> **Note:** Ztarknet indexing must be enabled for these endpoints.

#### Get Ztarknet Facts

`GET /api/v1/starks/facts/facts`

Retrieves Ztarknet facts by verifier ID and transaction ID.

**Query Parameters:**
- `verifier_id` - Verifier ID (required)
- `txid` - Transaction ID (required)

**Examples:**
```
http://localhost:8080/api/v1/starks/facts/facts?verifier_id=verifier123&txid=abc123def456
```

#### Get Ztarknet Facts by Verifier

`GET /api/v1/starks/facts/by-verifier`

Retrieves all Ztarknet facts for a verifier with pagination.

**Query Parameters:**
- `verifier_id` - Verifier ID (required)
- `limit` ![optional](https://img.shields.io/badge/-optional-blue) - Number of facts to return
- `offset` ![optional](https://img.shields.io/badge/-optional-blue) - Number of facts to skip

**Examples:**
```
http://localhost:8080/api/v1/starks/facts/by-verifier?verifier_id=verifier123&limit=10
http://localhost:8080/api/v1/starks/facts/by-verifier?verifier_id=verifier123
```

#### Get Ztarknet Facts by Transaction

`GET /api/v1/starks/facts/by-transaction`

Retrieves all Ztarknet facts for a transaction.

**Query Parameters:**
- `txid` - Transaction ID (required)

**Examples:**
```
http://localhost:8080/api/v1/starks/facts/by-transaction?txid=abc123def456
```

#### Get Ztarknet Facts by Block

`GET /api/v1/starks/facts/by-block`

Retrieves all Ztarknet facts for a specific block.

**Query Parameters:**
- `block_height` - Block height (required)

**Examples:**
```
http://localhost:8080/api/v1/starks/facts/by-block?block_height=100
http://localhost:8080/api/v1/starks/facts/by-block?block_height=1000
```

#### Get Ztarknet Facts by State

`GET /api/v1/starks/facts/by-state`

Retrieves Ztarknet facts by state hash.

**Query Parameters:**
- `state_hash` - State hash (required)

**Examples:**
```
http://localhost:8080/api/v1/starks/facts/by-state?state_hash=0x123abc
```

#### Get Ztarknet Facts by Program Hash

`GET /api/v1/starks/facts/by-program-hash`

Retrieves Ztarknet facts by program hash.

**Query Parameters:**
- `program_hash` - Program hash (required)

**Examples:**
```
http://localhost:8080/api/v1/starks/facts/by-program-hash?program_hash=0x456def
```

#### Get Ztarknet Facts by Inner Program Hash

`GET /api/v1/starks/facts/by-inner-program-hash`

Retrieves Ztarknet facts by inner program hash.

**Query Parameters:**
- `inner_program_hash` - Inner program hash (required)

**Examples:**
```
http://localhost:8080/api/v1/starks/facts/by-inner-program-hash?inner_program_hash=0x789ghi
```

#### Get Recent Ztarknet Facts

`GET /api/v1/starks/facts/recent`

Retrieves the most recent Ztarknet facts with pagination.

**Query Parameters:**
- `limit` ![optional](https://img.shields.io/badge/-optional-blue) - Number of facts to return
- `offset` ![optional](https://img.shields.io/badge/-optional-blue) - Number of facts to skip

**Examples:**
```
http://localhost:8080/api/v1/starks/facts/recent?limit=10
http://localhost:8080/api/v1/starks/facts/recent?limit=20&offset=10
```

#### Get State Transition

`GET /api/v1/starks/facts/state-transition`

Retrieves the state transition from old_state to new_state.

**Query Parameters:**
- `old_state` - Old state hash (required)
- `new_state` - New state hash (required)

**Examples:**
```
http://localhost:8080/api/v1/starks/facts/state-transition?old_state=0x123abc&new_state=0x456def
```

### Count

#### Count Verifiers

`GET /api/v1/starks/verifiers/count`

Returns the total count of verifiers.

**Response:**
```json
{
  "result": "success",
  "data": {
    "count": 42
  }
}
```

**Examples:**
```
http://localhost:8080/api/v1/starks/verifiers/count
```

#### Count STARK Proofs

`GET /api/v1/starks/proofs/count`

Returns the total count of STARK proofs with optional filters.

**Query Parameters:**
- `verifier_id` ![optional](https://img.shields.io/badge/-optional-blue) - Filter by verifier ID
- `block_height` ![optional](https://img.shields.io/badge/-optional-blue) - Filter by block height

**Response:**
```json
{
  "result": "success",
  "data": {
    "count": 2345
  }
}
```

**Examples:**
```
# Total count of all STARK proofs
http://localhost:8080/api/v1/starks/proofs/count

# Count by verifier
http://localhost:8080/api/v1/starks/proofs/count?verifier_id=verifier123

# Count by block height
http://localhost:8080/api/v1/starks/proofs/count?block_height=1000

# Count by verifier and block height
http://localhost:8080/api/v1/starks/proofs/count?verifier_id=verifier123&block_height=1500
```

#### Count Ztarknet Facts

`GET /api/v1/starks/facts/count`

Returns the total count of Ztarknet facts with optional filters.

**Query Parameters:**
- `verifier_id` ![optional](https://img.shields.io/badge/-optional-blue) - Filter by verifier ID
- `block_height` ![optional](https://img.shields.io/badge/-optional-blue) - Filter by block height

**Response:**
```json
{
  "result": "success",
  "data": {
    "count": 1234
  }
}
```

**Examples:**
```
# Total count of all Ztarknet facts
http://localhost:8080/api/v1/starks/facts/count

# Count by verifier
http://localhost:8080/api/v1/starks/facts/count?verifier_id=verifier123

# Count by block height
http://localhost:8080/api/v1/starks/facts/count?block_height=1000

# Count by verifier and block height
http://localhost:8080/api/v1/starks/facts/count?verifier_id=verifier123&block_height=1500
```

### Aggregations

#### Get Sum of Proof Sizes by Verifier

`GET /api/v1/starks/verifier/sum-proof-sizes`

Returns the aggregate sum of all STARK proof sizes for a given verifier.

**Query Parameters:**
- `verifier_id` - Verifier ID (required)

**Response:**
```json
{
  "result": "success",
  "data": {
    "total_proof_size": 1048576
  }
}
```

**Examples:**
```
http://localhost:8080/api/v1/starks/verifier/sum-proof-sizes?verifier_id=verifier123
```

---

## Base Routes

### Health Check

`GET /health`

Returns the health status of the API.

**Query Parameters:** None

**Examples:**
```
http://localhost:8080/health
```

**Response:**
```json
{
  "result": "healthy"
}
```
