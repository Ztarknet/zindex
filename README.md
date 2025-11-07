# Zindex

A lightweight indexer for Zcash transparent transactions and TZE (Transparent Zcash Extension) operations, with support for the STARK Verify TZE features.

## Overview

Zindex tracks the transparent transaction graph on Zcash, providing APIs to query spending relationships, UTXO states, and TZE data. It's designed specifically to support Ztarknet's L2 settlement pattern where Circle-STARK proofs are verified via TZE on Zcash L1.

## Features

- **Transaction Graph**: Track inputs, outputs, and spending relationships
- **UTXO Tracking**: Query unspent outputs
- **ACCOUNTS**: Track users/accounts utxos, history, and balances
- **TZE Indexing**: Index TZE inputs/outputs with precondition and witness data
- **STARK Anchors**: Specialized indexing for STARK verification TZE type
- **Anchor Chain**: Track L2 state progression through linked anchor UTXOs
- **REST API**: Query blockchain data with simple HTTP endpoints

## Development Status

**Current**: Proof of concept for Ztarknet L2 settlement

**Focus**: Demonstrating feasibility of STARK verification via TZE on Zcash L1

## Project Structure

```
zindex/
├── cmd/
│   └── run/           # Main entry point
├── configs/           # YAML configuration files
├── deploy/            # GCP deployment instructions
├── internal/          # Internal packages
│   ├── config/        # Configuration management
│   ├── db/postgres/   # PostgreSQL database layer
│   ├── provider/      # Zcash RPC provider
│   ├── accounts/      # Account tracking module
│   ├── tx_graph/      # Transaction graph module
│   ├── tze_graph/     # TZE transaction module
│   └── starks/        # STARK proof module
└── routes/            # API endpoints
    └── utils/         # API utilities (middleware, responses, requests)
```

## Quick Start

### Prerequisites

- Go 1.21.7 or higher
- PostgreSQL 15+
- Zcash node with RPC enabled

### Installation

1. Clone the repository:
```bash
git clone https://github.com/keep-starknet-strange/ztarknet/zindex.git
cd zindex
```

2. Install dependencies:
```bash
make deps
```

3. Configure the application:
```bash
cp configs/config.yaml configs/config.local.yaml
# Edit config.local.yaml with your settings
```

4. Set up PostgreSQL:
```bash
createdb zindex
createuser zindex
```

5. Build and run:
```bash
make build
make run
```

Or run in development mode:
```bash
make run-dev
```

### Configuration

Edit `configs/config.yaml` to configure:

- **RPC**: Zcash node connection details
- **API**: Server host, port, and CORS settings
- **Database**: PostgreSQL connection settings
- **Indexer**: Batch size, polling interval, reorg handling
- **Modules**: Enable/disable specific features (TX_GRAPH, TZE_GRAPH, STARKS, ACCOUNTS)

### Command Line Options

```bash
./bin/zindex --help

Options:
  --config string       Path to config file (default "configs/config.yaml")
  --rpc string          Zcash RPC URL (overrides config)
  --start-block int     Starting block height (default: resume from last indexed)
```

### API Endpoints

#### Health Check
```bash
GET /health
```

#### Accounts
```bash
GET /api/v1/account?address=<address>
GET /api/v1/accounts?limit=50&offset=0
```

#### Transaction Graph
```bash
GET /api/v1/transaction?txid=<txid>
GET /api/v1/transaction/graph?txid=<txid>&depth=3
```

#### TZE Transactions
```bash
GET /api/v1/tze/transaction?txid=<txid>
GET /api/v1/tze/transactions?type=<type>&limit=50&offset=0
GET /api/v1/tze/witnesses?txid=<txid>
```

#### STARK Proofs
```bash
GET /api/v1/proof?id=<id>
GET /api/v1/proof/transaction?txid=<txid>
GET /api/v1/proof/stats
GET /api/v1/proof/unverified?limit=50
```

### Docker

Build and run with Docker:

```bash
make docker-build
make docker-run
```

Stop the container:
```bash
make docker-stop
```

View logs:
```bash
make docker-logs
```

### Development

Format code:
```bash
make fmt
```

Run tests:
```bash
make test
```

Run tests with coverage:
```bash
make test-coverage
```

Lint code:
```bash
make lint
```

## Deployment

See the [deploy/README.md](deploy/README.md) for detailed instructions on deploying to Google Cloud Platform.

## Related Projects

- [Ztarknet](https://github.com/Ztarknet/ztarknet) - Starknet-style L2 rollup for Zcash
- [Zebra](https://github.com/Ztarknet/zebra) - Fork with TZE support
- [librustzcash](https://github.com/Ztarknet/librustzcash) - Fork with TZE implementation
- [Stwo](https://github.com/starkware-libs/stwo-cairo) - Circle STARK prover
