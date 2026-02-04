// Copyright (C) 2025, ADXYZ Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package storage

import (
	"time"
)

// ImpressionRecord represents a stored ad impression
type ImpressionRecord struct {
	ID           string
	AdID         string
	PublisherID  string
	PlacementID  string
	AdvertiserID string
	UserID       string
	DeviceID     string
	MinerID      string
	DeviceType   string
	Timestamp    time.Time
	Location     string
	GeoCountry   string
	GeoRegion    string
	Price        float64
	Currency     string
	WinningDSP   string
	AdPodID      string
	VideoURL     string
	Duration     int
	Completed    bool
}

// BidRecord represents a stored bid
type BidRecord struct {
	ID           string
	AuctionID    string
	BidderID     string
	BidRequest   string
	DSPID        string
	BidPrice     float64
	Won          bool
	ResponseTime int64
	CreativeID   string
	Amount       float64
	Currency     string
	Timestamp    time.Time
	Status       string // submitted, won, lost
	ClearPrice   float64
}

// PublisherStats represents aggregated publisher statistics
type PublisherStats struct {
	PublisherID      string
	Period           time.Time // start of period
	Impressions      int64
	TotalImpressions int64
	Revenue          float64
	TotalRevenue     float64
	AvgCPM           float64
	FillRate         float64
	AdRequests       int64
	TimeoutRate      float64
}

// AdvertiserStats represents aggregated advertiser statistics
type AdvertiserStats struct {
	AdvertiserID string
	Period       time.Time
	Impressions  int64
	Clicks       int64
	Spend        float64
	CTR          float64 // Click-through rate
	AvgCPC       float64 // Average cost per click
	Conversions  int64
}
