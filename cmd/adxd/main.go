// Copyright (C) 2025, ADXYZ Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	"github.com/luxfi/adx/auction"
	"github.com/luxfi/adx/blocklace"
	"github.com/luxfi/adx/core"
	"github.com/luxfi/adx/da"
	"github.com/luxfi/adx/pkg/ids"
	"github.com/luxfi/adx/pkg/log"
	"github.com/luxfi/adx/settlement"
	"github.com/luxfi/adx/tee"
)

var (
	// Node configuration flags
	dataDir        = flag.String("data-dir", "/tmp/adxd", "Data directory")
	nodeID         = flag.String("node-id", "", "Node ID")
	port           = flag.Int("port", 8000, "HTTP port")
	rpcPort        = flag.Int("rpc-port", 9000, "RPC port")
	p2pPort        = flag.Int("p2p-port", 10000, "P2P port")
	networkID      = flag.String("network-id", "adx-local", "Network ID")
	logLevel       = flag.String("log-level", "info", "Log level")
	
	// Bootstrap configuration
	bootstrap      = flag.Bool("bootstrap", false, "Run as bootstrap node")
	bootstrapNodes = flag.String("bootstrap-nodes", "", "Bootstrap nodes (comma-separated)")
	
	// Node capabilities
	isMiner        = flag.Bool("miner", false, "Enable mining")
	teeMode        = flag.String("tee-mode", "simulated", "TEE mode: simulated, sgx, nitro")
	
	// Version info
	Version   = "dev"
	BuildTime = "unknown"
	GitCommit = "unknown"
)

// Node represents an ADX network node
type Node struct {
	mu sync.RWMutex
	
	// Identity
	ID        ids.NodeID
	NetworkID string
	
	// Core components
	DAG          *blocklace.DAG
	Miner        *blocklace.CordialMiner
	Enclave      *tee.Enclave
	BudgetMgr    *settlement.BudgetManager
	FreqMgr      *core.FrequencyManager
	DALayer      *da.DataAvailability
	
	// Networking
	httpServer *http.Server
	rpcServer  *http.Server
	peers      map[ids.NodeID]*Peer
	
	// State
	auctions   map[ids.ID]*auction.Auction
	isBootstrap bool
	isMiner    bool
	
	// Logging
	log log.Logger
}

// Peer represents a connected peer
type Peer struct {
	ID       ids.NodeID
	Endpoint string
	LastSeen time.Time
}

func main() {
	flag.Parse()
	
	// Print version info
	fmt.Printf("ADX Daemon (adxd) %s (commit: %s, built: %s)\n", Version, GitCommit, BuildTime)
	
	// Validate configuration
	if *nodeID == "" {
		fmt.Println("Error: --node-id is required")
		os.Exit(1)
	}
	
	// Initialize logger
	logger := initLogger(*logLevel)
	defer logger.Sync()
	
	// Create and start node
	node, err := NewNode(*nodeID, *networkID, logger)
	if err != nil {
		fmt.Printf("Failed to create node: %v\n", err)
		os.Exit(1)
	}
	
	// Start node services
	if err := node.Start(); err != nil {
		fmt.Printf("Failed to start node: %v\n", err)
		os.Exit(1)
	}
	
	// Wait for shutdown signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	
	fmt.Println("\nShutting down...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	
	if err := node.Shutdown(ctx); err != nil {
		fmt.Printf("Error during shutdown: %v\n", err)
	}
	
	fmt.Println("Daemon stopped")
}

// NewNode creates a new ADX node
func NewNode(nodeID, networkID string, logger log.Logger) (*Node, error) {
	// Parse node ID
	var nid ids.NodeID
	copy(nid[:], []byte(nodeID))
	
	// Initialize TEE (simplified for now)
	enclave, err := tee.NewEnclave(tee.EnclaveSimulated, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create enclave: %w", err)
	}
	
	// Initialize components
	dag := blocklace.NewDAG(logger)
	budgetMgr := settlement.NewBudgetManager(logger)
	freqMgr := core.NewFrequencyManager(logger)
	daLayer := da.NewDataAvailability(da.DALayerLocal, logger)
	
	node := &Node{
		ID:          nid,
		NetworkID:   networkID,
		DAG:         dag,
		Enclave:     enclave,
		BudgetMgr:   budgetMgr,
		FreqMgr:     freqMgr,
		DALayer:     daLayer,
		peers:       make(map[ids.NodeID]*Peer),
		auctions:    make(map[ids.ID]*auction.Auction),
		isBootstrap: *bootstrap,
		isMiner:     *isMiner,
		log:         logger,
	}
	
	// Initialize miner if enabled
	if *isMiner {
		node.Miner = blocklace.NewCordialMiner(nid, dag, logger)
	}
	
	return node, nil
}

// Start starts the node services
func (n *Node) Start() error {
	n.log.Info("Starting ADX node")
	
	// Start HTTP server
	httpRouter := n.setupHTTPRoutes()
	n.httpServer = &http.Server{
		Addr:    fmt.Sprintf(":%d", *port),
		Handler: httpRouter,
	}
	
	go func() {
		n.log.Info("HTTP server listening")
		if err := n.httpServer.ListenAndServe(); err != http.ErrServerClosed {
			n.log.Error("HTTP server error")
		}
	}()
	
	// Start RPC server
	rpcRouter := n.setupRPCRoutes()
	n.rpcServer = &http.Server{
		Addr:    fmt.Sprintf(":%d", *rpcPort),
		Handler: rpcRouter,
	}
	
	go func() {
		n.log.Info("RPC server listening")
		if err := n.rpcServer.ListenAndServe(); err != http.ErrServerClosed {
			n.log.Error("RPC server error")
		}
	}()
	
	// Connect to bootstrap nodes
	if !n.isBootstrap && *bootstrapNodes != "" {
		n.connectToBootstrapNodes(*bootstrapNodes)
	}
	
	// Start mining if enabled
	if n.isMiner {
		go n.runMiningLoop()
	}
	
	// Start metrics collection
	go n.collectMetrics()
	
	return nil
}

// Shutdown gracefully shuts down the node
func (n *Node) Shutdown(ctx context.Context) error {
	n.log.Info("Shutting down node")
	
	// Shutdown HTTP servers
	if err := n.httpServer.Shutdown(ctx); err != nil {
		n.log.Error("HTTP server shutdown error")
	}
	
	if err := n.rpcServer.Shutdown(ctx); err != nil {
		n.log.Error("RPC server shutdown error")
	}
	
	return nil
}

// setupHTTPRoutes sets up HTTP API routes
func (n *Node) setupHTTPRoutes() *mux.Router {
	r := mux.NewRouter()
	
	// Health check
	r.HandleFunc("/health", n.handleHealth).Methods("GET")
	
	// Node info
	r.HandleFunc("/info", n.handleInfo).Methods("GET")
	
	// Metrics
	r.HandleFunc("/metrics", n.handleMetrics).Methods("GET")
	
	return r
}

// setupRPCRoutes sets up RPC API routes
func (n *Node) setupRPCRoutes() *mux.Router {
	r := mux.NewRouter()
	
	// Status
	r.HandleFunc("/status", n.handleStatus).Methods("GET")
	
	// Auction endpoints
	r.HandleFunc("/auction/create", n.handleCreateAuction).Methods("POST")
	r.HandleFunc("/auction/bid", n.handleSubmitBid).Methods("POST")
	r.HandleFunc("/auction/{id}/status", n.handleAuctionStatus).Methods("GET")
	
	// Budget endpoints
	r.HandleFunc("/budget/fund", n.handleFundBudget).Methods("POST")
	r.HandleFunc("/budget/{advertiser}/status", n.handleBudgetStatus).Methods("GET")
	
	// Network endpoints
	r.HandleFunc("/network/peers", n.handleGetPeers).Methods("GET")
	r.HandleFunc("/network/connect", n.handleConnectPeer).Methods("POST")
	
	return r
}

// HTTP Handlers

func (n *Node) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, `{"status":"healthy","node":"%s"}`, n.ID)
}

func (n *Node) handleInfo(w http.ResponseWriter, r *http.Request) {
	n.mu.RLock()
	info := map[string]interface{}{
		"node_id":     string(n.ID[:]),
		"network_id":  n.NetworkID,
		"version":     Version,
		"is_bootstrap": n.isBootstrap,
		"is_miner":    n.isMiner,
		"tee_mode":    *teeMode,
		"num_peers":   len(n.peers),
		"num_auctions": len(n.auctions),
	}
	n.mu.RUnlock()
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, `%v`, info)
}

func (n *Node) handleMetrics(w http.ResponseWriter, r *http.Request) {
	metrics := n.DALayer.GetMetrics()
	
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "# HELP adx_da_stored_total Total blobs stored\n")
	fmt.Fprintf(w, "# TYPE adx_da_stored_total counter\n")
	fmt.Fprintf(w, "adx_da_stored_total %d\n", metrics.Stored)
	fmt.Fprintf(w, "\n")
	fmt.Fprintf(w, "# HELP adx_da_retrieved_total Total blobs retrieved\n")
	fmt.Fprintf(w, "# TYPE adx_da_retrieved_total counter\n")
	fmt.Fprintf(w, "adx_da_retrieved_total %d\n", metrics.Retrieved)
}

// RPC Handlers

func (n *Node) handleStatus(w http.ResponseWriter, r *http.Request) {
	n.mu.RLock()
	status := map[string]interface{}{
		"node_id":    string(n.ID[:]),
		"network_id": n.NetworkID,
		"peers":      len(n.peers),
		"auctions":   len(n.auctions),
		"dag_height": 0, // Simplified
		"timestamp":  time.Now().Unix(),
	}
	n.mu.RUnlock()
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, `%v`, status)
}

func (n *Node) handleCreateAuction(w http.ResponseWriter, r *http.Request) {
	// Parse request
	var req struct {
		SlotID   string `json:"slot_id"`
		Reserve  uint64 `json:"reserve"`
		Duration int64  `json:"duration_ms"`
	}
	
	// Simplified - in production, decode JSON properly
	req.SlotID = ids.GenerateTestID().String()
	req.Reserve = 1000
	req.Duration = 100
	
	// Create auction
	auctionID := ids.GenerateTestID()
	auc := auction.NewAuction(
		auctionID,
		req.Reserve,
		time.Duration(req.Duration)*time.Millisecond,
		n.log,
	)
	
	n.mu.Lock()
	n.auctions[auctionID] = auc
	n.mu.Unlock()
	
	// Return response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, `{"auction_id":"%s","status":"created"}`, auctionID)
}

func (n *Node) handleSubmitBid(w http.ResponseWriter, r *http.Request) {
	// Simplified bid submission
	bidderID := ids.GenerateTestID()
	
	sealedBid := &auction.SealedBid{
		BidderID:  bidderID,
		Timestamp: time.Now(),
	}
	
	// Submit to auction (simplified - would need to find actual auction)
	n.mu.Lock()
	if auc, exists := n.auctions[ids.GenerateTestID()]; exists {
		auc.SubmitBid(sealedBid)
	}
	n.mu.Unlock()
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, `{"status":"bid_submitted"}`)
}

func (n *Node) handleAuctionStatus(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, `{"status":"running","bids":0}`)
}

func (n *Node) handleFundBudget(w http.ResponseWriter, r *http.Request) {
	// Simplified budget funding
	amount := uint64(1000000)
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, `{"status":"funded","amount":%d}`, amount)
}

func (n *Node) handleBudgetStatus(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, `{"balance":1000000,"pending":0}`)
}

func (n *Node) handleGetPeers(w http.ResponseWriter, r *http.Request) {
	n.mu.RLock()
	peers := make([]string, 0, len(n.peers))
	for _, peer := range n.peers {
		peers = append(peers, peer.Endpoint)
	}
	n.mu.RUnlock()
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, `{"peers":%v}`, peers)
}

func (n *Node) handleConnectPeer(w http.ResponseWriter, r *http.Request) {
	// Parse peer endpoint (simplified)
	endpoint := r.FormValue("endpoint")
	if endpoint == "" {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, `{"error":"endpoint required"}`)
		return
	}
	
	// Add peer
	var peerID ids.NodeID
	copy(peerID[:], []byte(endpoint))
	
	n.mu.Lock()
	n.peers[peerID] = &Peer{
		ID:       peerID,
		Endpoint: endpoint,
		LastSeen: time.Now(),
	}
	n.mu.Unlock()
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, `{"status":"connected"}`)
}

// Mining loop
func (n *Node) runMiningLoop() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	
	for range ticker.C {
		// Simplified mining - just log
		n.log.Debug("Mining round completed")
	}
}

// Connect to bootstrap nodes
func (n *Node) connectToBootstrapNodes(nodes string) {
	for _, node := range strings.Split(nodes, ",") {
		node = strings.TrimSpace(node)
		if node == "" {
			continue
		}
		
		n.log.Info("Connecting to bootstrap node")
		
		// Extract node ID from address
		parts := strings.Split(node, "/")
		if len(parts) > 0 {
			nodeID := parts[len(parts)-1]
			var nid ids.NodeID
			copy(nid[:], []byte(nodeID))
			
			n.mu.Lock()
			n.peers[nid] = &Peer{
				ID:       nid,
				Endpoint: node,
				LastSeen: time.Now(),
			}
			n.mu.Unlock()
		}
	}
}

// Collect metrics periodically
func (n *Node) collectMetrics() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	
	for range ticker.C {
		n.mu.RLock()
		n.log.Debug("Node metrics")
		n.mu.RUnlock()
	}
}

// Initialize logger
func initLogger(level string) log.Logger {
	return log.NewWithLevel(level)
}