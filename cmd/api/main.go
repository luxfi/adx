package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"math/big"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
	"github.com/luxfi/adx/pkg/rtb"
	"github.com/luxfi/adx/pkg/vast"
)

var (
	port     = flag.String("port", "8080", "API server port")
	env      = flag.String("env", "development", "Environment (development/production)")
	rtbURL   = flag.String("rtb", "http://localhost:9090", "RTB exchange URL")
	cdnURL   = flag.String("cdn", "https://cdn.lux.network", "CDN base URL")
)

func main() {
	flag.Parse()

	// Initialize RTB exchange wrapper
	exchange := &RTBExchangeWrapper{
		rtbExchange: &rtb.RTBExchange{
			AuctionTimeout: 100 * time.Millisecond,
			FloorPrice:     decimal.NewFromFloat(0.01),
			DSPs:           make(map[string]*rtb.DSPConnection),
			SSPs:           make(map[string]*rtb.SSPConnection),
			Revenue:        big.NewInt(0),
		},
	}

	// Initialize mock DSPs for testing
	initMockDSPs(exchange.rtbExchange)

	// Create VAST handler
	vastHandler := &vast.VASTHandler{
		Exchange:      exchange,
		Storage:       &MockStorage{},
		Analytics:     &MockAnalytics{},
		PrivacyMgr:    &MockPrivacy{},
		BlockchainMgr: &MockBlockchain{},
	}

	// Setup Gin router
	router := setupRouter(vastHandler, exchange)

	// Start server
	srv := &http.Server{
		Addr:    ":" + *port,
		Handler: router,
	}

	// Graceful shutdown
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	log.Printf("🚀 ADX API Server started on port %s", *port)
	log.Printf("📡 RTB Exchange: %s", *rtbURL)
	log.Printf("🌐 CDN: %s", *cdnURL)
	log.Printf("🔧 Environment: %s", *env)

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}

	log.Println("Server exiting")
}

func setupRouter(vastHandler *vast.VASTHandler, exchange *RTBExchangeWrapper) *gin.Engine {
	if *env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.Default()

	// CORS configuration
	config := cors.DefaultConfig()
	config.AllowOrigins = []string{"http://localhost:3000", "http://localhost:3001", "https://lux.network", "https://app.lux.network"}
	config.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
	config.AllowHeaders = []string{"Origin", "Content-Type", "Accept", "Authorization"}
	router.Use(cors.New(config))

	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "healthy",
			"time":   time.Now().Unix(),
		})
	})

	// API routes
	api := router.Group("/api/v1")
	{
		// VAST endpoint (PubNative compatible)
		api.GET("/vast", vastHandler.HandleVASTRequest)
		api.POST("/vast", vastHandler.HandleVASTRequest)

		// Campaign management
		api.POST("/campaigns", createCampaign)
		api.GET("/campaigns", listCampaigns)
		api.GET("/campaigns/:id", getCampaign)
		api.PUT("/campaigns/:id", updateCampaign)
		api.DELETE("/campaigns/:id", deleteCampaign)

		// Creative management
		api.POST("/creatives", uploadCreative)
		api.GET("/creatives", listCreatives)
		api.GET("/creatives/:id", getCreative)

		// Reporting
		api.GET("/reports/impressions", getImpressionReport)
		api.GET("/reports/revenue", getRevenueReport)
		api.GET("/reports/performance", getPerformanceReport)

		// Wallet integration
		api.POST("/wallet/connect", connectWallet)
		api.POST("/wallet/deposit", depositFunds)
		api.POST("/wallet/withdraw", withdrawFunds)
		api.GET("/wallet/balance", getWalletBalance)

		// RTB endpoints
		api.POST("/rtb/bid", handleBidRequest)
		api.GET("/rtb/stats", getRTBStats(exchange))
	}

	// Static files for creatives
	router.Static("/creatives", "./static/creatives")

	return router
}

// Campaign handlers
func createCampaign(c *gin.Context) {
	var req struct {
		Name        string  `json:"name" binding:"required"`
		Budget      float64 `json:"budget" binding:"required"`
		CPM         float64 `json:"cpm" binding:"required"`
		StartDate   string  `json:"start_date" binding:"required"`
		EndDate     string  `json:"end_date" binding:"required"`
		Targeting   gin.H   `json:"targeting"`
		CreativeIDs []string `json:"creative_ids"`
		WalletAddress string `json:"wallet_address"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	// Create campaign in database
	campaign := gin.H{
		"id":         fmt.Sprintf("camp_%d", time.Now().Unix()),
		"name":       req.Name,
		"budget":     req.Budget,
		"cpm":        req.CPM,
		"start_date": req.StartDate,
		"end_date":   req.EndDate,
		"targeting":  req.Targeting,
		"creative_ids": req.CreativeIDs,
		"wallet_address": req.WalletAddress,
		"status":     "active",
		"created_at": time.Now(),
	}

	c.JSON(201, campaign)
}

func listCampaigns(c *gin.Context) {
	// Mock campaigns for demo
	campaigns := []gin.H{
		{
			"id":         "camp_1",
			"name":       "Holiday Sale Campaign",
			"budget":     10000.0,
			"spent":      2500.0,
			"impressions": 1250000,
			"clicks":     12500,
			"ctr":        0.01,
			"status":     "active",
		},
		{
			"id":         "camp_2",
			"name":       "Brand Awareness",
			"budget":     5000.0,
			"spent":      3200.0,
			"impressions": 1600000,
			"clicks":     8000,
			"ctr":        0.005,
			"status":     "active",
		},
	}

	c.JSON(200, gin.H{
		"campaigns": campaigns,
		"total":     len(campaigns),
	})
}

func getCampaign(c *gin.Context) {
	id := c.Param("id")
	
	campaign := gin.H{
		"id":         id,
		"name":       "Test Campaign",
		"budget":     5000.0,
		"spent":      1200.0,
		"impressions": 600000,
		"clicks":     3000,
		"ctr":        0.005,
		"status":     "active",
		"targeting": gin.H{
			"geos":       []string{"US", "CA"},
			"devices":    []string{"ctv", "mobile"},
			"categories": []string{"entertainment", "sports"},
		},
	}

	c.JSON(200, campaign)
}

func updateCampaign(c *gin.Context) {
	id := c.Param("id")
	var req gin.H
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	req["id"] = id
	req["updated_at"] = time.Now()

	c.JSON(200, req)
}

func deleteCampaign(c *gin.Context) {
	id := c.Param("id")
	c.JSON(200, gin.H{
		"message": "Campaign deleted",
		"id":      id,
	})
}

// Creative handlers
func uploadCreative(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(400, gin.H{"error": "No file uploaded"})
		return
	}

	// Save file (in production, upload to CDN)
	filename := fmt.Sprintf("creative_%d_%s", time.Now().Unix(), file.Filename)
	
	creative := gin.H{
		"id":       fmt.Sprintf("cre_%d", time.Now().Unix()),
		"filename": filename,
		"url":      fmt.Sprintf("%s/creatives/%s", *cdnURL, filename),
		"type":     c.PostForm("type"),
		"duration": c.PostForm("duration"),
		"size":     file.Size,
		"created_at": time.Now(),
	}

	c.JSON(201, creative)
}

func listCreatives(c *gin.Context) {
	creatives := []gin.H{
		{
			"id":       "cre_1",
			"name":     "Holiday Video 30s",
			"type":     "video",
			"duration": 30,
			"url":      fmt.Sprintf("%s/creatives/holiday_30s.mp4", *cdnURL),
			"size":     5242880,
		},
		{
			"id":       "cre_2",
			"name":     "Brand Video 15s",
			"type":     "video",
			"duration": 15,
			"url":      fmt.Sprintf("%s/creatives/brand_15s.mp4", *cdnURL),
			"size":     2621440,
		},
	}

	c.JSON(200, gin.H{
		"creatives": creatives,
		"total":     len(creatives),
	})
}

func getCreative(c *gin.Context) {
	id := c.Param("id")
	
	creative := gin.H{
		"id":       id,
		"name":     "Test Creative",
		"type":     "video",
		"duration": 30,
		"url":      fmt.Sprintf("%s/creatives/test.mp4", *cdnURL),
		"size":     5242880,
	}

	c.JSON(200, creative)
}

// Reporting handlers
func getImpressionReport(c *gin.Context) {
	report := gin.H{
		"period": gin.H{
			"start": c.Query("start_date"),
			"end":   c.Query("end_date"),
		},
		"data": []gin.H{
			{"date": "2024-08-25", "impressions": 125000, "clicks": 1250, "ctr": 0.01},
			{"date": "2024-08-26", "impressions": 135000, "clicks": 1485, "ctr": 0.011},
			{"date": "2024-08-27", "impressions": 142000, "clicks": 1562, "ctr": 0.011},
		},
		"totals": gin.H{
			"impressions": 402000,
			"clicks":      4297,
			"ctr":         0.0107,
		},
	}

	c.JSON(200, report)
}

func getRevenueReport(c *gin.Context) {
	report := gin.H{
		"period": gin.H{
			"start": c.Query("start_date"),
			"end":   c.Query("end_date"),
		},
		"data": []gin.H{
			{"date": "2024-08-25", "revenue": 250.00, "cost": 200.00, "profit": 50.00},
			{"date": "2024-08-26", "revenue": 270.00, "cost": 216.00, "profit": 54.00},
			{"date": "2024-08-27", "revenue": 284.00, "cost": 227.20, "profit": 56.80},
		},
		"totals": gin.H{
			"revenue": 804.00,
			"cost":    643.20,
			"profit":  160.80,
		},
	}

	c.JSON(200, report)
}

func getPerformanceReport(c *gin.Context) {
	report := gin.H{
		"campaigns": []gin.H{
			{
				"id":          "camp_1",
				"name":        "Holiday Campaign",
				"impressions": 250000,
				"clicks":      2500,
				"ctr":         0.01,
				"cpm":         2.00,
				"spend":       500.00,
			},
		},
		"creatives": []gin.H{
			{
				"id":          "cre_1",
				"name":        "Holiday Video",
				"impressions": 150000,
				"clicks":      1800,
				"ctr":         0.012,
				"vcr":         0.85, // Video completion rate
			},
		},
	}

	c.JSON(200, report)
}

// Wallet handlers
func connectWallet(c *gin.Context) {
	var req struct {
		Address   string `json:"address" binding:"required"`
		ChainID   int    `json:"chain_id" binding:"required"`
		Signature string `json:"signature" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{
		"address": req.Address,
		"chain_id": req.ChainID,
		"connected": true,
		"balance": 1000.0, // Mock balance
	})
}

func depositFunds(c *gin.Context) {
	var req struct {
		Amount   float64 `json:"amount" binding:"required"`
		TxHash   string  `json:"tx_hash"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{
		"amount": req.Amount,
		"tx_hash": req.TxHash,
		"status": "confirmed",
		"new_balance": 1000.0 + req.Amount,
	})
}

func withdrawFunds(c *gin.Context) {
	var req struct {
		Amount   float64 `json:"amount" binding:"required"`
		Address  string  `json:"address" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{
		"amount": req.Amount,
		"address": req.Address,
		"tx_hash": fmt.Sprintf("0x%x", time.Now().Unix()),
		"status": "pending",
	})
}

func getWalletBalance(c *gin.Context) {
	c.JSON(200, gin.H{
		"balance": 1000.0,
		"pending": 50.0,
		"currency": "USDC",
	})
}

// RTB handlers
func handleBidRequest(c *gin.Context) {
	var req gin.H
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	// Mock bid response
	c.JSON(200, gin.H{
		"id": req["id"],
		"seatbid": []gin.H{
			{
				"bid": []gin.H{
					{
						"id":    fmt.Sprintf("bid_%d", time.Now().Unix()),
						"impid": "1",
						"price": 2.50,
						"adm":   "<VAST>...</VAST>",
					},
				},
			},
		},
	})
}

func getRTBStats(exchange *RTBExchangeWrapper) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(200, gin.H{
			"impressions": exchange.rtbExchange.ImpressionCount,
			"bids":        exchange.rtbExchange.BidCount,
			"wins":        exchange.rtbExchange.WinCount,
			"revenue":     exchange.rtbExchange.Revenue.String(),
			"dsps":        len(exchange.rtbExchange.DSPs),
			"ssps":        len(exchange.rtbExchange.SSPs),
		})
	}
}

// Mock implementations for testing
type MockStorage struct{}

func (m *MockStorage) StoreImpression(imp *vast.ImpressionRecord) error {
	return nil
}

func (m *MockStorage) GetImpression(id string) (*vast.ImpressionRecord, error) {
	return &vast.ImpressionRecord{}, nil
}

type MockAnalytics struct{}

func (m *MockAnalytics) TrackImpression(imp *vast.ImpressionRecord) {}
func (m *MockAnalytics) TrackClick(clickID, impID string) {}
func (m *MockAnalytics) GetMetrics(start, end time.Time) map[string]interface{} {
	return map[string]interface{}{}
}

type MockPrivacy struct{}

func (m *MockPrivacy) CheckCompliance(consent string, gdpr int, ccpa string) bool {
	return true
}
func (m *MockPrivacy) AnonymizeData(data interface{}) interface{} {
	return data
}

type MockBlockchain struct{}

func (m *MockBlockchain) RecordImpression(imp *vast.ImpressionRecord, wallet string, chainID int) error {
	return nil
}
func (m *MockBlockchain) VerifyProofOfView(pov string) bool {
	return true
}
func (m *MockBlockchain) ProcessPayment(wallet string, amount float64, chainID int) error {
	return nil
}

// RTBExchangeWrapper wraps rtb.RTBExchange to implement vast.RTBExchange interface
type RTBExchangeWrapper struct {
	rtbExchange *rtb.RTBExchange
}

// RunAuction implements vast.RTBExchange interface
func (w *RTBExchangeWrapper) RunAuction(ctx context.Context, req *vast.OpenRTBRequest) (*vast.OpenRTBResponse, error) {
	// Simple mock auction
	return &vast.OpenRTBResponse{
		ID: req.ID,
		SeatBid: []vast.SeatBid{
			{
				Seat: "dsp1",
				Bid: []vast.Bid{
					{
						ID:      fmt.Sprintf("bid_%d", time.Now().Unix()),
						ImpID:   "1",
						Price:   2.50,
						ADomain: []string{"example.com"},
						ADURL:   fmt.Sprintf("%s/creatives/test.mp4", *cdnURL),
						NURL:    "https://example.com/click",
						Cur:     "USD",
					},
				},
			},
		},
	}, nil
}

func initMockDSPs(exchange *rtb.RTBExchange) {
	// Add some mock DSPs for testing
	exchange.DSPs["dsp1"] = &rtb.DSPConnection{
		ID:          "dsp1",
		Name:        "Test DSP 1",
		Endpoint:    "http://localhost:9001",
		QPS:         100,
		Timeout:     100 * time.Millisecond,
		BidderCode:  "td1",
		SeatID:      "seat1",
		RateLimiter: NewRateLimiter(100),
	}

	exchange.DSPs["dsp2"] = &rtb.DSPConnection{
		ID:          "dsp2",
		Name:        "Test DSP 2",
		Endpoint:    "http://localhost:9002",
		QPS:         100,
		Timeout:     100 * time.Millisecond,
		BidderCode:  "td2",
		SeatID:      "seat2",
		RateLimiter: NewRateLimiter(100),
	}
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(qps int) *rtb.RateLimiter {
	return &rtb.RateLimiter{}
}