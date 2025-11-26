package polymarket

import (
	"fmt"

	"github.com/shopspring/decimal"

	"github.com/ivanzzeth/polymarket-go-clob-client/types"
)

// ConvertToComplementaryOrder converts a limit order to its complementary token order
// Based on Polymarket's complementary token mechanism:
//   - Buy token @ P  → Sell complementary @ (1-P)
//   - Sell token @ P → Buy complementary @ (1-P)
//
// This allows traders to achieve the same position using whichever side has better liquidity
// For example: Buy YES @ 0.6 = Sell NO @ 0.4
func ConvertToComplementaryOrder(order *types.UserOrder, complementaryTokenID string) (*types.UserOrder, error) {
	if order == nil {
		return nil, fmt.Errorf("order cannot be nil")
	}

	// Convert side: BUY ↔ SELL
	var convertedSide types.OrderSide
	if order.Side == types.OrderSideBuy {
		convertedSide = types.OrderSideSell
	} else if order.Side == types.OrderSideSell {
		convertedSide = types.OrderSideBuy
	} else {
		return nil, fmt.Errorf("invalid order side: %s", order.Side)
	}

	// Convert price: P → (1 - P)
	one := decimal.NewFromInt(1)
	convertedPrice := one.Sub(order.Price)

	// Validate converted price is in valid range [0, 1]
	if convertedPrice.LessThan(decimal.Zero) || convertedPrice.GreaterThan(one) {
		return nil, fmt.Errorf("converted price %s is out of valid range [0, 1]", convertedPrice)
	}

	// Create converted order with same parameters except side, token, and price
	converted := &types.UserOrder{
		TokenID:    complementaryTokenID,
		Price:      convertedPrice,
		Size:       order.Size,
		Side:       convertedSide,
		FeeRateBps: order.FeeRateBps,
		Nonce:      order.Nonce,
		Expiration: order.Expiration,
		Taker:      order.Taker,
	}

	return converted, nil
}

// ConvertMarketOrderToComplementary converts a market order to its complementary token order
// It works by: market order → limit order → convert → limit order → market order
func ConvertMarketOrderToComplementary(order *types.UserMarketOrder, complementaryTokenID string) (*types.UserMarketOrder, error) {
	if order == nil {
		return nil, fmt.Errorf("order cannot be nil")
	}

	// Convert to limit order
	limitOrder, err := marketOrderToLimitOrder(order)
	if err != nil {
		return nil, fmt.Errorf("failed to convert market order to limit order: %w", err)
	}

	// Convert limit order
	convertedLimit, err := ConvertToComplementaryOrder(limitOrder, complementaryTokenID)
	if err != nil {
		return nil, fmt.Errorf("failed to convert limit order: %w", err)
	}

	// Convert back to market order
	convertedMarket, err := limitOrderToMarketOrder(convertedLimit, order.Amount)
	if err != nil {
		return nil, fmt.Errorf("failed to convert limit order to market order: %w", err)
	}

	return convertedMarket, nil
}

// marketOrderToLimitOrder converts a market order to a limit order
func marketOrderToLimitOrder(order *types.UserMarketOrder) (*types.UserOrder, error) {
	if order.Price == nil {
		return nil, fmt.Errorf("market order must have price set for conversion")
	}

	limitOrder := &types.UserOrder{
		TokenID: order.TokenID,
		Price:   *order.Price,
		Size:    order.Amount,
		Side:    order.Side,
		Nonce:   order.Nonce,
		Taker:   order.Taker,
	}

	// Convert FeeRateBps if present
	if order.FeeRateBps != nil {
		limitOrder.FeeRateBps = *order.FeeRateBps
	}

	return limitOrder, nil
}

// limitOrderToMarketOrder converts a limit order back to a market order
func limitOrderToMarketOrder(order *types.UserOrder, amount decimal.Decimal) (*types.UserMarketOrder, error) {
	marketOrder := &types.UserMarketOrder{
		TokenID: order.TokenID,
		Price:   &order.Price,
		Amount:  amount,
		Side:    order.Side,
		Nonce:   order.Nonce,
		Taker:   order.Taker,
	}

	// Convert FeeRateBps to pointer
	if order.FeeRateBps != 0 {
		marketOrder.FeeRateBps = &order.FeeRateBps
	}

	return marketOrder, nil
}

// ConvertToOppositeSideOrder converts a limit order to the opposite side (BUY ↔ SELL) with optional spread.
// The spread parameter adjusts the price to create a market making spread:
//   - BUY → SELL: price becomes P + spread (sell at higher price)
//   - SELL → BUY: price becomes P - spread (buy at lower price)
//
// Examples without spread (spread = 0):
//   - Buy YES @ 0.49  → Sell YES @ 0.49
//   - Sell NO @ 0.51  → Buy NO @ 0.51
//
// Examples with spread (spread = 0.02):
//   - Buy YES @ 0.49  → Sell YES @ 0.51 (0.49 + 0.02)
//   - Sell NO @ 0.51  → Buy NO @ 0.49 (0.51 - 0.02)
//
// When combined with ConvertToComplementaryOrder, you can create market making strategies:
//   Buy YES @ 0.49 → (opposite + spread) → Sell YES @ 0.51 → (complementary) → Buy NO @ 0.49
//   Result: Buy YES @ 0.49 + Buy NO @ 0.49 = 0.98 cost, merge to 1.0, profit = 0.02
//
// This is useful for:
//   - Market making with spreads
//   - Creating arbitrage opportunities
//   - Testing order matching with spreads
//   - Flexible strategy composition
func ConvertToOppositeSideOrder(order *types.UserOrder, spread decimal.Decimal) (*types.UserOrder, error) {
	if order == nil {
		return nil, fmt.Errorf("order cannot be nil")
	}

	// Validate spread is non-negative
	if spread.LessThan(decimal.Zero) {
		return nil, fmt.Errorf("spread must be non-negative, got %s", spread)
	}

	// Convert side: BUY ↔ SELL
	var convertedSide types.OrderSide
	var convertedPrice decimal.Decimal

	if order.Side == types.OrderSideBuy {
		// BUY → SELL: increase price (sell at higher price)
		convertedSide = types.OrderSideSell
		convertedPrice = order.Price.Add(spread)
	} else if order.Side == types.OrderSideSell {
		// SELL → BUY: decrease price (buy at lower price)
		convertedSide = types.OrderSideBuy
		convertedPrice = order.Price.Sub(spread)
	} else {
		return nil, fmt.Errorf("invalid order side: %s", order.Side)
	}

	// Validate converted price is in valid range [0, 1]
	one := decimal.NewFromInt(1)
	if convertedPrice.LessThan(decimal.Zero) || convertedPrice.GreaterThan(one) {
		return nil, fmt.Errorf("converted price %s is out of valid range [0, 1]", convertedPrice)
	}

	// Create converted order with same parameters except side and price
	converted := &types.UserOrder{
		TokenID:    order.TokenID,
		Price:      convertedPrice,
		Size:       order.Size,
		Side:       convertedSide,
		FeeRateBps: order.FeeRateBps,
		Nonce:      order.Nonce,
		Expiration: order.Expiration,
		Taker:      order.Taker,
	}

	return converted, nil
}

// ConvertMarketOrderToOppositeSide converts a market order to the opposite side (BUY ↔ SELL) with optional spread.
// The spread parameter adjusts the price to create a market making spread:
//   - BUY → SELL: price becomes P + spread (sell at higher price)
//   - SELL → BUY: price becomes P - spread (buy at lower price)
func ConvertMarketOrderToOppositeSide(order *types.UserMarketOrder, spread decimal.Decimal) (*types.UserMarketOrder, error) {
	if order == nil {
		return nil, fmt.Errorf("order cannot be nil")
	}

	// Validate spread is non-negative
	if spread.LessThan(decimal.Zero) {
		return nil, fmt.Errorf("spread must be non-negative, got %s", spread)
	}

	// Convert side: BUY ↔ SELL
	var convertedSide types.OrderSide
	var convertedPrice *decimal.Decimal

	if order.Price != nil {
		var price decimal.Decimal
		if order.Side == types.OrderSideBuy {
			// BUY → SELL: increase price (sell at higher price)
			convertedSide = types.OrderSideSell
			price = order.Price.Add(spread)
		} else if order.Side == types.OrderSideSell {
			// SELL → BUY: decrease price (buy at lower price)
			convertedSide = types.OrderSideBuy
			price = order.Price.Sub(spread)
		} else {
			return nil, fmt.Errorf("invalid order side: %s", order.Side)
		}

		// Validate converted price is in valid range [0, 1]
		one := decimal.NewFromInt(1)
		if price.LessThan(decimal.Zero) || price.GreaterThan(one) {
			return nil, fmt.Errorf("converted price %s is out of valid range [0, 1]", price)
		}

		convertedPrice = &price
	} else {
		// No price specified, just flip the side
		if order.Side == types.OrderSideBuy {
			convertedSide = types.OrderSideSell
		} else if order.Side == types.OrderSideSell {
			convertedSide = types.OrderSideBuy
		} else {
			return nil, fmt.Errorf("invalid order side: %s", order.Side)
		}
	}

	// Create converted order with same parameters except side and price
	converted := &types.UserMarketOrder{
		TokenID:    order.TokenID,
		Price:      convertedPrice,
		Amount:     order.Amount,
		Side:       convertedSide,
		FeeRateBps: order.FeeRateBps,
		Nonce:      order.Nonce,
		Taker:      order.Taker,
	}

	return converted, nil
}

// ConvertToMatchingSameSideOrder converts an order to a matching same-side order on the complementary token with optional spread.
// This is a convenience function that combines opposite side + complementary conversions.
//
// The conversion keeps the order side (BUY/SELL) the same while switching to the complementary token:
//   - Buy YES @ P  → Buy NO @ (1-P) - spread
//   - Sell YES @ P → Sell NO @ (1-P) + spread
//   - Buy NO @ P   → Buy YES @ (1-P) - spread
//   - Sell NO @ P  → Sell YES @ (1-P) + spread
//
// The spread parameter creates market making opportunities:
//   - For BUY orders: reduces the matching order price, creating profit when both filled
//   - For SELL orders: increases the matching order price, creating profit when both filled
//
// Example use case - Generate matching BUY orders without spread:
//
//	original := Buy YES @ 0.49
//	matching, _ := ConvertToMatchingSameSideOrder(original, complementaryTokenID, decimal.Zero)
//	// Result: matching = Buy NO @ 0.51
//	// Cost: 0.49 + 0.51 = 1.0, merge to 1.0, profit = 0
//
// Example use case - Generate matching BUY orders WITH spread (market making):
//
//	original := Buy YES @ 0.49
//	spread := decimal.NewFromFloat(0.02)
//	matching, _ := ConvertToMatchingSameSideOrder(original, complementaryTokenID, spread)
//	// Result: matching = Buy NO @ 0.49  (0.51 - 0.02)
//	// Cost: 0.49 + 0.49 = 0.98, merge to 1.0, profit = 0.02 ✅
//
// Example use case - Generate matching SELL orders WITH spread:
//
//	original := Sell NO @ 0.45
//	spread := decimal.NewFromFloat(0.02)
//	matching, _ := ConvertToMatchingSameSideOrder(original, complementaryTokenID, spread)
//	// Result: matching = Sell YES @ 0.57  (0.55 + 0.02)
//	// Revenue: 0.45 + 0.57 = 1.02, split cost 1.0, profit = 0.02 ✅
//
// This creates two orders on the same side that can match each other, triggering CTF operations:
//   - Two BUY orders trigger the split operation (mint YES + NO from collateral)
//   - Two SELL orders trigger the merge operation (burn YES + NO to collateral)
//
// This is useful for:
//   - Market making with guaranteed profit (when both orders fill)
//   - Liquidity provision while earning spreads
//   - Testing CTF split/merge operations
//   - Arbitrage strategies utilizing CTF operations
func ConvertToMatchingSameSideOrder(order *types.UserOrder, complementaryTokenID string, spread decimal.Decimal) (*types.UserOrder, error) {
	if order == nil {
		return nil, fmt.Errorf("order cannot be nil")
	}

	// Step 1: Convert to opposite side with spread (BUY ↔ SELL)
	opposite, err := ConvertToOppositeSideOrder(order, spread)
	if err != nil {
		return nil, fmt.Errorf("failed to convert to opposite side: %w", err)
	}

	// Step 2: Convert to complementary token (also flips side back)
	matching, err := ConvertToComplementaryOrder(opposite, complementaryTokenID)
	if err != nil {
		return nil, fmt.Errorf("failed to convert to complementary: %w", err)
	}

	return matching, nil
}

// ConvertMarketOrderToMatchingSameSide converts a market order to a matching same-side order on the complementary token with optional spread.
// This is the market order version of ConvertToMatchingSameSideOrder.
func ConvertMarketOrderToMatchingSameSide(order *types.UserMarketOrder, complementaryTokenID string, spread decimal.Decimal) (*types.UserMarketOrder, error) {
	if order == nil {
		return nil, fmt.Errorf("order cannot be nil")
	}

	// Step 1: Convert to opposite side with spread (BUY ↔ SELL)
	opposite, err := ConvertMarketOrderToOppositeSide(order, spread)
	if err != nil {
		return nil, fmt.Errorf("failed to convert to opposite side: %w", err)
	}

	// Step 2: Convert to complementary token (also flips side back)
	matching, err := ConvertMarketOrderToComplementary(opposite, complementaryTokenID)
	if err != nil {
		return nil, fmt.Errorf("failed to convert to complementary: %w", err)
	}

	return matching, nil
}
