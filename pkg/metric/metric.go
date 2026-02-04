// Copyright (C) 2025, ADXYZ Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package metric

import (
	metrics "github.com/luxfi/metric"
	"github.com/prometheus/client_golang/prometheus"
)

// Metrics holds all metrics for ADX using luxfi/metric
type Metrics struct {
	metricsInstance metrics.Metrics

	// Auction metrics
	AuctionsCreated   metrics.Counter
	BidsSubmitted     metrics.Counter
	AuctionsProcessed metrics.Counter

	// DA metrics
	DAStored    metrics.Counter
	DARetrieved metrics.Counter

	// Network metrics
	PeersConnected metrics.Gauge
	BlocksCreated  metrics.Counter

	// Attack defense metrics
	AttacksDetected   metrics.CounterVec
	AttacksBlocked    metrics.CounterVec
	RequestsProcessed metrics.CounterVec

	// Performance metrics
	AuctionDuration metrics.Histogram
	BidLatency      metrics.Histogram
	DALatency       metrics.Histogram
}

// NewMetrics creates a new metrics instance using luxfi/metric
func NewMetrics() (*Metrics, error) {
	// Create factory and metrics instance
	factory := metrics.NewPrometheusFactory()
	metricsInstance := factory.New("adx")

	m := &Metrics{
		metricsInstance: metricsInstance,
	}

	// Create auction metrics
	m.AuctionsCreated = metricsInstance.NewCounter("auction_created_total", "Total number of auctions created")
	m.BidsSubmitted = metricsInstance.NewCounter("auction_bids_submitted_total", "Total number of bids submitted")
	m.AuctionsProcessed = metricsInstance.NewCounter("auction_processed_total", "Total number of auctions processed")

	// Create DA metrics
	m.DAStored = metricsInstance.NewCounter("da_stored_total", "Total blobs stored to DA layer")
	m.DARetrieved = metricsInstance.NewCounter("da_retrieved_total", "Total blobs retrieved from DA layer")

	// Create network metrics
	m.PeersConnected = metricsInstance.NewGauge("network_peers_connected", "Number of connected peers")
	m.BlocksCreated = metricsInstance.NewCounter("consensus_blocks_created_total", "Total number of blocks created")

	// Create attack defense metrics
	m.AttacksDetected = metricsInstance.NewCounterVec(
		"security_attacks_detected_total",
		"Total number of attacks detected by type",
		[]string{"type"},
	)

	m.AttacksBlocked = metricsInstance.NewCounterVec(
		"security_attacks_blocked_total",
		"Total number of attacks blocked by type",
		[]string{"type"},
	)

	m.RequestsProcessed = metricsInstance.NewCounterVec(
		"api_requests_processed_total",
		"Total number of API requests processed",
		[]string{"method", "status"},
	)

	// Create performance metrics
	m.AuctionDuration = metricsInstance.NewHistogram(
		"auction_duration_seconds",
		"Time to process an auction",
		prometheus.DefBuckets,
	)

	m.BidLatency = metricsInstance.NewHistogram(
		"auction_bid_latency_seconds",
		"Time to process a bid",
		prometheus.DefBuckets,
	)

	m.DALatency = metricsInstance.NewHistogram(
		"da_latency_seconds",
		"Time to store/retrieve from DA layer",
		prometheus.DefBuckets,
	)

	return m, nil
}

// GetGatherer returns the prometheus gatherer for metrics export
func (m *Metrics) GetGatherer() prometheus.Gatherer {
	if registry := m.metricsInstance.Registry(); registry != nil {
		return registry
	}
	return prometheus.DefaultGatherer
}

// GetRegisterer returns the prometheus registerer
func (m *Metrics) GetRegisterer() prometheus.Registerer {
	if registry := m.metricsInstance.Registry(); registry != nil {
		return registry
	}
	return prometheus.DefaultRegisterer
}
