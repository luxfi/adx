package dex

import (
	"fmt"
	"github.com/shopspring/decimal"
)

// Engine represents a minimal DEX engine for asset transfers
type Engine struct {
	balances map[string]map[string]decimal.Decimal // assetID -> account -> balance
}

// NewEngine creates a new DEX engine
func NewEngine() *Engine {
	return &Engine{
		balances: make(map[string]map[string]decimal.Decimal),
	}
}

// TransferAsset transfers an asset between accounts
func (e *Engine) TransferAsset(assetID, from, to string, amount decimal.Decimal) error {
	if amount.LessThanOrEqual(decimal.Zero) {
		return fmt.Errorf("amount must be positive")
	}

	// Initialize asset maps if needed
	if e.balances[assetID] == nil {
		e.balances[assetID] = make(map[string]decimal.Decimal)
	}

	// Check from balance
	fromBalance, exists := e.balances[assetID][from]
	if !exists || fromBalance.LessThan(amount) {
		return fmt.Errorf("insufficient balance")
	}

	// Perform transfer
	e.balances[assetID][from] = fromBalance.Sub(amount)
	
	toBalance := e.balances[assetID][to]
	e.balances[assetID][to] = toBalance.Add(amount)

	return nil
}

// GetBalance returns the balance for an account and asset
func (e *Engine) GetBalance(assetID, account string) decimal.Decimal {
	if e.balances[assetID] == nil {
		return decimal.Zero
	}
	return e.balances[assetID][account]
}

// SetBalance sets the balance for an account and asset (for testing/initialization)
func (e *Engine) SetBalance(assetID, account string, amount decimal.Decimal) {
	if e.balances[assetID] == nil {
		e.balances[assetID] = make(map[string]decimal.Decimal)
	}
	e.balances[assetID][account] = amount
}

// MintAsset creates new tokens for an account
func (e *Engine) MintAsset(assetID, account string, amount decimal.Decimal) error {
	if amount.LessThanOrEqual(decimal.Zero) {
		return fmt.Errorf("amount must be positive")
	}
	
	if e.balances[assetID] == nil {
		e.balances[assetID] = make(map[string]decimal.Decimal)
	}
	
	currentBalance := e.balances[assetID][account]
	e.balances[assetID][account] = currentBalance.Add(amount)
	
	return nil
}

// Order represents a trade order
type Order struct {
	OrderID  string
	AssetID  string
	Price    decimal.Decimal
	Quantity decimal.Decimal
	IsBuy    bool
}

// AddOrder adds a new order to the engine
func (e *Engine) AddOrder(order *Order) error {
	// Simplified order management - in production would have proper order book
	return nil
}

// BurnAsset removes tokens from an account
func (e *Engine) BurnAsset(assetID, account string, amount decimal.Decimal) error {
	if amount.LessThanOrEqual(decimal.Zero) {
		return fmt.Errorf("amount must be positive")
	}
	
	if e.balances[assetID] == nil || e.balances[assetID][account].LessThan(amount) {
		return fmt.Errorf("insufficient balance")
	}
	
	e.balances[assetID][account] = e.balances[assetID][account].Sub(amount)
	
	return nil
}