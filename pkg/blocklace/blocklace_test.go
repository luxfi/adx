// Copyright (C) 2025, ADXYZ Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package blocklace

import (
	"testing"
	"time"
	
	"github.com/luxfi/adx/pkg/core"
	"github.com/luxfi/adx/pkg/ids"
	"github.com/luxfi/adx/pkg/log"
	"github.com/stretchr/testify/require"
)

func TestDAGBasicOperations(t *testing.T) {
	require := require.New(t)
	logger := log.NoOp()
	
	dag := NewDAG(logger)
	require.NotNil(dag)
	
	// Add first header
	header1 := &core.BaseHeader{
		Type:      core.HeaderTypeAuction,
		ID:        ids.GenerateTestID(),
		Timestamp: time.Now(),
		Height:    1,
	}
	
	nodeID1 := ids.GenerateNodeID()
	err := dag.AddHeader(header1, nodeID1)
	require.NoError(err)
	
	// Add second header
	header2 := &core.BaseHeader{
		Type:      core.HeaderTypeSettlement,
		ID:        ids.GenerateTestID(),
		Timestamp: time.Now(),
		Height:    2,
	}
	
	nodeID2 := ids.GenerateNodeID()
	err = dag.AddHeader(header2, nodeID2)
	require.NoError(err)
	
	// Check sequence (includes genesis + 2 headers = 3 total)
	sequence := dag.GetSequence()
	require.Len(sequence, 3)
	
	// Check metrics (genesis + 2 headers = 3 vertices)
	metrics := dag.GetMetrics()
	require.Equal(3, metrics.Vertices)
	require.GreaterOrEqual(metrics.Delivered, 3)
}

func TestCordialMinerBasic(t *testing.T) {
	require := require.New(t)
	logger := log.NoOp()
	
	dag := NewDAG(logger)
	nodeID := ids.GenerateNodeID()
	
	miner := NewCordialMiner(nodeID, dag, logger)
	require.NotNil(miner)
	
	// Check miner properties
	require.Equal(nodeID, miner.ID)
	require.NotNil(miner.DAG)
}

func TestDAGConcurrentAdditions(t *testing.T) {
	require := require.New(t)
	logger := log.NoOp()
	
	dag := NewDAG(logger)
	
	// Add headers concurrently
	numHeaders := 100
	done := make(chan bool, numHeaders)
	
	for i := 0; i < numHeaders; i++ {
		go func(index int) {
			header := &core.BaseHeader{
				Type:      core.HeaderTypeAuction,
				ID:        ids.GenerateTestID(),
				Timestamp: time.Now(),
				Height:    uint64(index),
			}
			
			nodeID := ids.GenerateNodeID()
			err := dag.AddHeader(header, nodeID)
			require.NoError(err)
			done <- true
		}(i)
	}
	
	// Wait for all additions
	for i := 0; i < numHeaders; i++ {
		<-done
	}
	
	// Check metrics (genesis + numHeaders = 101 vertices)
	metrics := dag.GetMetrics()
	require.Equal(numHeaders+1, metrics.Vertices)
}

func TestVertexValidation(t *testing.T) {
	require := require.New(t)
	logger := log.NoOp()
	
	dag := NewDAG(logger)
	
	// Add a header with very high height (should work)
	header := &core.BaseHeader{
		ID:     ids.GenerateTestID(),
		Height: 100,
		Timestamp: time.Now(),
	}
	
	nodeID := ids.GenerateNodeID()
	err := dag.AddHeader(header, nodeID)
	// May return error for orphan block
	if err != nil {
		require.Equal(ErrOrphanBlock, err)
	}
}

func TestDAGReorganization(t *testing.T) {
	require := require.New(t)
	logger := log.NoOp()
	
	dag := NewDAG(logger)
	
	// Build initial chain
	for i := 0; i < 10; i++ {
		header := &core.BaseHeader{
			Type:      core.HeaderTypeAuction,
			ID:        ids.GenerateTestID(),
			Timestamp: time.Now(),
			Height:    uint64(i + 1),
		}
		
		nodeID := ids.GenerateNodeID()
		err := dag.AddHeader(header, nodeID)
		require.NoError(err)
	}
	
	// Create more headers from different nodes
	for i := 10; i < 15; i++ {
		header := &core.BaseHeader{
			Type:      core.HeaderTypeAuction,
			ID:        ids.GenerateTestID(),
			Timestamp: time.Now(),
			Height:    uint64(i + 1),
		}
		
		nodeID := ids.GenerateNodeID()
		err := dag.AddHeader(header, nodeID)
		require.NoError(err)
	}
	
	// The DAG should handle the fork
	sequence := dag.GetSequence()
	require.Greater(len(sequence), 10)
}

func BenchmarkDAGAddHeader(b *testing.B) {
	logger := log.NoOp()
	dag := NewDAG(logger)
	
	headers := make([]*core.BaseHeader, b.N)
	nodeIDs := make([]ids.NodeID, b.N)
	
	for i := 0; i < b.N; i++ {
		headers[i] = &core.BaseHeader{
			Type:      core.HeaderTypeAuction,
			ID:        ids.GenerateTestID(),
			Timestamp: time.Now(),
			Height:    uint64(i + 1),
		}
		nodeIDs[i] = ids.GenerateNodeID()
	}
	
	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		dag.AddHeader(headers[i], nodeIDs[i])
	}
}

func BenchmarkCordialMinerCreation(b *testing.B) {
	logger := log.NoOp()
	dag := NewDAG(logger)
	
	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		nodeID := ids.GenerateNodeID()
		NewCordialMiner(nodeID, dag, logger)
	}
}