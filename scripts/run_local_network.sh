#!/bin/bash
# Copyright (C) 2025, ADXYZ Inc. All rights reserved.
# Run a local ADX network with multiple nodes

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Configuration
BASE_PORT=8000
BASE_RPC_PORT=9000
BASE_P2P_PORT=10000
NUM_NODES=${1:-5}  # Default to 5 nodes
NETWORK_ID="adx-local"
LOG_LEVEL=${LOG_LEVEL:-"info"}

echo -e "${GREEN}=== ADX Local Network Setup ===${NC}"
echo "Starting $NUM_NODES nodes..."

# Clean up function
cleanup() {
    echo -e "${YELLOW}Cleaning up...${NC}"
    pkill -f "adxd" || true
    rm -rf /tmp/adx-local-*
}

# Set up trap for cleanup
trap cleanup EXIT

# Create data directories
for i in $(seq 1 $NUM_NODES); do
    DATA_DIR="/tmp/adx-local-node-$i"
    rm -rf $DATA_DIR
    mkdir -p $DATA_DIR/logs
done

# Build the daemon binary
echo -e "${GREEN}Building ADX daemon (adxd)...${NC}"
go build -o bin/adxd ./cmd/adxd

# Generate bootstrap node key
BOOTSTRAP_KEY=$(openssl rand -hex 32)
BOOTSTRAP_ID="node-1"
BOOTSTRAP_ADDR="/ip4/127.0.0.1/tcp/$((BASE_P2P_PORT))"

# Start bootstrap node (Node 1)
echo -e "${GREEN}Starting bootstrap node...${NC}"
DATA_DIR="/tmp/adx-local-node-1"
./bin/adxd \
    --data-dir=$DATA_DIR \
    --node-id=$BOOTSTRAP_ID \
    --port=$((BASE_PORT)) \
    --rpc-port=$((BASE_RPC_PORT)) \
    --p2p-port=$((BASE_P2P_PORT)) \
    --network-id=$NETWORK_ID \
    --log-level=$LOG_LEVEL \
    --bootstrap \
    --miner \
    --tee-mode=simulated \
    > $DATA_DIR/logs/node.log 2>&1 &

NODE1_PID=$!
echo "Bootstrap node PID: $NODE1_PID"

# Wait for bootstrap node to start
sleep 3

# Start remaining nodes
for i in $(seq 2 $NUM_NODES); do
    echo -e "${GREEN}Starting node $i...${NC}"
    
    DATA_DIR="/tmp/adx-local-node-$i"
    NODE_ID="node-$i"
    PORT=$((BASE_PORT + i - 1))
    RPC_PORT=$((BASE_RPC_PORT + i - 1))
    P2P_PORT=$((BASE_P2P_PORT + i - 1))
    
    # Every 3rd node is a miner
    MINER_FLAG=""
    if [ $((i % 3)) -eq 0 ]; then
        MINER_FLAG="--miner"
        echo "  Node $i is a miner"
    fi
    
    # Start node
    ./bin/adxd \
        --data-dir=$DATA_DIR \
        --node-id=$NODE_ID \
        --port=$PORT \
        --rpc-port=$RPC_PORT \
        --p2p-port=$P2P_PORT \
        --network-id=$NETWORK_ID \
        --log-level=$LOG_LEVEL \
        --bootstrap-nodes="${BOOTSTRAP_ADDR}/p2p/${BOOTSTRAP_ID}" \
        --tee-mode=simulated \
        $MINER_FLAG \
        > $DATA_DIR/logs/node.log 2>&1 &
    
    echo "  Node $i PID: $!"
    sleep 1
done

echo -e "${GREEN}=== Network Started ===${NC}"
echo ""
echo "Nodes are running with:"
echo "  HTTP ports: $BASE_PORT-$((BASE_PORT + NUM_NODES - 1))"
echo "  RPC ports:  $BASE_RPC_PORT-$((BASE_RPC_PORT + NUM_NODES - 1))"
echo "  P2P ports:  $BASE_P2P_PORT-$((BASE_P2P_PORT + NUM_NODES - 1))"
echo ""
echo "Logs are available at: /tmp/adx-local-node-*/logs/"
echo ""
echo "To check node status:"
echo "  curl http://localhost:$BASE_RPC_PORT/status"
echo ""
echo "To submit a bid:"
echo "  curl -X POST http://localhost:$BASE_RPC_PORT/auction/bid -d '{...}'"
echo ""
echo "Press Ctrl+C to stop the network"

# Wait for interrupt
wait