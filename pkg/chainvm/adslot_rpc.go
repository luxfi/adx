package chainvm

import (
	"context"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"math"
	"time"

	"github.com/luxfi/adx/pkg/dex"
	"github.com/shopspring/decimal"
)

// PendingRelease represents a time-locked fund release
type PendingRelease struct {
	Publisher   string          `json:"publisher"`
	Amount      decimal.Decimal `json:"amount"`
	ReleaseTime time.Time       `json:"release_time"`
}

// VMState represents the state of the VM
type VMState struct {
	adSlots          map[uint64]*AdSlot
	adSlotOrders     map[string]*AdSlotOrder
	adMM_Pools       map[uint64]*AdMM_Pool
	campaigns        map[string]*Campaign
	reservations     map[string]*Reservation
	publisherBalances map[string]decimal.Decimal
	pendingReleases  []PendingRelease
}

// AdMM_Pool represents an automated market maker pool for ad slots
type AdMM_Pool struct {
	SlotID        uint64          `json:"slot_id"`
	Reserve0      decimal.Decimal `json:"reserve0"`       // Ad slot tokens
	Reserve1      decimal.Decimal `json:"reserve1"`       // AUSD tokens
	ReserveAUSD   decimal.Decimal `json:"reserve_ausd"`   // AUSD liquidity
	ReserveSlots  uint64          `json:"reserve_slots"`  // Ad slot supply
	K             decimal.Decimal `json:"k"`              // Constant product
	TotalLP       decimal.Decimal `json:"total_lp"`       // Total LP tokens
	LPTokenSupply decimal.Decimal `json:"lp_token_supply"`
	LastPrice     decimal.Decimal `json:"last_price"`
	TimeDecayRate decimal.Decimal `json:"time_decay_rate"` // λ in pricing formula
	CreatedAt     time.Time       `json:"created_at"`
}

// SetAdSlot stores an ad slot in the state
func (v *VMState) SetAdSlot(slot *AdSlot) error {
	if v.adSlots == nil {
		v.adSlots = make(map[uint64]*AdSlot)
	}
	v.adSlots[slot.ID] = slot
	return nil
}

// GetAdSlot retrieves an ad slot from the state
func (v *VMState) GetAdSlot(id uint64) (*AdSlot, error) {
	if v.adSlots == nil {
		return nil, fmt.Errorf("ad slot not found")
	}
	slot, ok := v.adSlots[id]
	if !ok {
		return nil, fmt.Errorf("ad slot not found")
	}
	return slot, nil
}

// SetAdSlotOrder stores an order in the state
func (v *VMState) SetAdSlotOrder(order *AdSlotOrder) error {
	if v.adSlotOrders == nil {
		v.adSlotOrders = make(map[string]*AdSlotOrder)
	}
	v.adSlotOrders[order.OrderID] = order
	return nil
}

// GetAdSlotOrder retrieves an order from the state
func (v *VMState) GetAdSlotOrder(orderID string) (*AdSlotOrder, error) {
	if v.adSlotOrders == nil {
		return nil, fmt.Errorf("order not found")
	}
	order, ok := v.adSlotOrders[orderID]
	if !ok {
		return nil, fmt.Errorf("order not found")
	}
	return order, nil
}

// SetAdMM_Pool stores an AMM pool in the state
func (v *VMState) SetAdMM_Pool(slotID uint64, pool *AdMM_Pool) error {
	if v.adMM_Pools == nil {
		v.adMM_Pools = make(map[uint64]*AdMM_Pool)
	}
	v.adMM_Pools[slotID] = pool
	return nil
}

// GetAdMM_Pool retrieves an AMM pool from the state
func (v *VMState) GetAdMM_Pool(slotID uint64) (*AdMM_Pool, bool) {
	if v.adMM_Pools == nil {
		return nil, false
	}
	pool, ok := v.adMM_Pools[slotID]
	return pool, ok
}

// SetCampaign stores a campaign in the state
func (v *VMState) SetCampaign(campaignID string, campaign *Campaign) error {
	if v.campaigns == nil {
		v.campaigns = make(map[string]*Campaign)
	}
	v.campaigns[campaignID] = campaign
	return nil
}

// GetCampaign retrieves a campaign from the state
func (v *VMState) GetCampaign(campaignID string) (*Campaign, bool) {
	if v.campaigns == nil {
		return nil, false
	}
	campaign, ok := v.campaigns[campaignID]
	return campaign, ok
}

// SetReservation stores a reservation in the state
func (v *VMState) SetReservation(reservationID string, reservation *Reservation) error {
	if v.reservations == nil {
		v.reservations = make(map[string]*Reservation)
	}
	v.reservations[reservationID] = reservation
	return nil
}

// GetReservation retrieves a reservation from the state
func (v *VMState) GetReservation(reservationID string) (*Reservation, bool) {
	if v.reservations == nil {
		return nil, false
	}
	reservation, ok := v.reservations[reservationID]
	return reservation, ok
}

// SetPublisherBalance sets a publisher's balance
func (v *VMState) SetPublisherBalance(publisher string, balance decimal.Decimal) error {
	if v.publisherBalances == nil {
		v.publisherBalances = make(map[string]decimal.Decimal)
	}
	v.publisherBalances[publisher] = balance
	return nil
}

// GetPublisherBalance gets a publisher's balance
func (v *VMState) GetPublisherBalance(publisher string) decimal.Decimal {
	if v.publisherBalances == nil {
		return decimal.Zero
	}
	balance, ok := v.publisherBalances[publisher]
	if !ok {
		return decimal.Zero
	}
	return balance
}

// AddPendingRelease adds a pending release to the queue
func (v *VMState) AddPendingRelease(publisher string, amount decimal.Decimal, releaseTime time.Time) error {
	release := PendingRelease{
		Publisher:   publisher,
		Amount:      amount,
		ReleaseTime: releaseTime,
	}
	v.pendingReleases = append(v.pendingReleases, release)
	return nil
}

// Request and response types for RPC methods
type RevealBidRequest struct {
	AuctionID     string          `json:"auction_id"`
	BidID         string          `json:"bid_id"`
	OrderID       string          `json:"order_id"`
	Reveal        []byte          `json:"reveal"`
	RevealedPrice decimal.Decimal `json:"revealed_price"`
	Nonce         string          `json:"nonce"`
}

type RevealBidResponse struct {
	Success       bool            `json:"success"`
	Message       string          `json:"message"`
	RevealedPrice decimal.Decimal `json:"revealed_price,omitempty"`
}

type CreateAdMM_PoolRequest struct {
	TokenA            string          `json:"token_a"`
	TokenB            string          `json:"token_b"`
	AmountA           decimal.Decimal `json:"amount_a"`
	AmountB           decimal.Decimal `json:"amount_b"`
	SlotID            uint64          `json:"slot_id"`
	InitialAUSD       decimal.Decimal `json:"initial_ausd"`
	InitialSlots      uint64          `json:"initial_slots"`
	TimeDecayRate     decimal.Decimal `json:"time_decay_rate"`
	LiquidityProvider string          `json:"liquidity_provider"`
}

type CreateAdMM_PoolResponse struct {
	PoolID       string          `json:"pool_id"`
	Success      bool            `json:"success"`
	Message      string          `json:"message"`
	LPTokens     decimal.Decimal `json:"lp_tokens"`
	InitialPrice decimal.Decimal `json:"initial_price"`
}

type SwapAdMM_Request struct {
	PoolID            string          `json:"pool_id"`
	SlotID            uint64          `json:"slot_id"`
	TokenIn           string          `json:"token_in"`
	AmountIn          decimal.Decimal `json:"amount_in"`
	MinAmountOut      decimal.Decimal `json:"min_amount_out"`
	BuyAUSD           bool            `json:"buy_ausd"`
	ExpectedAmountOut decimal.Decimal `json:"expected_amount_out"`
}

type SwapAdMM_Response struct {
	AmountOut      decimal.Decimal `json:"amount_out"`
	Success        bool            `json:"success"`
	Message        string          `json:"message"`
	NewPrice       decimal.Decimal `json:"new_price"`
	SlippageActual decimal.Decimal `json:"slippage_actual"`
}

type RecordDeliveryRequest struct {
	AdSlotID    uint64    `json:"ad_slot_id"`
	SlotID      uint64    `json:"slot_id"`      // Alias for AdSlotID
	Impressions uint64    `json:"impressions"`
	Count       uint64    `json:"count"`        // Alias for Impressions
	Timestamp   time.Time `json:"timestamp"`
}

type RecordDeliveryResponse struct {
	Success         bool   `json:"success"`
	Message         string `json:"message"`
	DeliveredCount  uint64 `json:"delivered_count"`
	TotalDelivered  uint64 `json:"total_delivered"`
	RemainingSupply uint64 `json:"remaining_supply"`
}


// convertToGDexOrder converts an AdSlotOrder to a dex.Order
func convertToGDexOrder(order *AdSlotOrder, slot *AdSlot) *dex.Order {
	return &dex.Order{
		OrderID:  order.OrderID,
		AssetID:  fmt.Sprintf("adslot-%d", slot.ID),
		Price:    order.Price,
		Quantity: decimal.NewFromInt(int64(order.Quantity)),
		IsBuy:    order.OrderType == "buy",
	}
}

// AdSlotManager - Semi-Fungible Tokens for perishable ad inventory
// Implements high-performance DEX primitives with time-decay pricing
type AdSlotManager struct {
	state  *VMState
	dex    *dex.Engine
	nextID uint64
}

// estimateOrderFill estimates how much of an order will be filled
func (a *AdSlotManager) estimateOrderFill(order *AdSlotOrder, slot *AdSlot) uint64 {
	// Simplified estimation - in production would check order book depth
	if order.OrderType == "buy" && order.Price.GreaterThanOrEqual(slot.FloorCPM) {
		return order.Quantity // Assume full fill if above floor
	}
	return order.Quantity / 2 // Partial fill estimate
}

// hashCommitment creates a commitment hash for sealed bid verification
func (a *AdSlotManager) hashCommitment(price decimal.Decimal, nonce string) string {
	h := sha256.New()
	h.Write([]byte(price.String()))
	h.Write([]byte(nonce))
	return fmt.Sprintf("%x", h.Sum(nil))
}

// AdSlot represents perishable ad inventory with time-decay pricing
type AdSlot struct {
	ID               uint64                 `json:"id"`
	Publisher        string                 `json:"publisher"`
	Placement        string                 `json:"placement"`        // "ctv-preroll", "banner-300x250"
	TargetingHash    string                 `json:"targeting_hash"`   // Hash of targeting predicate
	StartTime        time.Time              `json:"start_time"`       // Delivery window start
	EndTime          time.Time              `json:"end_time"`         // Perishable expiration!
	MaxImpressions   uint64                 `json:"max_impressions"`  // Total supply
	DeliveredImprs   uint64                 `json:"delivered_imprs"`  // Already served
	MinViewability   float64                `json:"min_viewability"`  // Quality floor %
	FloorCPM         decimal.Decimal        `json:"floor_cpm"`        // Minimum price
	Active           bool                   `json:"active"`
	Targeting        TargetingPredicate     `json:"targeting"`
	SecondaryMarkets []SecondaryListing     `json:"secondary_markets,omitempty"`
}

// TargetingPredicate defines audience and quality constraints
type TargetingPredicate struct {
	GeoTargets   []string `json:"geo_targets"`   // ["US", "CA", "UK"]
	DeviceTypes  []string `json:"device_types"`  // ["CTV", "mobile", "desktop"]
	Categories   []string `json:"categories"`    // ["IAB1", "IAB2"] 
	MinAge       uint32   `json:"min_age,omitempty"`
	MaxAge       uint32   `json:"max_age,omitempty"`
	CustomFields map[string]interface{} `json:"custom_fields,omitempty"`
}

// SecondaryListing represents resale of unused ad slots
type SecondaryListing struct {
	SellerID     string          `json:"seller_id"`
	Quantity     uint64          `json:"quantity"`
	AskPrice     decimal.Decimal `json:"ask_price"`
	ListedAt     time.Time       `json:"listed_at"`
	FlashLoanOK  bool            `json:"flash_loan_ok"` // Allow flash borrows
}


// AdSlotOrder represents limit/market orders for ad slots
type AdSlotOrder struct {
	OrderID       string          `json:"order_id"`              // Primary order identifier
	ID            string          `json:"id"`                   // Alternative ID field
	TraderID      string          `json:"trader_id"`
	AdSlotID      uint64          `json:"ad_slot_id"`           // Primary slot ID
	SlotID        uint64          `json:"slot_id"`              // Alternative slot ID
	IsBuy         bool            `json:"is_buy"`
	OrderType     string          `json:"order_type"`           // "buy", "sell", "limit", "market", "commit-reveal"
	Price         decimal.Decimal `json:"price"`                // Current price
	LimitPrice    decimal.Decimal `json:"limit_price"`          // CPM in AUSD
	Quantity      uint64          `json:"quantity"`             // Number of impressions
	FilledQty     uint64          `json:"filled_qty"`           // Filled quantity
	Filled        uint64          `json:"filled"`               // Alternative filled field  
	Status        string          `json:"status"`               // "active", "filled", "canceled", "expired"
	Timestamp     time.Time       `json:"timestamp"`            // Creation timestamp
	CreatedAt     time.Time       `json:"created_at"`           // Alternative creation time
	ExpiryTime    time.Time       `json:"expiry_time"`          // When order expires
	ExpiresAt     time.Time       `json:"expires_at,omitempty"`  // Alternative expiry time
	CommitHash    string          `json:"commit_hash,omitempty"`    // For sealed bids
	RevealedPrice decimal.Decimal `json:"revealed_price,omitempty"` // After reveal
	Revealed      bool            `json:"revealed,omitempty"`
}

// RPC Methods

// CreateAdSlot - Mint new perishable ad inventory tokens
func (a *AdSlotManager) CreateAdSlot(ctx context.Context, req *CreateAdSlotRequest) (*CreateAdSlotResponse, error) {
	// Validate time window
	if req.StartTime.After(req.EndTime) {
		return nil, fmt.Errorf("invalid time window")
	}
	if req.EndTime.Before(time.Now()) {
		return nil, fmt.Errorf("window already ended")
	}
	if req.MaxImpressions == 0 {
		return nil, fmt.Errorf("no impressions")
	}

	// Generate deterministic targeting hash
	targetingHash := a.hashTargeting(req.Targeting)

	// Create slot
	slotID := a.nextID
	a.nextID++

	slot := &AdSlot{
		ID:             slotID,
		Publisher:      req.Publisher,
		Placement:      req.Placement,
		TargetingHash:  targetingHash,
		StartTime:      req.StartTime,
		EndTime:        req.EndTime,
		MaxImpressions: req.MaxImpressions,
		MinViewability: req.MinViewability,
		FloorCPM:       req.FloorCPM,
		Active:         true,
		Targeting:      req.Targeting,
	}

	// Store in state
	a.state.SetAdSlot(slot)

	// Mint SFT to publisher (using DEX engine as registry)
	if err := a.dex.MintAsset(fmt.Sprintf("adslot-%d", slotID), req.Publisher, decimal.NewFromInt(int64(req.MaxImpressions))); err != nil {
		return nil, fmt.Errorf("failed to mint SFT: %v", err)
	}

	return &CreateAdSlotResponse{
		Success: true,
		SlotID:  slotID,
		TokenID: fmt.Sprintf("adslot-%d", slotID),
	}, nil
}

// PlaceOrder - Place limit/market order for ad slots
func (a *AdSlotManager) PlaceOrder(ctx context.Context, req *PlaceOrderRequest) (*PlaceOrderResponse, error) {
	// Validate slot exists and is active
	slot, err := a.state.GetAdSlot(req.SlotID)
	if err != nil {
		return nil, fmt.Errorf("slot not found: %v", err)
	}
	if !slot.Active {
		return nil, fmt.Errorf("slot inactive")
	}
	if time.Now().After(slot.EndTime) {
		return nil, fmt.Errorf("slot expired")
	}

	// Validate order
	if req.Quantity == 0 {
		return nil, fmt.Errorf("invalid quantity")
	}

	// Check price constraints
	currentPrice := a.calculateCurrentPrice(slot)
	if req.IsBuy && req.LimitPrice.LessThan(currentPrice) {
		return nil, fmt.Errorf("bid below current price")
	}

	// Create order
	order := &AdSlotOrder{
		OrderID:    req.OrderID,
		ID:         req.OrderID,
		TraderID:   req.TraderID,
		AdSlotID:   req.SlotID,
		SlotID:     req.SlotID,
		IsBuy:      req.IsBuy,
		OrderType:  req.OrderType,
		Price:      req.LimitPrice,
		LimitPrice: req.LimitPrice,
		Quantity:   req.Quantity,
		FilledQty:  0,
		Filled:     0,
		Status:     "active",
		Timestamp:  time.Now(),
		CreatedAt:  time.Now(),
		ExpiryTime: req.ExpiresAt,
		ExpiresAt:  req.ExpiresAt,
	}

	// Handle commit-reveal orders
	if req.OrderType == "commit-reveal" {
		if req.CommitHash == "" {
			return nil, fmt.Errorf("commit hash required")
		}
		order.CommitHash = req.CommitHash
	}

	// Store order
	a.state.SetAdSlotOrder(order)

	// Add to matching engine via DEX
	dexOrder := convertToGDexOrder(order, slot)
	if err := a.dex.AddOrder(dexOrder); err != nil {
		return nil, fmt.Errorf("failed to add order: %v", err)
	}

	return &PlaceOrderResponse{
		Success:      true,
		OrderID:      req.OrderID,
		CurrentPrice: currentPrice,
		EstimatedFill: decimal.NewFromInt(int64(a.estimateOrderFill(order, slot))),
	}, nil
}

// RevealBid - Reveal sealed bid in commit-reveal auction
func (a *AdSlotManager) RevealBid(ctx context.Context, req *RevealBidRequest) (*RevealBidResponse, error) {
	order, err := a.state.GetAdSlotOrder(req.OrderID)
	if err != nil {
		return nil, fmt.Errorf("order not found: %v", err)
	}
	if order.OrderType != "commit-reveal" {
		return nil, fmt.Errorf("not a commit-reveal order")
	}
	if order.Revealed {
		return nil, fmt.Errorf("already revealed")
	}

	// Validate commitment
	expectedHash := a.hashCommitment(req.RevealedPrice, req.Nonce)
	if expectedHash != order.CommitHash {
		return nil, fmt.Errorf("invalid reveal")
	}

	// Update order with revealed price
	order.RevealedPrice = req.RevealedPrice
	order.Revealed = true
	order.LimitPrice = req.RevealedPrice // Use revealed price for matching

	a.state.SetAdSlotOrder(order)

	return &RevealBidResponse{
		Success:       true,
		RevealedPrice: req.RevealedPrice,
	}, nil
}

// CreateAdMM_Pool - Create AMM pool for continuous liquidity
func (a *AdSlotManager) CreateAdMM_Pool(ctx context.Context, req *CreateAdMM_PoolRequest) (*CreateAdMM_PoolResponse, error) {
	// Validate slot
	_, err := a.state.GetAdSlot(req.SlotID)
	if err != nil {
		return nil, fmt.Errorf("slot not found: %v", err)
	}

	// Check for existing pool
	if _, exists := a.state.GetAdMM_Pool(req.SlotID); exists {
		return nil, fmt.Errorf("pool already exists")
	}

	// Calculate initial price and LP tokens
	initialPrice := req.InitialAUSD.Div(decimal.NewFromInt(int64(req.InitialSlots)))
	// Calculate LP tokens as geometric mean approximation
	lpTokensValue := req.InitialAUSD.Mul(decimal.NewFromInt(int64(req.InitialSlots)))
	lpTokens := decimal.NewFromFloat(math.Sqrt(lpTokensValue.InexactFloat64()))

	pool := &AdMM_Pool{
		SlotID:        req.SlotID,
		ReserveAUSD:   req.InitialAUSD,
		ReserveSlots:  req.InitialSlots,
		LastPrice:     initialPrice,
		LPTokenSupply: lpTokens,
		TimeDecayRate: req.TimeDecayRate,
		CreatedAt:     time.Now(),
	}

	a.state.SetAdMM_Pool(req.SlotID, pool)

	// Transfer initial liquidity would happen here
	// Note: transferAUSD method needs to be implemented
	// if err := a.transferAUSD(req.LiquidityProvider, "pool", req.InitialAUSD); err != nil {
	//     return nil, fmt.Errorf("AUSD transfer failed: %v", err)
	// }

	return &CreateAdMM_PoolResponse{
		Success:      true,
		PoolID:       fmt.Sprintf("%d", req.SlotID),
		LPTokens:     lpTokens,
		InitialPrice: initialPrice,
	}, nil
}

// SwapAdMM - Execute AMM swap (continuous liquidity)
func (a *AdSlotManager) SwapAdMM(ctx context.Context, req *SwapAdMM_Request) (*SwapAdMM_Response, error) {
	pool, exists := a.state.GetAdMM_Pool(req.SlotID)
	if !exists {
		return nil, fmt.Errorf("pool not found")
	}

	slot, _ := a.state.GetAdSlot(req.SlotID)
	
	// Calculate swap with time decay
	swapAmount := a.calculateAMM_Swap(pool, slot, uint64(req.AmountIn.IntPart()), req.BuyAUSD)
	if swapAmount.LessThanOrEqual(decimal.Zero) {
		return nil, fmt.Errorf("insufficient liquidity")
	}

	// Execute swap
	if req.BuyAUSD {
		// Selling slots for AUSD
		pool.ReserveSlots += uint64(req.AmountIn.IntPart())
		pool.ReserveAUSD = pool.ReserveAUSD.Sub(swapAmount)
	} else {
		// Buying slots with AUSD
		pool.ReserveAUSD = pool.ReserveAUSD.Add(req.AmountIn)
		pool.ReserveSlots -= uint64(swapAmount.IntPart())
	}

	// Update pool price
	if pool.ReserveSlots > 0 {
		pool.LastPrice = pool.ReserveAUSD.Div(decimal.NewFromInt(int64(pool.ReserveSlots)))
	}

	a.state.SetAdMM_Pool(req.SlotID, pool)

	return &SwapAdMM_Response{
		Success:    true,
		AmountOut:  swapAmount,
		NewPrice:       pool.LastPrice,
		SlippageActual: a.calculateSlippage(req.ExpectedAmountOut, swapAmount),
	}, nil
}

// RecordDelivery - Record impression delivery (burns tokens)
func (a *AdSlotManager) RecordDelivery(ctx context.Context, req *RecordDeliveryRequest) (*RecordDeliveryResponse, error) {
	slot, err := a.state.GetAdSlot(req.AdSlotID)
	if err != nil {
		return nil, fmt.Errorf("slot not found: %v", err)
	}

	// Validate delivery window
	now := time.Now()
	if now.Before(slot.StartTime) || now.After(slot.EndTime) {
		return nil, fmt.Errorf("outside delivery window")
	}

	// Check capacity
	if slot.DeliveredImprs+req.Count > slot.MaxImpressions {
		return nil, fmt.Errorf("exceeds capacity")
	}

	// Record delivery
	slot.DeliveredImprs += req.Count
	a.state.SetAdSlot(slot)

	// Burn delivered tokens from circulation
	if err := a.dex.BurnAsset(fmt.Sprintf("adslot-%d", req.SlotID), slot.Publisher, decimal.NewFromInt(int64(req.Count))); err != nil {
		return nil, fmt.Errorf("failed to burn tokens: %v", err)
	}

	return &RecordDeliveryResponse{
		Success:           true,
		DeliveredCount:    req.Count,
		TotalDelivered:    slot.DeliveredImprs,
		RemainingSupply:   slot.MaxImpressions - slot.DeliveredImprs,
	}, nil
}

// Helper functions

func (a *AdSlotManager) hashTargeting(targeting TargetingPredicate) string {
	h := sha256.New()
	
	// Hash all targeting fields deterministically
	for _, geo := range targeting.GeoTargets {
		h.Write([]byte(geo))
	}
	for _, device := range targeting.DeviceTypes {
		h.Write([]byte(device))
	}
	for _, cat := range targeting.Categories {
		h.Write([]byte(cat))
	}
	
	// Add age constraints
	ageBuf := make([]byte, 8)
	binary.LittleEndian.PutUint32(ageBuf[0:4], targeting.MinAge)
	binary.LittleEndian.PutUint32(ageBuf[4:8], targeting.MaxAge)
	h.Write(ageBuf)
	
	return fmt.Sprintf("%x", h.Sum(nil))
}

func (a *AdSlotManager) calculateCurrentPrice(slot *AdSlot) decimal.Decimal {
	now := time.Now()
	
	// Expired = worthless
	if now.After(slot.EndTime) || !slot.Active {
		return decimal.Zero
	}
	
	// Not started = full price
	if now.Before(slot.StartTime) {
		return slot.FloorCPM
	}
	
	// Linear time decay
	timeRemaining := slot.EndTime.Sub(now).Seconds()
	totalWindow := slot.EndTime.Sub(slot.StartTime).Seconds()
	
	if totalWindow <= 0 {
		return slot.FloorCPM
	}
	
	// Price = floor + (50% premium * time_remaining / total_window)
	premium := slot.FloorCPM.Div(decimal.NewFromInt(2))
	timeRatio := decimal.NewFromFloat(timeRemaining / totalWindow)
	
	return slot.FloorCPM.Add(premium.Mul(timeRatio))
}

func (a *AdSlotManager) calculateAMM_Swap(pool *AdMM_Pool, slot *AdSlot, amountIn uint64, buyAUSD bool) decimal.Decimal {
	// Constant product AMM with time decay: k = reserves_ausd * reserves_slots * time_factor
	if pool.ReserveAUSD.LessThanOrEqual(decimal.Zero) || pool.ReserveSlots == 0 {
		return decimal.Zero
	}
	
	// Apply time decay to effective reserves
	timeDecay := a.calculateTimeDecay(slot, pool.TimeDecayRate)
	effectiveK := pool.ReserveAUSD.Mul(decimal.NewFromInt(int64(pool.ReserveSlots))).Mul(timeDecay)
	
	if buyAUSD {
		// Selling slots for AUSD: new_slots = old_slots + amount_in
		newSlots := decimal.NewFromInt(int64(pool.ReserveSlots + amountIn))
		newAUSD := effectiveK.Div(newSlots)
		return pool.ReserveAUSD.Sub(newAUSD)
	} else {
		// Buying slots with AUSD: new_ausd = old_ausd + amount_in
		newAUSD := pool.ReserveAUSD.Add(decimal.NewFromInt(int64(amountIn)))
		newSlots := effectiveK.Div(newAUSD)
		return decimal.NewFromInt(int64(pool.ReserveSlots)).Sub(newSlots)
	}
}

func (a *AdSlotManager) calculateTimeDecay(slot *AdSlot, decayRate decimal.Decimal) decimal.Decimal {
	now := time.Now()
	if now.After(slot.EndTime) {
		return decimal.Zero
	}
	
	timeRemaining := slot.EndTime.Sub(now).Seconds()
	totalWindow := slot.EndTime.Sub(slot.StartTime).Seconds()
	
	if totalWindow <= 0 {
		return decimal.NewFromInt(1)
	}
	
	// Exponential decay: e^(-λ * (1 - time_remaining/total_window))
	normalizedTime := decimal.NewFromFloat(1.0 - (timeRemaining / totalWindow))
	exponent := decayRate.Mul(normalizedTime).Neg()
	
	// Approximate e^x for small x
	return decimal.NewFromInt(1).Add(exponent)
}

// calculateSlippage calculates actual vs expected slippage
func (a *AdSlotManager) calculateSlippage(expected, actual decimal.Decimal) decimal.Decimal {
	if expected.IsZero() {
		return decimal.Zero
	}
	
	diff := actual.Sub(expected).Abs()
	return diff.Div(expected).Mul(decimal.NewFromInt(100)) // Return as percentage
}

// Request/Response types

type CreateAdSlotRequest struct {
	Publisher      string             `json:"publisher"`
	Placement      string             `json:"placement"`
	Targeting      TargetingPredicate `json:"targeting"`
	StartTime      time.Time          `json:"start_time"`
	EndTime        time.Time          `json:"end_time"`
	MaxImpressions uint64             `json:"max_impressions"`
	MinViewability float64            `json:"min_viewability"`
	FloorCPM       decimal.Decimal    `json:"floor_cpm"`
}

type CreateAdSlotResponse struct {
	Success bool   `json:"success"`
	SlotID  uint64 `json:"slot_id"`
	TokenID string `json:"token_id"`
}

type PlaceOrderRequest struct {
	OrderID    string          `json:"order_id"`
	TraderID   string          `json:"trader_id"`
	SlotID     uint64          `json:"slot_id"`
	IsBuy      bool            `json:"is_buy"`
	OrderType  string          `json:"order_type"`
	LimitPrice decimal.Decimal `json:"limit_price"`
	Quantity   uint64          `json:"quantity"`
	ExpiresAt  time.Time       `json:"expires_at,omitempty"`
	CommitHash string          `json:"commit_hash,omitempty"`
}

type PlaceOrderResponse struct {
	Success       bool            `json:"success"`
	OrderID       string          `json:"order_id"`
	CurrentPrice  decimal.Decimal `json:"current_price"`
	EstimatedFill decimal.Decimal `json:"estimated_fill"`
}

// Additional request/response types would follow similar patterns...