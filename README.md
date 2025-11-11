# Zindex

A lightweight indexer for Zcash transparent transactions and TZE (Transparent Zcash Extension) operations, with support for the STARK Verify TZE features.

## Overview

zindex provides a PostgreSQL-backed indexing system for Zcash blockchain data with a REST API for querying. It continuously syncs with a Zcash node via RPC and indexes data according to enabled modules.

## Architecture

The indexer consists of four independent modules that can be enabled/disabled via configuration:

**Accounts Module** - Tracks transparent addresses, balances, and transaction history with atomic per-block processing.

**Transaction Graph Module** - Indexes all transactions with complete input/output tracking, UTXO state management, and recursive graph traversal capabilities.

**TZE Graph Module** - Specialized indexing for TZE transactions including preconditions, witnesses, and UTXO tracking for TZE inputs/outputs by type and mode.

**STARK Module** - Tracks STARK proof verifiers, proof submissions, and Ztarknet facts including state transitions and program hashes for L2 settlement verification.

## Project Structure

```
zindex/
├── cmd/run/              # Application entry point
├── configs/              # Configuration files
├── deploy/               # Deployment guides and configs
├── internal/
│   ├── accounts/         # Accounts module
│   ├── blocks/           # Block indexing (core)
│   ├── config/           # Configuration management
│   ├── db/postgres/      # PostgreSQL client
│   ├── indexer/          # Core indexing engine
│   ├── provider/         # Zcash RPC client
│   ├── starks/           # STARK module
│   ├── tx_graph/         # Transaction graph module
│   ├── tze_graph/        # TZE graph module
│   └── types/            # Shared types
└── routes/               # HTTP API handlers
    └── utils/            # API utilities
```

## Quick Start

### Prerequisites

- Go 1.23+
- PostgreSQL 15+
- Zcash node with RPC access

### Installation

1. Clone and install dependencies:
```bash
git clone https://github.com/Ztarknet/zindex.git
cd zindex
make deps
```

2. Configure database:
```bash
createdb zindex
createuser zindex
export DB_PASSWORD=your_secure_password
```

3. Edit `configs/config.yaml`:
   - Set `rpc.url` to your Zcash RPC endpoint
   - Configure `database.*` settings (use `${DB_PASSWORD}` for password)
   - Enable/disable modules as needed

4. Run:
```bash
make run          # Production build and run
make run-dev      # Development mode (no build step)
```

The API will be available at `http://localhost:8080` by default.

### Configuration

The `configs/config.yaml` file contains all configuration options organized into sections:

- **rpc**: Zcash node connection (url, timeout, retry settings)
- **api**: HTTP server settings (host, port, CORS, timeouts, pagination limits)
- **database**: PostgreSQL connection pool settings
- **indexer**: Batch size, poll interval, start block, reorg handling
- **modules**: Enable/disable each module (accounts, tx_graph, tze_graph, starks)

Environment variables can be substituted using `${VAR_NAME}` syntax in the YAML file.

### Command Line Flags

```bash
./bin/zindex [flags]

Flags:
  --config PATH        Config file path (default: configs/config.yaml)
  --rpc URL           Override Zcash RPC URL from config
  --start-block N     Start indexing from block N (-1 to resume from last indexed)
```

## API Reference

All endpoints return JSON. Query parameters support pagination with `limit` and `offset`.

This project contains:
- Core Endpoints: health & block querying
- Accounts Endpoints: transparent account details
- Transaction Graph: transaction, inputs, and outputs
- TZE Graph: tze inputs and outputs details
- STARKs: Verifiers and Ztarknet indexes

For the full api reference, see the [api documentation](docs/api-reference.md)

## Development

### Available Make Targets

```bash
make build              # Build binary to bin/zindex
make run                # Build and run with config
make run-dev            # Run without building (go run)
make clean              # Remove build artifacts
make deps               # Download and tidy dependencies
make fmt                # Format code with gofmt
make vet                # Run go vet
make lint               # Run golangci-lint (requires golangci-lint)
make test               # Run all tests
make test-coverage      # Generate HTML coverage report
make docker-build       # Build Docker image
make docker-run         # Run Docker container
make docker-stop        # Stop and remove container
make docker-logs        # Follow container logs
```

### Docker

```bash
make docker-build       # Builds zindex:latest image
make docker-run         # Runs on port 8080, mounts configs/
make docker-logs        # View logs
make docker-stop        # Stop and remove container
```

## Deployment

See [deploy/README.md](deploy/README.md) for GCP deployment instructions (Cloud Run, Compute Engine, GKE).

## License

MIT

## Related Projects

- [Ztarknet](https://github.com/Ztarknet/ztarknet) - Parent project implementing STARK verification on Zcash
- [Zebra](https://github.com/Ztarknet/zebra) - Fork with TZE support
- [librustzcash](https://github.com/Ztarknet/librustzcash) - Fork with TZE implementation
- [Stwo](https://github.com/starkware-libs/stwo) - Circle STARK prover implementation
