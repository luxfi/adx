// Copyright (C) 2025, ADXYZ Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package main

import (
	"bytes"
	"crypto/rand"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/luxfi/adx/pkg/ids"
)

var (
	// Attack configuration flags
	targetURL  = flag.String("target", "http://localhost:9000", "Target RPC endpoint")
	attackType = flag.String("type", "flood", "Attack type: flood, replay, byzantine, dos, arbitrage")
	duration   = flag.Duration("duration", 60*time.Second, "Attack duration")
	workers    = flag.Int("workers", 100, "Number of concurrent workers")
	rps        = flag.Int("rps", 1000, "Requests per second")

	// Statistics
	totalRequests int64
	successCount  int64
	errorCount    int64
	latencySum    int64
	maxLatency    int64
)

// BidRequest represents an auction bid
type BidRequest struct {
	AuctionID    string `json:"auction_id"`
	BidderID     string `json:"bidder_id"`
	Amount       uint64 `json:"amount"`
	EncryptedBid string `json:"encrypted_bid"`
	Signature    string `json:"signature"`
	Nonce        uint64 `json:"nonce"`
}

// AuctionRequest represents auction creation
type AuctionRequest struct {
	SlotID     string   `json:"slot_id"`
	Reserve    uint64   `json:"reserve"`
	Duration   int64    `json:"duration_ms"`
	Publishers []string `json:"publishers"`
}

func main() {
	flag.Parse()

	fmt.Printf("=== ADX DEX Attack Simulator ===\n")
	fmt.Printf("Target: %s\n", *targetURL)
	fmt.Printf("Attack: %s\n", *attackType)
	fmt.Printf("Duration: %v\n", *duration)
	fmt.Printf("Workers: %d\n", *workers)
	fmt.Printf("Target RPS: %d\n", *rps)
	fmt.Println()

	// Select attack based on type
	switch *attackType {
	case "flood":
		runFloodAttack()
	case "replay":
		runReplayAttack()
	case "byzantine":
		runByzantineAttack()
	case "dos":
		runDoSAttack()
	case "arbitrage":
		runArbitrageAttack()
	default:
		fmt.Printf("Unknown attack type: %s\n", *attackType)
		return
	}

	printStatistics()
}

// runFloodAttack floods the DEX with legitimate-looking bids
func runFloodAttack() {
	fmt.Println("Starting flood attack...")

	var wg sync.WaitGroup
	rateLimiter := time.NewTicker(time.Second / time.Duration(*rps))
	defer rateLimiter.Stop()

	stopChan := make(chan bool)
	go func() {
		time.Sleep(*duration)
		close(stopChan)
	}()

	// Start workers
	for i := 0; i < *workers; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()

			for {
				select {
				case <-stopChan:
					return
				case <-rateLimiter.C:
					// Create random bid
					bid := generateRandomBid()
					sendBid(bid)
				}
			}
		}(i)
	}

	wg.Wait()
}

// runReplayAttack captures and replays valid transactions
func runReplayAttack() {
	fmt.Println("Starting replay attack...")

	// First, capture some valid bids
	capturedBids := captureBids(10)
	if len(capturedBids) == 0 {
		fmt.Println("No bids captured, creating synthetic ones")
		for i := 0; i < 10; i++ {
			capturedBids = append(capturedBids, generateRandomBid())
		}
	}

	var wg sync.WaitGroup
	stopChan := make(chan bool)

	go func() {
		time.Sleep(*duration)
		close(stopChan)
	}()

	// Replay captured bids with modifications
	for i := 0; i < *workers; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()

			for {
				select {
				case <-stopChan:
					return
				default:
					// Pick random captured bid and replay with new nonce
					bid := capturedBids[workerID%len(capturedBids)]
					bid.Nonce = generateNonce()
					sendBid(bid)

					time.Sleep(time.Millisecond * 10)
				}
			}
		}(i)
	}

	wg.Wait()
}

// runByzantineAttack simulates Byzantine behavior
func runByzantineAttack() {
	fmt.Println("Starting Byzantine attack...")

	var wg sync.WaitGroup
	stopChan := make(chan bool)

	go func() {
		time.Sleep(*duration)
		close(stopChan)
	}()

	// Create conflicting bids from same bidder
	bidderID := ids.GenerateTestID().String()

	for i := 0; i < *workers; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()

			for {
				select {
				case <-stopChan:
					return
				default:
					// Send multiple conflicting bids for same auction
					auctionID := ids.GenerateTestID().String()

					// Bid 1: High amount
					bid1 := BidRequest{
						AuctionID: auctionID,
						BidderID:  bidderID,
						Amount:    10000,
						Nonce:     generateNonce(),
					}
					sendBid(bid1)

					// Bid 2: Low amount (conflicting)
					bid2 := BidRequest{
						AuctionID: auctionID,
						BidderID:  bidderID,
						Amount:    100,
						Nonce:     generateNonce(),
					}
					sendBid(bid2)

					// Bid 3: Zero amount (invalid)
					bid3 := BidRequest{
						AuctionID: auctionID,
						BidderID:  bidderID,
						Amount:    0,
						Nonce:     generateNonce(),
					}
					sendBid(bid3)

					time.Sleep(time.Millisecond * 100)
				}
			}
		}(i)
	}

	wg.Wait()
}

// runDoSAttack attempts denial of service
func runDoSAttack() {
	fmt.Println("Starting DoS attack...")

	var wg sync.WaitGroup
	stopChan := make(chan bool)

	go func() {
		time.Sleep(*duration)
		close(stopChan)
	}()

	// Multiple attack vectors
	for i := 0; i < *workers/3; i++ {
		// Large payload attack
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case <-stopChan:
					return
				default:
					sendLargePayload()
				}
			}
		}()

		// Malformed request attack
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case <-stopChan:
					return
				default:
					sendMalformedRequest()
				}
			}
		}()

		// Connection exhaustion
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case <-stopChan:
					return
				default:
					openSlowConnection()
				}
			}
		}()
	}

	wg.Wait()
}

// runArbitrageAttack exploits price differences
func runArbitrageAttack() {
	fmt.Println("Starting arbitrage attack...")

	var wg sync.WaitGroup
	stopChan := make(chan bool)

	go func() {
		time.Sleep(*duration)
		close(stopChan)
	}()

	// Monitor auctions and exploit price differences
	for i := 0; i < *workers; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()

			for {
				select {
				case <-stopChan:
					return
				default:
					// Create low reserve auction
					auctionID := createAuction(100) // Low reserve

					if auctionID != "" {
						// Immediately bid slightly above reserve
						bid := BidRequest{
							AuctionID: auctionID,
							BidderID:  ids.GenerateTestID().String(),
							Amount:    101,
							Nonce:     generateNonce(),
						}
						sendBid(bid)

						// Wait for auction to close
						time.Sleep(time.Millisecond * 100)

						// Create high reserve auction for same slot
						createAuction(10000) // High reserve
					}

					time.Sleep(time.Millisecond * 50)
				}
			}
		}(i)
	}

	wg.Wait()
}

// Helper functions

func generateRandomBid() BidRequest {
	return BidRequest{
		AuctionID:    ids.GenerateTestID().String(),
		BidderID:     ids.GenerateTestID().String(),
		Amount:       uint64(randInt(100, 10000)),
		EncryptedBid: generateRandomHex(64),
		Signature:    generateRandomHex(64),
		Nonce:        generateNonce(),
	}
}

func sendBid(bid BidRequest) {
	start := time.Now()

	data, _ := json.Marshal(bid)
	resp, err := http.Post(*targetURL+"/auction/bid", "application/json", bytes.NewReader(data))

	latency := time.Since(start).Microseconds()
	atomic.AddInt64(&totalRequests, 1)
	atomic.AddInt64(&latencySum, latency)

	if latency > atomic.LoadInt64(&maxLatency) {
		atomic.StoreInt64(&maxLatency, latency)
	}

	if err != nil {
		atomic.AddInt64(&errorCount, 1)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		atomic.AddInt64(&successCount, 1)
	} else {
		atomic.AddInt64(&errorCount, 1)
	}
}

func createAuction(reserve uint64) string {
	req := AuctionRequest{
		SlotID:   ids.GenerateTestID().String(),
		Reserve:  reserve,
		Duration: 100,
	}

	data, _ := json.Marshal(req)
	resp, err := http.Post(*targetURL+"/auction/create", "application/json", bytes.NewReader(data))
	if err != nil {
		return ""
	}
	defer resp.Body.Close()

	var result map[string]string
	json.NewDecoder(resp.Body).Decode(&result)
	return result["auction_id"]
}

func captureBids(count int) []BidRequest {
	// In real scenario, this would intercept actual network traffic
	// For simulation, we generate synthetic bids
	var bids []BidRequest
	for i := 0; i < count; i++ {
		bids = append(bids, generateRandomBid())
	}
	return bids
}

func sendLargePayload() {
	// Send 10MB payload
	largeData := make([]byte, 10*1024*1024)
	rand.Read(largeData)

	http.Post(*targetURL+"/auction/bid", "application/octet-stream", bytes.NewReader(largeData))
	atomic.AddInt64(&totalRequests, 1)
}

func sendMalformedRequest() {
	// Send invalid JSON
	malformed := []byte("{invalid json: true, }")
	http.Post(*targetURL+"/auction/bid", "application/json", bytes.NewReader(malformed))
	atomic.AddInt64(&totalRequests, 1)
}

func openSlowConnection() {
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	// Open connection but send data very slowly
	pr, pw := io.Pipe()
	go func() {
		for i := 0; i < 1000; i++ {
			pw.Write([]byte("a"))
			time.Sleep(100 * time.Millisecond)
		}
		pw.Close()
	}()

	req, _ := http.NewRequest("POST", *targetURL+"/auction/bid", pr)
	client.Do(req)
	atomic.AddInt64(&totalRequests, 1)
}

func generateNonce() uint64 {
	nonce, _ := rand.Int(rand.Reader, big.NewInt(1<<62))
	return nonce.Uint64()
}

func generateRandomHex(length int) string {
	bytes := make([]byte, length/2)
	rand.Read(bytes)
	return fmt.Sprintf("%x", bytes)
}

func randInt(min, max int) int {
	n, _ := rand.Int(rand.Reader, big.NewInt(int64(max-min)))
	return int(n.Int64()) + min
}

func printStatistics() {
	fmt.Println("\n=== Attack Statistics ===")
	fmt.Printf("Total Requests:  %d\n", atomic.LoadInt64(&totalRequests))
	fmt.Printf("Successful:      %d\n", atomic.LoadInt64(&successCount))
	fmt.Printf("Errors:          %d\n", atomic.LoadInt64(&errorCount))

	if totalRequests > 0 {
		successRate := float64(successCount) / float64(totalRequests) * 100
		fmt.Printf("Success Rate:    %.2f%%\n", successRate)

		avgLatency := float64(latencySum) / float64(totalRequests) / 1000
		fmt.Printf("Avg Latency:     %.2f ms\n", avgLatency)
		fmt.Printf("Max Latency:     %.2f ms\n", float64(maxLatency)/1000)
	}
}
