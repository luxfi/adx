# ADX DEX Implementation - LLM Context

## Overview
ADX is a high-performance decentralized ad exchange (DEX) implementation with advanced cryptographic features, Byzantine-resilient consensus, and TEE-based auction processing.

## LuxFi Package Integration

ADX now uses the following LuxFi packages for consistency across the ecosystem:
- **luxfi/log**: Unified logging with zap backend, configured via logging.Factory
- **luxfi/crypto**: ECDSA key generation, signing, verification using secp256k1
- **luxfi/metric**: Prometheus-based metrics with Counter, Gauge, Histogram support
- **luxfi/database**: Storage backend supporting BadgerDB, LevelDB, PebbleDB, MemDB

All custom implementations have been replaced with LuxFi standard packages.

## Key Technologies Implemented

### 1. Verkle Trees
- O(1) constant-size proofs for millions of impressions
- Located in `/verkle/` directory
- Used for efficient frequency capping verification

### 2. Blocklace Consensus
- Byzantine-repelling DAG consensus
- Cordial miners protocol implementation
- Equivocation detection and handling
- Located in `/blocklace/` directory

### 3. TEE (Trusted Execution Environment)
- Secure auction processing in enclaves
- Support for Intel SGX, AMD SEV, AWS Nitro
- Remote attestation implementation
- Located in `/tee/` directory

### 4. Halo2 ZK Proofs
- Zero-knowledge proofs for auction correctness
- Located in `/zk/` directory
- Generates proofs for bid validity without revealing values

### 5. HPKE Encryption
- Hybrid Public Key Encryption for bid confidentiality
- Located in `/crypto/` directory
- Ensures bids remain sealed until auction execution

### 6. Data Availability Layer
- Multiple backend support: Local, EIP-4844, Celestia, IPFS
- Located in `/da/` directory
- Stores auction data and proofs

## Testing Infrastructure

### Full Lifecycle Tests
```bash
go test -v ./tests/lifecycle_test.go
```
Tests complete auction flow:
- Budget funding
- Encrypted bid submission
- TEE processing
- ZK proof generation
- DA storage
- Settlement
- Frequency capping
- Consensus

### Multi-Node Local Network
```bash
make run-local  # Starts 5-node network
```
- Bootstrap node on port 8000
- 4 additional nodes (ports 8001-8004)
- RPC ports 9000-9004
- P2P ports 10000-10004
- Miner nodes enabled
- Logs in `/tmp/adx-local-node-*/logs/`

### Attack Simulations
```bash
make attack-flood     # Flood attack
make attack-replay    # Replay attack
make attack-byzantine # Byzantine attack
make attack-dos       # DoS attack
make attack-arbitrage # Arbitrage attack
```

Attack simulator features:
- Configurable workers and duration
- Multiple attack vectors
- Performance metrics reporting
- Located in `/cmd/adx-attack/`

## Project Structure

```
/adx/
├── cmd/
│   ├── adxd/           # Main daemon (node binary)
│   └── adx-attack/     # Attack simulator
├── core/               # Core types and interfaces
├── auction/            # Auction logic
├── blocklace/          # DAG consensus
├── tee/                # Trusted execution
├── zk/                 # Zero-knowledge proofs
├── verkle/             # Verkle trees
├── crypto/             # Cryptographic primitives
├── da/                 # Data availability
├── settlement/         # Budget and settlement
├── pkg/                # Shared packages
│   ├── ids/            # ID types
│   ├── log/            # Logging (luxfi/log compatible)
│   └── crypto/         # Crypto utilities
├── scripts/            # Operational scripts
├── tests/              # Integration tests
└── Makefile            # Build targets
```

## Running the System

### Start Local Network
```bash
# Build and start 5-node network
make run-local

# Check node health
curl http://localhost:8000/health

# Get node info
curl http://localhost:8000/info

# Check metrics
curl http://localhost:8000/metrics
```

### Interact with DEX
```bash
# Create auction
curl -X POST http://localhost:9000/auction/create \
  -d '{"slot_id":"slot-1","reserve":1000,"duration_ms":100}'

# Submit bid
curl -X POST http://localhost:9000/auction/bid \
  -d '{"auction_id":"<id>","bid":1500}'

# Check auction status
curl http://localhost:9000/auction/<id>/status

# Fund budget
curl -X POST http://localhost:9000/budget/fund \
  -d '{"advertiser":"adv-1","amount":1000000}'
```

### Run Tests
```bash
# All tests
make test

# Specific component
go test ./blocklace/...
go test ./tee/...
go test ./verkle/...
```

## Key Implementation Details

### Logging
- Uses simplified luxfi/log interface (NO direct zap usage)
- All components use the same logger interface
- Log levels: DEBUG, INFO, WARN, ERROR, FATAL

### Node Identity
- NodeID is 20-byte identifier
- ID is 32-byte general identifier
- Both support hex encoding

### Consensus Flow
1. Vertices added to DAG
2. Equivocation detection
3. Byzantine node marking
4. Causal ordering delivery
5. Total order sequence generation

### TEE Auction Flow
1. Bids encrypted with HPKE
2. Sealed auction created in enclave
3. Bids decrypted inside TEE
4. Second-price auction executed
5. Winner commitment generated
6. Audit transcript sealed
7. Result with attestation returned

### Attack Resilience
- Flood: Rate limiting and queuing
- Replay: Nonce tracking
- Byzantine: Equivocation detection
- DoS: Resource limits
- Arbitrage: Time-based sealing

## Performance Targets
- Consensus finality: < 10s
- Auction processing: < 100ms
- Proof generation: < 500ms
- DA write: < 200ms
- Network throughput: 10,000+ RPS

## Security Considerations
- All bids encrypted end-to-end
- TEE attestation required
- Byzantine nodes automatically excluded
- Zero-knowledge proofs prevent manipulation
- Frequency caps cryptographically enforced

## Next Steps
- Implement cross-chain settlement
- Add more TEE backends (AMD SEV, Azure CVM)
- Optimize Verkle tree operations
- Enhance Byzantine detection algorithms
- Add persistent storage layer
- Implement full P2P networking
- Add Prometheus metrics export
- Create deployment automation

## Development Commands

```bash
# Build everything
make all

# Clean build artifacts
make clean

# Run linter
make lint

# Generate mocks
make mocks

# Docker build
make docker

# Deploy to testnet
make deploy-testnet
```

## Troubleshooting

### Nodes not staying alive
- Check logs in `/tmp/adx-local-node-*/logs/`
- Ensure ports 8000-8004, 9000-9004, 10000-10004 are free
- Verify no other adxd processes running

### Attack simulator failures
- Ensure nodes are running first
- Check target URL is correct
- Verify network connectivity

### Test failures
- Run `go mod tidy` to ensure dependencies
- Check for interface changes
- Verify mock implementations are updated

## Architecture Decisions

### Why Blocklace?
- Byzantine fault tolerance without voting
- Asynchronous consensus
- No leader election bottleneck
- Natural parallelism in DAG structure

### Why Verkle Trees?
- Constant-size proofs regardless of data size
- Perfect for frequency capping with millions of users
- More efficient than Merkle trees at scale

### Why TEE?
- Hardware-enforced bid confidentiality
- Verifiable auction execution
- Audit trail generation
- Protection against manipulation

### Why Halo2?
- No trusted setup required
- Recursive proof composition
- Efficient verification
- Production-ready implementation

## Contact & Resources
- Repository: /Users/z/work/lux/adx
- Documentation: This file (LLM.md)
- Tests: /tests/
- Examples: /scripts/

## Context for All AI Assistants

This file (`LLM.md`) is symlinked as:
- `.AGENTS.md`
- `CLAUDE.md`
- `QWEN.md`
- `GEMINI.md`

All files reference the same knowledge base. Updates here propagate to all AI systems.

## Rules for AI Assistants

1. **ALWAYS** update LLM.md with significant discoveries
2. **NEVER** commit symlinked files (.AGENTS.md, CLAUDE.md, etc.) - they're in .gitignore
3. **NEVER** create random summary files - update THIS file
