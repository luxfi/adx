# LuxFi ADX Makefile
# High-Performance CTV Ad Exchange

SHELL := /bin/bash
.PHONY: all build test clean help

# Version and build info
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME := $(shell date -u '+%Y-%m-%d_%H:%M:%S')
GIT_COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")

# Go parameters
GO := go
GOBUILD := $(GO) build
GOCLEAN := $(GO) clean
GOTEST := $(GO) test
GOGET := $(GO) get
GOMOD := $(GO) mod
GOVET := $(GO) vet
GOFMT := gofmt

# Build parameters
CGO_ENABLED ?= 0
LDFLAGS := -ldflags "-X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME) -X main.GitCommit=$(GIT_COMMIT)"

# Binary output
BINARY_NAME := adx-exchange
MINER_BINARY := adx-miner
DAEMON_BINARY := adxd
ATTACK_BINARY := adx-attack
BINARY_DIR := bin

# Test parameters
TEST_TIMEOUT := 30s
BENCH_TIME := 10s

# Default target - build, test, then benchmark
all: build test bench
	@echo "✅ ADX build, test, and benchmark complete!"

help:
	@echo "LuxFi ADX Makefile Commands:"
	@echo ""
	@echo "Development:"
	@echo "  make build         - Build all binaries"
	@echo "  make test          - Run all tests"
	@echo "  make bench         - Run benchmarks"
	@echo "  make clean         - Clean build artifacts"
	@echo ""
	@echo "Running:"
	@echo "  make run-exchange  - Run ADX exchange"
	@echo "  make run-miner     - Run home miner"
	@echo "  make run-local     - Run 5-node local network"
	@echo "  make run-local-large - Run 10-node local network"
	@echo ""
	@echo "Attack Testing:"
	@echo "  make attack-flood  - Run flood attack simulation"
	@echo "  make attack-replay - Run replay attack"
	@echo "  make attack-byzantine - Run Byzantine attack"
	@echo "  make attack-dos    - Run DoS attack"
	@echo "  make attack-all    - Run all attack scenarios"
	@echo ""
	@echo "Integration Tests:"
	@echo "  make test-lifecycle - Run full lifecycle test"
	@echo "  make test-concurrent - Test concurrent auctions"
	@echo "  make test-byzantine - Test Byzantine resilience"
	@echo ""
	@echo "Docker:"
	@echo "  make docker-build  - Build Docker images"
	@echo "  make docker-run    - Run with Docker Compose"
	@echo "  make docker-clean  - Clean Docker resources"

# Build targets
build: build-exchange build-miner build-daemon build-attack
	@echo "✅ All binaries built successfully"

build-exchange:
	@echo "🔨 Building ADX exchange..."
	@mkdir -p $(BINARY_DIR)
	@CGO_ENABLED=$(CGO_ENABLED) $(GOBUILD) $(LDFLAGS) -o $(BINARY_DIR)/$(BINARY_NAME) ./cmd/adx-exchange

build-miner:
	@echo "🔨 Building ADX miner..."
	@mkdir -p $(BINARY_DIR)
	@CGO_ENABLED=$(CGO_ENABLED) $(GOBUILD) $(LDFLAGS) -o $(BINARY_DIR)/$(MINER_BINARY) ./cmd/adx-miner

build-daemon:
	@echo "🔨 Building ADX daemon (adxd)..."
	@mkdir -p $(BINARY_DIR)
	@CGO_ENABLED=$(CGO_ENABLED) $(GOBUILD) $(LDFLAGS) -o $(BINARY_DIR)/$(DAEMON_BINARY) ./cmd/adxd

build-attack:
	@echo "⚔️  Building attack simulator..."
	@mkdir -p $(BINARY_DIR)
	@CGO_ENABLED=$(CGO_ENABLED) $(GOBUILD) $(LDFLAGS) -o $(BINARY_DIR)/$(ATTACK_BINARY) ./cmd/adx-attack

# Test targets
test:
	@echo "🧪 Running tests..."
	@$(GOTEST) -v -timeout $(TEST_TIMEOUT) ./pkg/vast/... ./pkg/rtb/... ./pkg/miner/...

test-coverage:
	@echo "📊 Running tests with coverage..."
	@$(GOTEST) -v -coverprofile=coverage.out -timeout $(TEST_TIMEOUT) ./...
	@$(GO) tool cover -html=coverage.out -o coverage.html

bench:
	@echo "⚡ Running benchmarks..."
	@$(GOTEST) -bench=. -benchtime=$(BENCH_TIME) ./pkg/rtb/...

# Code quality
fmt:
	@echo "📝 Formatting code..."
	@$(GOFMT) -w .

vet:
	@echo "🔍 Running go vet..."
	@$(GOVET) ./...

lint:
	@echo "🔍 Running linter..."
	@golangci-lint run

# Running targets
run-exchange:
	@echo "🚀 Starting ADX exchange..."
	@$(BINARY_DIR)/$(BINARY_NAME)

run-miner:
	@echo "⛏️ Starting ADX miner..."
	@$(BINARY_DIR)/$(MINER_BINARY) --tunnel localxpose --cache-size 10GB

# Local network targets
run-local: build-daemon
	@echo "🌐 Starting 5-node local network..."
	@chmod +x scripts/run_local_network.sh
	@./scripts/run_local_network.sh 5

run-local-large: build-daemon
	@echo "🌐 Starting 10-node local network..."
	@chmod +x scripts/run_local_network.sh
	@./scripts/run_local_network.sh 10

stop-local:
	@echo "🛑 Stopping local network..."
	@pkill -f "adxd" || true
	@rm -rf /tmp/adx-local-*

# Lifecycle test targets
test-lifecycle:
	@echo "🔄 Running full lifecycle test..."
	@$(GOTEST) -v -timeout 60s ./tests -run TestFullLifecycle

test-concurrent:
	@echo "⚡ Testing concurrent auctions..."
	@$(GOTEST) -v -timeout 60s ./tests -run TestConcurrentAuctions

test-byzantine:
	@echo "👹 Testing Byzantine resilience..."
	@$(GOTEST) -v -timeout 60s ./tests -run TestByzantineResilience

# Attack simulation targets
attack-flood: build-attack
	@echo "🌊 Running flood attack..."
	@$(BINARY_DIR)/$(ATTACK_BINARY) -type flood -duration 30s -workers 50 -rps 1000

attack-replay: build-attack
	@echo "🔁 Running replay attack..."
	@$(BINARY_DIR)/$(ATTACK_BINARY) -type replay -duration 30s -workers 20

attack-byzantine: build-attack
	@echo "👹 Running Byzantine attack..."
	@$(BINARY_DIR)/$(ATTACK_BINARY) -type byzantine -duration 30s -workers 10

attack-dos: build-attack
	@echo "💣 Running DoS attack..."
	@$(BINARY_DIR)/$(ATTACK_BINARY) -type dos -duration 30s -workers 100

attack-arbitrage: build-attack
	@echo "💰 Running arbitrage attack..."
	@$(BINARY_DIR)/$(ATTACK_BINARY) -type arbitrage -duration 30s -workers 5

attack-all: build-attack
	@echo "⚔️  Running all attack scenarios..."
	@$(BINARY_DIR)/$(ATTACK_BINARY) -type flood -duration 10s
	@sleep 2
	@$(BINARY_DIR)/$(ATTACK_BINARY) -type replay -duration 10s  
	@sleep 2
	@$(BINARY_DIR)/$(ATTACK_BINARY) -type byzantine -duration 10s
	@sleep 2
	@$(BINARY_DIR)/$(ATTACK_BINARY) -type dos -duration 10s
	@sleep 2
	@$(BINARY_DIR)/$(ATTACK_BINARY) -type arbitrage -duration 10s

# Docker targets
docker-build:
	@echo "🐳 Building Docker images..."
	@docker build -t luxfi/adx-exchange:$(VERSION) -f docker/exchange/Dockerfile .
	@docker build -t luxfi/adx-miner:$(VERSION) -f docker/miner/Dockerfile .

docker-run:
	@echo "🐳 Starting ADX with Docker Compose..."
	@docker-compose up -d

docker-stop:
	@echo "🛑 Stopping Docker services..."
	@docker-compose down

docker-clean:
	@echo "🧹 Cleaning Docker resources..."
	@docker-compose down -v
	@docker rmi luxfi/adx-exchange:$(VERSION) luxfi/adx-miner:$(VERSION) 2>/dev/null || true

# FoundationDB setup
setup-fdb:
	@echo "📀 Setting up FoundationDB..."
	@wget https://github.com/apple/foundationdb/releases/download/7.3.27/foundationdb-clients_7.3.27-1_amd64.deb
	@sudo dpkg -i foundationdb-clients_7.3.27-1_amd64.deb
	@rm foundationdb-clients_7.3.27-1_amd64.deb

# Clean target
clean:
	@echo "🧹 Cleaning build artifacts..."
	@$(GOCLEAN)
	@rm -rf $(BINARY_DIR)
	@rm -f coverage.out coverage.html

# Installation
install:
	@echo "📦 Installing ADX binaries..."
	@$(GO) install -v ./cmd/adx-exchange
	@$(GO) install -v ./cmd/adx-miner

# Dependencies
deps:
	@echo "📦 Downloading dependencies..."
	@$(GOMOD) download
	@$(GOMOD) tidy

# CI targets
ci: deps fmt vet test build
	@echo "✅ CI pipeline complete"

# Development setup
dev-setup:
	@echo "🛠️ Setting up development environment..."
	@$(GOGET) -u github.com/golangci/golangci-lint/cmd/golangci-lint
	@$(GOGET) -u github.com/apple/foundationdb/bindings/go/src/fdb
	@$(GOGET) -u github.com/prebid/openrtb/v20
	@echo "✅ Development environment ready"

.DEFAULT_GOAL := help