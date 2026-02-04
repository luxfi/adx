// Copyright (C) 2025, ADXYZ Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package tee

import (
	"testing"

	"github.com/luxfi/adx/pkg/ids"
	"github.com/luxfi/adx/pkg/log"
	"github.com/stretchr/testify/require"
)

func TestEnclaveCreation(t *testing.T) {
	require := require.New(t)
	logger := log.NoOp()

	// Test simulated enclave
	enclave, err := NewEnclave(EnclaveSimulated, logger)
	require.NoError(err)
	require.NotNil(enclave)
	require.Equal(EnclaveSimulated, enclave.Type)

	// Verify attestation
	require.NotEmpty(enclave.attestation)
	require.NotZero(enclave.createdAt)
}

func TestEnclaveAuction(t *testing.T) {
	require := require.New(t)
	logger := log.NoOp()

	enclave, err := NewEnclave(EnclaveSimulated, logger)
	require.NoError(err)

	auctionID := ids.GenerateTestID()
	reserve := uint64(100)

	// Create encrypted bids (simplified)
	encryptedBids := [][]byte{
		[]byte("encrypted_bid_1"),
		[]byte("encrypted_bid_2"),
		[]byte("encrypted_bid_3"),
	}

	result, err := enclave.RunAuction(auctionID, reserve, encryptedBids)
	require.NoError(err)
	require.NotNil(result)

	// Check result structure
	require.NotEmpty(result.WinnerID)
	require.Greater(result.ClearingPrice, uint64(0))
	require.NotEmpty(result.Proof)
	require.NotZero(result.ProcessedAt)
}

func TestEnclaveFrequencyCapping(t *testing.T) {
	require := require.New(t)
	logger := log.NoOp()

	enclave, err := NewEnclave(EnclaveSimulated, logger)
	require.NoError(err)

	userID := "user123"
	campaignID := "campaign456"

	// First impression
	allowed, err := enclave.CheckFrequencyCap(userID, campaignID, 3)
	require.NoError(err)
	require.True(allowed)

	// Second impression
	allowed, err = enclave.CheckFrequencyCap(userID, campaignID, 3)
	require.NoError(err)
	require.True(allowed)

	// Third impression
	allowed, err = enclave.CheckFrequencyCap(userID, campaignID, 3)
	require.NoError(err)
	require.True(allowed)

	// Fourth impression (should be capped)
	allowed, err = enclave.CheckFrequencyCap(userID, campaignID, 3)
	require.NoError(err)
	require.False(allowed)
}

func TestEnclaveConcurrentOperations(t *testing.T) {
	require := require.New(t)
	logger := log.NoOp()

	enclave, err := NewEnclave(EnclaveSimulated, logger)
	require.NoError(err)

	numOperations := 100
	done := make(chan bool, numOperations)

	// Run concurrent auctions
	for i := 0; i < numOperations; i++ {
		go func(index int) {
			auctionID := ids.GenerateTestID()
			reserve := uint64(100 + index)

			encryptedBids := [][]byte{
				[]byte("bid1"),
				[]byte("bid2"),
			}

			result, err := enclave.RunAuction(auctionID, reserve, encryptedBids)
			require.NoError(err)
			require.NotNil(result)

			done <- true
		}(i)
	}

	// Wait for all operations
	for i := 0; i < numOperations; i++ {
		<-done
	}
}

func TestEnclaveAttestation(t *testing.T) {
	require := require.New(t)
	logger := log.NoOp()

	enclave, err := NewEnclave(EnclaveSimulated, logger)
	require.NoError(err)

	// Verify attestation is properly set
	require.NotEmpty(enclave.attestation)

	// For simulated mode, attestation should be mock
	require.Contains(string(enclave.attestation), "SIMULATED")
}

func TestEnclaveSecureStorage(t *testing.T) {
	require := require.New(t)
	logger := log.NoOp()

	enclave, err := NewEnclave(EnclaveSimulated, logger)
	require.NoError(err)

	// Store sensitive data
	key := "secret_key_123"
	value := []byte("sensitive_data")

	err = enclave.StoreSecure(key, value)
	require.NoError(err)

	// Retrieve sensitive data
	retrieved, err := enclave.RetrieveSecure(key)
	require.NoError(err)
	require.Equal(value, retrieved)
}

func BenchmarkEnclaveAuction(b *testing.B) {
	logger := log.NoOp()
	enclave, _ := NewEnclave(EnclaveSimulated, logger)

	auctionID := ids.GenerateTestID()
	reserve := uint64(100)
	encryptedBids := [][]byte{
		[]byte("bid1"),
		[]byte("bid2"),
		[]byte("bid3"),
		[]byte("bid4"),
		[]byte("bid5"),
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		enclave.RunAuction(auctionID, reserve, encryptedBids)
	}
}

func BenchmarkEnclaveFrequencyCheck(b *testing.B) {
	logger := log.NoOp()
	enclave, _ := NewEnclave(EnclaveSimulated, logger)

	userID := "user123"
	campaignID := "campaign456"

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		enclave.CheckFrequencyCap(userID, campaignID, 10)
	}
}

func BenchmarkEnclaveSecureStorage(b *testing.B) {
	logger := log.NoOp()
	enclave, _ := NewEnclave(EnclaveSimulated, logger)

	value := []byte("sensitive_data_12345")

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		key := ids.GenerateTestID().String()
		enclave.StoreSecure(key, value)
	}
}
