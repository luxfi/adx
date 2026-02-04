// Copyright (C) 2025, ADXYZ Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package auction

import (
	"testing"
	"time"

	"github.com/luxfi/adx/pkg/crypto"
	"github.com/luxfi/adx/pkg/ids"
	"github.com/luxfi/adx/pkg/log"
)

func BenchmarkAuctionSubmitBid(b *testing.B) {
	logger := log.NoOp()
	auctionID := ids.GenerateTestID()
	reserve := uint64(100)
	duration := 100 * time.Millisecond

	auction := NewAuction(auctionID, reserve, duration, logger)

	// Pre-generate bids
	bids := make([]*SealedBid, b.N)
	for i := 0; i < b.N; i++ {
		bids[i] = &SealedBid{
			BidderID:     ids.GenerateTestID(),
			Commitment:   crypto.CreateCommitment([]byte{byte(i)}),
			EncryptedBid: []byte("encrypted"),
			RangeProof:   []byte("proof"),
			Timestamp:    time.Now(),
		}
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		auction.SubmitBid(bids[i])
	}
}

func BenchmarkAuctionRunWithDecryption(b *testing.B) {
	logger := log.NoOp()

	for i := 0; i < b.N; i++ {
		b.StopTimer()

		auctionID := ids.GenerateTestID()
		reserve := uint64(100)
		duration := 100 * time.Millisecond

		auction := NewAuction(auctionID, reserve, duration, logger)

		// Submit 100 bids
		for j := 0; j < 100; j++ {
			bid := &SealedBid{
				BidderID:     ids.GenerateTestID(),
				Commitment:   crypto.CreateCommitment([]byte{byte(j)}),
				EncryptedBid: []byte("encrypted"),
				RangeProof:   []byte("proof"),
				Timestamp:    time.Now(),
			}
			auction.SubmitBid(bid)
		}

		b.StartTimer()

		// Run auction
		auction.RunAuction([]byte("decryption_key"))
	}
}

func BenchmarkParallelBidSubmission(b *testing.B) {
	logger := log.NoOp()
	auctionID := ids.GenerateTestID()
	reserve := uint64(100)
	duration := 100 * time.Millisecond

	auction := NewAuction(auctionID, reserve, duration, logger)

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			bid := &SealedBid{
				BidderID:     ids.GenerateTestID(),
				Commitment:   crypto.CreateCommitment([]byte("bid")),
				EncryptedBid: []byte("encrypted"),
				RangeProof:   []byte("proof"),
				Timestamp:    time.Now(),
			}
			auction.SubmitBid(bid)
		}
	})
}

func BenchmarkCryptoCommitment(b *testing.B) {
	data := []byte("test_bid_data_12345")

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		crypto.CreateCommitment(data)
	}
}

func BenchmarkHPKEEncryption(b *testing.B) {
	hpke := crypto.NewHPKE()
	pubKey, _, _ := hpke.GenerateKeyPair()
	plaintext := []byte("bid_amount_10000")

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		crypto.EncryptWithHPKE(pubKey, plaintext)
	}
}
