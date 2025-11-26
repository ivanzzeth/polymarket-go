package polymarket

import (
	"math/big"
	"testing"

	"github.com/shopspring/decimal"

	"github.com/ivanzzeth/polymarket-go-clob-client/types"
)

func TestConvertOrder(t *testing.T) {
	tests := []struct {
		name                 string
		order                *types.UserOrder
		complementaryTokenID string
		wantErr              bool
		validateResult       func(*testing.T, *types.UserOrder)
	}{
		{
			name: "Buy YES @ 0.6 converts to Sell NO @ 0.4",
			order: &types.UserOrder{
				TokenID:    "token-yes",
				Price:      decimal.NewFromFloat(0.6),
				Size:       decimal.NewFromInt(100),
				Side:       types.OrderSideBuy,
				FeeRateBps: 10,
			},
			complementaryTokenID: "token-no",
			wantErr:              false,
			validateResult: func(t *testing.T, result *types.UserOrder) {
				if result.TokenID != "token-no" {
					t.Errorf("TokenID = %v, want token-no", result.TokenID)
				}
				expectedPrice := decimal.NewFromFloat(0.4)
				if !result.Price.Equal(expectedPrice) {
					t.Errorf("Price = %v, want %v", result.Price, expectedPrice)
				}
				if result.Side != types.OrderSideSell {
					t.Errorf("Side = %v, want %v", result.Side, types.OrderSideSell)
				}
				if !result.Size.Equal(decimal.NewFromInt(100)) {
					t.Errorf("Size = %v, want 100", result.Size)
				}
				if result.FeeRateBps != 10 {
					t.Errorf("FeeRateBps = %v, want 10", result.FeeRateBps)
				}
			},
		},
		{
			name: "Sell YES @ 0.3 converts to Buy NO @ 0.7",
			order: &types.UserOrder{
				TokenID:    "token-yes",
				Price:      decimal.NewFromFloat(0.3),
				Size:       decimal.NewFromInt(50),
				Side:       types.OrderSideSell,
				FeeRateBps: 5,
			},
			complementaryTokenID: "token-no",
			wantErr:              false,
			validateResult: func(t *testing.T, result *types.UserOrder) {
				if result.TokenID != "token-no" {
					t.Errorf("TokenID = %v, want token-no", result.TokenID)
				}
				expectedPrice := decimal.NewFromFloat(0.7)
				if !result.Price.Equal(expectedPrice) {
					t.Errorf("Price = %v, want %v", result.Price, expectedPrice)
				}
				if result.Side != types.OrderSideBuy {
					t.Errorf("Side = %v, want %v", result.Side, types.OrderSideBuy)
				}
			},
		},
		{
			name: "Buy NO @ 0.45 converts to Sell YES @ 0.55",
			order: &types.UserOrder{
				TokenID:    "token-no",
				Price:      decimal.NewFromFloat(0.45),
				Size:       decimal.NewFromInt(200),
				Side:       types.OrderSideBuy,
				FeeRateBps: 0,
			},
			complementaryTokenID: "token-yes",
			wantErr:              false,
			validateResult: func(t *testing.T, result *types.UserOrder) {
				if result.TokenID != "token-yes" {
					t.Errorf("TokenID = %v, want token-yes", result.TokenID)
				}
				expectedPrice := decimal.NewFromFloat(0.55)
				if !result.Price.Equal(expectedPrice) {
					t.Errorf("Price = %v, want %v", result.Price, expectedPrice)
				}
				if result.Side != types.OrderSideSell {
					t.Errorf("Side = %v, want %v", result.Side, types.OrderSideSell)
				}
			},
		},
		{
			name: "Sell NO @ 0.8 converts to Buy YES @ 0.2",
			order: &types.UserOrder{
				TokenID:    "token-no",
				Price:      decimal.NewFromFloat(0.8),
				Size:       decimal.NewFromInt(75),
				Side:       types.OrderSideSell,
				FeeRateBps: 15,
			},
			complementaryTokenID: "token-yes",
			wantErr:              false,
			validateResult: func(t *testing.T, result *types.UserOrder) {
				if result.TokenID != "token-yes" {
					t.Errorf("TokenID = %v, want token-yes", result.TokenID)
				}
				expectedPrice := decimal.NewFromFloat(0.2)
				if !result.Price.Equal(expectedPrice) {
					t.Errorf("Price = %v, want %v", result.Price, expectedPrice)
				}
				if result.Side != types.OrderSideBuy {
					t.Errorf("Side = %v, want %v", result.Side, types.OrderSideBuy)
				}
			},
		},
		{
			name: "Edge case: Price 0 converts to Price 1",
			order: &types.UserOrder{
				TokenID: "token-yes",
				Price:   decimal.Zero,
				Size:    decimal.NewFromInt(10),
				Side:    types.OrderSideBuy,
			},
			complementaryTokenID: "token-no",
			wantErr:              false,
			validateResult: func(t *testing.T, result *types.UserOrder) {
				expectedPrice := decimal.NewFromInt(1)
				if !result.Price.Equal(expectedPrice) {
					t.Errorf("Price = %v, want %v", result.Price, expectedPrice)
				}
			},
		},
		{
			name: "Edge case: Price 1 converts to Price 0",
			order: &types.UserOrder{
				TokenID: "token-yes",
				Price:   decimal.NewFromInt(1),
				Size:    decimal.NewFromInt(10),
				Side:    types.OrderSideBuy,
			},
			complementaryTokenID: "token-no",
			wantErr:              false,
			validateResult: func(t *testing.T, result *types.UserOrder) {
				if !result.Price.Equal(decimal.Zero) {
					t.Errorf("Price = %v, want 0", result.Price)
				}
			},
		},
		{
			name: "Preserves optional fields: Nonce, Expiration, Taker",
			order: &types.UserOrder{
				TokenID:    "token-yes",
				Price:      decimal.NewFromFloat(0.5),
				Size:       decimal.NewFromInt(100),
				Side:       types.OrderSideBuy,
				FeeRateBps: 10,
				Nonce:      big.NewInt(12345),
				Expiration: func() *int64 { v := int64(1234567890); return &v }(),
				Taker:      "0xABCDEF1234567890",
			},
			complementaryTokenID: "token-no",
			wantErr:              false,
			validateResult: func(t *testing.T, result *types.UserOrder) {
				if result.Nonce.Cmp(big.NewInt(12345)) != 0 {
					t.Errorf("Nonce = %v, want 12345", result.Nonce)
				}
				if result.Expiration == nil || *result.Expiration != 1234567890 {
					t.Errorf("Expiration = %v, want 1234567890", result.Expiration)
				}
				if result.Taker != "0xABCDEF1234567890" {
					t.Errorf("Taker = %v, want 0xABCDEF1234567890", result.Taker)
				}
			},
		},
		{
			name:                 "Error: nil order",
			order:                nil,
			complementaryTokenID: "token-no",
			wantErr:              true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ConvertToComplementaryOrder(tt.order, tt.complementaryTokenID)

			if (err != nil) != tt.wantErr {
				t.Errorf("ConvertToComplementaryOrder() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err == nil && tt.validateResult != nil {
				tt.validateResult(t, result)
			}
		})
	}
}

func TestConvertOrderPriceFormula(t *testing.T) {
	// Test the mathematical relationship: P_converted = 1 - P_original
	testPrices := []float64{0.0, 0.1, 0.25, 0.5, 0.75, 0.9, 1.0}

	for _, price := range testPrices {
		t.Run(decimal.NewFromFloat(price).String(), func(t *testing.T) {
			order := &types.UserOrder{
				TokenID: "token-a",
				Price:   decimal.NewFromFloat(price),
				Size:    decimal.NewFromInt(1),
				Side:    types.OrderSideBuy,
			}

			result, err := ConvertToComplementaryOrder(order, "token-b")
			if err != nil {
				t.Fatalf("ConvertToComplementaryOrder() error = %v", err)
			}

			expected := decimal.NewFromInt(1).Sub(decimal.NewFromFloat(price))
			if !result.Price.Equal(expected) {
				t.Errorf("ConvertToComplementaryOrder() price = %v, want %v (1 - %v)", result.Price, expected, price)
			}
		})
	}
}

func TestConvertOrderSideConversion(t *testing.T) {
	tests := []struct {
		originalSide types.OrderSide
		expectedSide types.OrderSide
	}{
		{types.OrderSideBuy, types.OrderSideSell},
		{types.OrderSideSell, types.OrderSideBuy},
	}

	for _, tt := range tests {
		t.Run(string(tt.originalSide), func(t *testing.T) {
			order := &types.UserOrder{
				TokenID: "token-a",
				Price:   decimal.NewFromFloat(0.5),
				Size:    decimal.NewFromInt(1),
				Side:    tt.originalSide,
			}

			result, err := ConvertToComplementaryOrder(order, "token-b")
			if err != nil {
				t.Fatalf("ConvertToComplementaryOrder() error = %v", err)
			}

			if result.Side != tt.expectedSide {
				t.Errorf("ConvertToComplementaryOrder() side = %v, want %v", result.Side, tt.expectedSide)
			}
		})
	}
}

func TestConvertMarketOrder(t *testing.T) {
	tests := []struct {
		name                 string
		order                *types.UserMarketOrder
		complementaryTokenID string
		wantErr              bool
		validateResult       func(*testing.T, *types.UserMarketOrder)
	}{
		{
			name: "Buy market order YES @ 0.6 converts to Sell market order NO @ 0.4",
			order: &types.UserMarketOrder{
				TokenID: "token-yes",
				Price:   func() *decimal.Decimal { p := decimal.NewFromFloat(0.6); return &p }(),
				Amount:  decimal.NewFromInt(100),
				Side:    types.OrderSideBuy,
			},
			complementaryTokenID: "token-no",
			wantErr:              false,
			validateResult: func(t *testing.T, result *types.UserMarketOrder) {
				if result.TokenID != "token-no" {
					t.Errorf("TokenID = %v, want token-no", result.TokenID)
				}
				expectedPrice := decimal.NewFromFloat(0.4)
				if result.Price == nil || !result.Price.Equal(expectedPrice) {
					t.Errorf("Price = %v, want %v", result.Price, expectedPrice)
				}
				if result.Side != types.OrderSideSell {
					t.Errorf("Side = %v, want %v", result.Side, types.OrderSideSell)
				}
				if !result.Amount.Equal(decimal.NewFromInt(100)) {
					t.Errorf("Amount = %v, want 100", result.Amount)
				}
			},
		},
		{
			name: "Sell market order YES @ 0.3 converts to Buy market order NO @ 0.7",
			order: &types.UserMarketOrder{
				TokenID: "token-yes",
				Price:   func() *decimal.Decimal { p := decimal.NewFromFloat(0.3); return &p }(),
				Amount:  decimal.NewFromInt(50),
				Side:    types.OrderSideSell,
			},
			complementaryTokenID: "token-no",
			wantErr:              false,
			validateResult: func(t *testing.T, result *types.UserMarketOrder) {
				if result.TokenID != "token-no" {
					t.Errorf("TokenID = %v, want token-no", result.TokenID)
				}
				expectedPrice := decimal.NewFromFloat(0.7)
				if result.Price == nil || !result.Price.Equal(expectedPrice) {
					t.Errorf("Price = %v, want %v", result.Price, expectedPrice)
				}
				if result.Side != types.OrderSideBuy {
					t.Errorf("Side = %v, want %v", result.Side, types.OrderSideBuy)
				}
			},
		},
		{
			name: "Buy market order NO @ 0.45 converts to Sell market order YES @ 0.55",
			order: &types.UserMarketOrder{
				TokenID: "token-no",
				Price:   func() *decimal.Decimal { p := decimal.NewFromFloat(0.45); return &p }(),
				Amount:  decimal.NewFromInt(200),
				Side:    types.OrderSideBuy,
			},
			complementaryTokenID: "token-yes",
			wantErr:              false,
			validateResult: func(t *testing.T, result *types.UserMarketOrder) {
				if result.TokenID != "token-yes" {
					t.Errorf("TokenID = %v, want token-yes", result.TokenID)
				}
				expectedPrice := decimal.NewFromFloat(0.55)
				if result.Price == nil || !result.Price.Equal(expectedPrice) {
					t.Errorf("Price = %v, want %v", result.Price, expectedPrice)
				}
				if result.Side != types.OrderSideSell {
					t.Errorf("Side = %v, want %v", result.Side, types.OrderSideSell)
				}
			},
		},
		{
			name: "Sell market order NO @ 0.8 converts to Buy market order YES @ 0.2",
			order: &types.UserMarketOrder{
				TokenID: "token-no",
				Price:   func() *decimal.Decimal { p := decimal.NewFromFloat(0.8); return &p }(),
				Amount:  decimal.NewFromInt(75),
				Side:    types.OrderSideSell,
			},
			complementaryTokenID: "token-yes",
			wantErr:              false,
			validateResult: func(t *testing.T, result *types.UserMarketOrder) {
				if result.TokenID != "token-yes" {
					t.Errorf("TokenID = %v, want token-yes", result.TokenID)
				}
				expectedPrice := decimal.NewFromFloat(0.2)
				if result.Price == nil || !result.Price.Equal(expectedPrice) {
					t.Errorf("Price = %v, want %v", result.Price, expectedPrice)
				}
				if result.Side != types.OrderSideBuy {
					t.Errorf("Side = %v, want %v", result.Side, types.OrderSideBuy)
				}
			},
		},
		{
			name: "Preserves optional fields: FeeRateBps, Nonce, Taker",
			order: &types.UserMarketOrder{
				TokenID:    "token-yes",
				Price:      func() *decimal.Decimal { p := decimal.NewFromFloat(0.5); return &p }(),
				Amount:     decimal.NewFromInt(100),
				Side:       types.OrderSideBuy,
				FeeRateBps: func() *int { v := 10; return &v }(),
				Nonce:      big.NewInt(12345),
				Taker:      "0xABCDEF1234567890",
			},
			complementaryTokenID: "token-no",
			wantErr:              false,
			validateResult: func(t *testing.T, result *types.UserMarketOrder) {
				if result.FeeRateBps == nil || *result.FeeRateBps != 10 {
					t.Errorf("FeeRateBps = %v, want 10", result.FeeRateBps)
				}
				if result.Nonce.Cmp(big.NewInt(12345)) != 0 {
					t.Errorf("Nonce = %v, want 12345", result.Nonce)
				}
				if result.Taker != "0xABCDEF1234567890" {
					t.Errorf("Taker = %v, want 0xABCDEF1234567890", result.Taker)
				}
			},
		},
		{
			name: "Edge case: Price 0 converts to Price 1",
			order: &types.UserMarketOrder{
				TokenID: "token-yes",
				Price:   func() *decimal.Decimal { p := decimal.Zero; return &p }(),
				Amount:  decimal.NewFromInt(10),
				Side:    types.OrderSideBuy,
			},
			complementaryTokenID: "token-no",
			wantErr:              false,
			validateResult: func(t *testing.T, result *types.UserMarketOrder) {
				expectedPrice := decimal.NewFromInt(1)
				if result.Price == nil || !result.Price.Equal(expectedPrice) {
					t.Errorf("Price = %v, want %v", result.Price, expectedPrice)
				}
			},
		},
		{
			name: "Edge case: Price 1 converts to Price 0",
			order: &types.UserMarketOrder{
				TokenID: "token-yes",
				Price:   func() *decimal.Decimal { p := decimal.NewFromInt(1); return &p }(),
				Amount:  decimal.NewFromInt(10),
				Side:    types.OrderSideBuy,
			},
			complementaryTokenID: "token-no",
			wantErr:              false,
			validateResult: func(t *testing.T, result *types.UserMarketOrder) {
				if result.Price == nil || !result.Price.Equal(decimal.Zero) {
					t.Errorf("Price = %v, want 0", result.Price)
				}
			},
		},
		{
			name:                 "Error: nil market order",
			order:                nil,
			complementaryTokenID: "token-no",
			wantErr:              true,
		},
		{
			name: "Error: market order without price",
			order: &types.UserMarketOrder{
				TokenID: "token-yes",
				Price:   nil,
				Amount:  decimal.NewFromInt(10),
				Side:    types.OrderSideBuy,
			},
			complementaryTokenID: "token-no",
			wantErr:              true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ConvertMarketOrderToComplementary(tt.order, tt.complementaryTokenID)

			if (err != nil) != tt.wantErr {
				t.Errorf("ConvertMarketOrderToComplementary() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err == nil && tt.validateResult != nil {
				tt.validateResult(t, result)
			}
		})
	}
}

func TestMarketOrderToLimitOrder(t *testing.T) {
	tests := []struct {
		name           string
		order          *types.UserMarketOrder
		wantErr        bool
		validateResult func(*testing.T, *types.UserOrder)
	}{
		{
			name: "Convert market order with all fields",
			order: &types.UserMarketOrder{
				TokenID:    "token-yes",
				Price:      func() *decimal.Decimal { p := decimal.NewFromFloat(0.6); return &p }(),
				Amount:     decimal.NewFromInt(100),
				Side:       types.OrderSideBuy,
				FeeRateBps: func() *int { v := 10; return &v }(),
				Nonce:      big.NewInt(12345),
				Taker:      "0xABCDEF1234567890",
			},
			wantErr: false,
			validateResult: func(t *testing.T, result *types.UserOrder) {
				if result.TokenID != "token-yes" {
					t.Errorf("TokenID = %v, want token-yes", result.TokenID)
				}
				expectedPrice := decimal.NewFromFloat(0.6)
				if !result.Price.Equal(expectedPrice) {
					t.Errorf("Price = %v, want %v", result.Price, expectedPrice)
				}
				if !result.Size.Equal(decimal.NewFromInt(100)) {
					t.Errorf("Size = %v, want 100", result.Size)
				}
				if result.Side != types.OrderSideBuy {
					t.Errorf("Side = %v, want %v", result.Side, types.OrderSideBuy)
				}
				if result.FeeRateBps != 10 {
					t.Errorf("FeeRateBps = %v, want 10", result.FeeRateBps)
				}
				if result.Nonce.Cmp(big.NewInt(12345)) != 0 {
					t.Errorf("Nonce = %v, want 12345", result.Nonce)
				}
				if result.Taker != "0xABCDEF1234567890" {
					t.Errorf("Taker = %v, want 0xABCDEF1234567890", result.Taker)
				}
			},
		},
		{
			name: "Convert market order without optional fields",
			order: &types.UserMarketOrder{
				TokenID: "token-yes",
				Price:   func() *decimal.Decimal { p := decimal.NewFromFloat(0.5); return &p }(),
				Amount:  decimal.NewFromInt(50),
				Side:    types.OrderSideSell,
			},
			wantErr: false,
			validateResult: func(t *testing.T, result *types.UserOrder) {
				if result.FeeRateBps != 0 {
					t.Errorf("FeeRateBps = %v, want 0", result.FeeRateBps)
				}
			},
		},
		{
			name: "Error: market order without price",
			order: &types.UserMarketOrder{
				TokenID: "token-yes",
				Price:   nil,
				Amount:  decimal.NewFromInt(10),
				Side:    types.OrderSideBuy,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := marketOrderToLimitOrder(tt.order)

			if (err != nil) != tt.wantErr {
				t.Errorf("marketOrderToLimitOrder() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err == nil && tt.validateResult != nil {
				tt.validateResult(t, result)
			}
		})
	}
}

func TestLimitOrderToMarketOrder(t *testing.T) {
	tests := []struct {
		name           string
		order          *types.UserOrder
		amount         decimal.Decimal
		validateResult func(*testing.T, *types.UserMarketOrder)
	}{
		{
			name: "Convert limit order with all fields",
			order: &types.UserOrder{
				TokenID:    "token-yes",
				Price:      decimal.NewFromFloat(0.6),
				Size:       decimal.NewFromInt(100),
				Side:       types.OrderSideBuy,
				FeeRateBps: 10,
				Nonce:      big.NewInt(12345),
				Expiration: func() *int64 { v := int64(1234567890); return &v }(),
				Taker:      "0xABCDEF1234567890",
			},
			amount: decimal.NewFromInt(100),
			validateResult: func(t *testing.T, result *types.UserMarketOrder) {
				if result.TokenID != "token-yes" {
					t.Errorf("TokenID = %v, want token-yes", result.TokenID)
				}
				expectedPrice := decimal.NewFromFloat(0.6)
				if result.Price == nil || !result.Price.Equal(expectedPrice) {
					t.Errorf("Price = %v, want %v", result.Price, expectedPrice)
				}
				if !result.Amount.Equal(decimal.NewFromInt(100)) {
					t.Errorf("Amount = %v, want 100", result.Amount)
				}
				if result.Side != types.OrderSideBuy {
					t.Errorf("Side = %v, want %v", result.Side, types.OrderSideBuy)
				}
				if result.FeeRateBps == nil || *result.FeeRateBps != 10 {
					t.Errorf("FeeRateBps = %v, want 10", result.FeeRateBps)
				}
				if result.Nonce.Cmp(big.NewInt(12345)) != 0 {
					t.Errorf("Nonce = %v, want 12345", result.Nonce)
				}
				if result.Taker != "0xABCDEF1234567890" {
					t.Errorf("Taker = %v, want 0xABCDEF1234567890", result.Taker)
				}
			},
		},
		{
			name: "Convert limit order without optional fields",
			order: &types.UserOrder{
				TokenID: "token-yes",
				Price:   decimal.NewFromFloat(0.5),
				Size:    decimal.NewFromInt(50),
				Side:    types.OrderSideSell,
			},
			amount: decimal.NewFromInt(50),
			validateResult: func(t *testing.T, result *types.UserMarketOrder) {
				if result.FeeRateBps != nil {
					t.Errorf("FeeRateBps = %v, want nil", result.FeeRateBps)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := limitOrderToMarketOrder(tt.order, tt.amount)
			if err != nil {
				t.Fatalf("limitOrderToMarketOrder() error = %v", err)
			}

			if tt.validateResult != nil {
				tt.validateResult(t, result)
			}
		})
	}
}

func TestConvertToOppositeSideOrder(t *testing.T) {
	tests := []struct {
		name           string
		order          *types.UserOrder
		wantErr        bool
		validateResult func(*testing.T, *types.UserOrder)
	}{
		{
			name: "Buy YES @ 0.49 converts to Sell YES @ 0.49",
			order: &types.UserOrder{
				TokenID:    "token-yes",
				Price:      decimal.NewFromFloat(0.49),
				Size:       decimal.NewFromInt(100),
				Side:       types.OrderSideBuy,
				FeeRateBps: 10,
			},
			wantErr: false,
			validateResult: func(t *testing.T, result *types.UserOrder) {
				if result.TokenID != "token-yes" {
					t.Errorf("TokenID = %v, want token-yes", result.TokenID)
				}
				expectedPrice := decimal.NewFromFloat(0.49)
				if !result.Price.Equal(expectedPrice) {
					t.Errorf("Price = %v, want %v", result.Price, expectedPrice)
				}
				if result.Side != types.OrderSideSell {
					t.Errorf("Side = %v, want %v", result.Side, types.OrderSideSell)
				}
				if !result.Size.Equal(decimal.NewFromInt(100)) {
					t.Errorf("Size = %v, want 100", result.Size)
				}
				if result.FeeRateBps != 10 {
					t.Errorf("FeeRateBps = %v, want 10", result.FeeRateBps)
				}
			},
		},
		{
			name: "Sell NO @ 0.51 converts to Buy NO @ 0.51",
			order: &types.UserOrder{
				TokenID:    "token-no",
				Price:      decimal.NewFromFloat(0.51),
				Size:       decimal.NewFromInt(50),
				Side:       types.OrderSideSell,
				FeeRateBps: 15,
			},
			wantErr: false,
			validateResult: func(t *testing.T, result *types.UserOrder) {
				if result.TokenID != "token-no" {
					t.Errorf("TokenID = %v, want token-no", result.TokenID)
				}
				expectedPrice := decimal.NewFromFloat(0.51)
				if !result.Price.Equal(expectedPrice) {
					t.Errorf("Price = %v, want %v", result.Price, expectedPrice)
				}
				if result.Side != types.OrderSideBuy {
					t.Errorf("Side = %v, want %v", result.Side, types.OrderSideBuy)
				}
				if !result.Size.Equal(decimal.NewFromInt(50)) {
					t.Errorf("Size = %v, want 50", result.Size)
				}
				if result.FeeRateBps != 15 {
					t.Errorf("FeeRateBps = %v, want 15", result.FeeRateBps)
				}
			},
		},
		{
			name:    "Nil order returns error",
			order:   nil,
			wantErr: true,
		},
		{
			name: "Preserves all order fields except side",
			order: &types.UserOrder{
				TokenID:    "token-test",
				Price:      decimal.NewFromFloat(0.75),
				Size:       decimal.NewFromInt(200),
				Side:       types.OrderSideBuy,
				FeeRateBps: 20,
				Nonce:      big.NewInt(12345),
				Expiration: func() *int64 { exp := int64(99999); return &exp }(),
				Taker:      "0x1234567890abcdef",
			},
			wantErr: false,
			validateResult: func(t *testing.T, result *types.UserOrder) {
				if result.TokenID != "token-test" {
					t.Errorf("TokenID = %v, want token-test", result.TokenID)
				}
				expectedPrice := decimal.NewFromFloat(0.75)
				if !result.Price.Equal(expectedPrice) {
					t.Errorf("Price = %v, want %v", result.Price, expectedPrice)
				}
				if !result.Size.Equal(decimal.NewFromInt(200)) {
					t.Errorf("Size = %v, want 200", result.Size)
				}
				if result.Side != types.OrderSideSell {
					t.Errorf("Side = %v, want %v", result.Side, types.OrderSideSell)
				}
				if result.FeeRateBps != 20 {
					t.Errorf("FeeRateBps = %v, want 20", result.FeeRateBps)
				}
				if result.Nonce.Cmp(big.NewInt(12345)) != 0 {
					t.Errorf("Nonce = %v, want 12345", result.Nonce)
				}
				if result.Expiration == nil || *result.Expiration != 99999 {
					t.Errorf("Expiration = %v, want 99999", result.Expiration)
				}
				if result.Taker != "0x1234567890abcdef" {
					t.Errorf("Taker = %v, want 0x1234567890abcdef", result.Taker)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ConvertToOppositeSideOrder(tt.order, decimal.Zero)

			if (err != nil) != tt.wantErr {
				t.Errorf("ConvertToOppositeSideOrder() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err == nil && tt.validateResult != nil {
				tt.validateResult(t, result)
			}
		})
	}
}

func TestConvertMarketOrderToOppositeSide(t *testing.T) {
	tests := []struct {
		name           string
		order          *types.UserMarketOrder
		wantErr        bool
		validateResult func(*testing.T, *types.UserMarketOrder)
	}{
		{
			name: "Buy YES @ 0.49 converts to Sell YES @ 0.49",
			order: &types.UserMarketOrder{
				TokenID: "token-yes",
				Price:   func() *decimal.Decimal { p := decimal.NewFromFloat(0.49); return &p }(),
				Amount:  decimal.NewFromInt(100),
				Side:    types.OrderSideBuy,
			},
			wantErr: false,
			validateResult: func(t *testing.T, result *types.UserMarketOrder) {
				if result.TokenID != "token-yes" {
					t.Errorf("TokenID = %v, want token-yes", result.TokenID)
				}
				expectedPrice := decimal.NewFromFloat(0.49)
				if result.Price == nil || !result.Price.Equal(expectedPrice) {
					t.Errorf("Price = %v, want %v", result.Price, expectedPrice)
				}
				if result.Side != types.OrderSideSell {
					t.Errorf("Side = %v, want %v", result.Side, types.OrderSideSell)
				}
				if !result.Amount.Equal(decimal.NewFromInt(100)) {
					t.Errorf("Amount = %v, want 100", result.Amount)
				}
			},
		},
		{
			name: "Sell NO @ 0.51 converts to Buy NO @ 0.51",
			order: &types.UserMarketOrder{
				TokenID: "token-no",
				Price:   func() *decimal.Decimal { p := decimal.NewFromFloat(0.51); return &p }(),
				Amount:  decimal.NewFromInt(50),
				Side:    types.OrderSideSell,
			},
			wantErr: false,
			validateResult: func(t *testing.T, result *types.UserMarketOrder) {
				if result.TokenID != "token-no" {
					t.Errorf("TokenID = %v, want token-no", result.TokenID)
				}
				expectedPrice := decimal.NewFromFloat(0.51)
				if result.Price == nil || !result.Price.Equal(expectedPrice) {
					t.Errorf("Price = %v, want %v", result.Price, expectedPrice)
				}
				if result.Side != types.OrderSideBuy {
					t.Errorf("Side = %v, want %v", result.Side, types.OrderSideBuy)
				}
				if !result.Amount.Equal(decimal.NewFromInt(50)) {
					t.Errorf("Amount = %v, want 50", result.Amount)
				}
			},
		},
		{
			name:    "Nil order returns error",
			order:   nil,
			wantErr: true,
		},
		{
			name: "Preserves all order fields except side",
			order: &types.UserMarketOrder{
				TokenID:    "token-test",
				Price:      func() *decimal.Decimal { p := decimal.NewFromFloat(0.75); return &p }(),
				Amount:     decimal.NewFromInt(200),
				Side:       types.OrderSideBuy,
				FeeRateBps: func() *int { fee := 20; return &fee }(),
				Nonce:      big.NewInt(12345),
				Taker:      "0x1234567890abcdef",
			},
			wantErr: false,
			validateResult: func(t *testing.T, result *types.UserMarketOrder) {
				if result.TokenID != "token-test" {
					t.Errorf("TokenID = %v, want token-test", result.TokenID)
				}
				expectedPrice := decimal.NewFromFloat(0.75)
				if result.Price == nil || !result.Price.Equal(expectedPrice) {
					t.Errorf("Price = %v, want %v", result.Price, expectedPrice)
				}
				if !result.Amount.Equal(decimal.NewFromInt(200)) {
					t.Errorf("Amount = %v, want 200", result.Amount)
				}
				if result.Side != types.OrderSideSell {
					t.Errorf("Side = %v, want %v", result.Side, types.OrderSideSell)
				}
				if result.FeeRateBps == nil || *result.FeeRateBps != 20 {
					t.Errorf("FeeRateBps = %v, want 20", result.FeeRateBps)
				}
				if result.Nonce.Cmp(big.NewInt(12345)) != 0 {
					t.Errorf("Nonce = %v, want 12345", result.Nonce)
				}
				if result.Taker != "0x1234567890abcdef" {
					t.Errorf("Taker = %v, want 0x1234567890abcdef", result.Taker)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ConvertMarketOrderToOppositeSide(tt.order, decimal.Zero)

			if (err != nil) != tt.wantErr {
				t.Errorf("ConvertMarketOrderToOppositeSide() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err == nil && tt.validateResult != nil {
				tt.validateResult(t, result)
			}
		})
	}
}

// TestCombinedConversion tests the combination of opposite side + complementary conversion
// This demonstrates how to create matching orders on the same side
func TestCombinedConversion(t *testing.T) {
	t.Run("Buy YES @ 0.49 -> Opposite -> Complementary -> Buy NO @ 0.51", func(t *testing.T) {
		// Original: Buy YES @ 0.49
		original := &types.UserOrder{
			TokenID:    "token-yes",
			Price:      decimal.NewFromFloat(0.49),
			Size:       decimal.NewFromInt(100),
			Side:       types.OrderSideBuy,
			FeeRateBps: 10,
		}

		// Step 1: Convert to opposite side -> Sell YES @ 0.49
		opposite, err := ConvertToOppositeSideOrder(original, decimal.Zero)
		if err != nil {
			t.Fatalf("ConvertToOppositeSideOrder() error = %v", err)
		}

		if opposite.Side != types.OrderSideSell {
			t.Errorf("After opposite conversion, Side = %v, want %v", opposite.Side, types.OrderSideSell)
		}
		if opposite.TokenID != "token-yes" {
			t.Errorf("After opposite conversion, TokenID = %v, want token-yes", opposite.TokenID)
		}
		if !opposite.Price.Equal(decimal.NewFromFloat(0.49)) {
			t.Errorf("After opposite conversion, Price = %v, want 0.49", opposite.Price)
		}

		// Step 2: Convert to complementary -> Buy NO @ 0.51
		complementary, err := ConvertToComplementaryOrder(opposite, "token-no")
		if err != nil {
			t.Fatalf("ConvertToComplementaryOrder() error = %v", err)
		}

		if complementary.Side != types.OrderSideBuy {
			t.Errorf("After complementary conversion, Side = %v, want %v", complementary.Side, types.OrderSideBuy)
		}
		if complementary.TokenID != "token-no" {
			t.Errorf("After complementary conversion, TokenID = %v, want token-no", complementary.TokenID)
		}
		expectedPrice := decimal.NewFromFloat(0.51)
		if !complementary.Price.Equal(expectedPrice) {
			t.Errorf("After complementary conversion, Price = %v, want 0.51", complementary.Price)
		}

		// Verify: Both orders are BUY and prices sum to 1.0
		if original.Side != types.OrderSideBuy || complementary.Side != types.OrderSideBuy {
			t.Errorf("Result: both orders should be BUY, got %v and %v", original.Side, complementary.Side)
		}

		priceSum := original.Price.Add(complementary.Price)
		expectedSum := decimal.NewFromInt(1)
		if !priceSum.Equal(expectedSum) {
			t.Errorf("Prices should sum to 1.0, got %v", priceSum)
		}
	})

	t.Run("Sell NO @ 0.45 -> Opposite -> Complementary -> Sell YES @ 0.55", func(t *testing.T) {
		// Original: Sell NO @ 0.45
		original := &types.UserOrder{
			TokenID:    "token-no",
			Price:      decimal.NewFromFloat(0.45),
			Size:       decimal.NewFromInt(50),
			Side:       types.OrderSideSell,
			FeeRateBps: 10,
		}

		// Step 1: Convert to opposite side -> Buy NO @ 0.45
		opposite, err := ConvertToOppositeSideOrder(original, decimal.Zero)
		if err != nil {
			t.Fatalf("ConvertToOppositeSideOrder() error = %v", err)
		}

		if opposite.Side != types.OrderSideBuy {
			t.Errorf("After opposite conversion, Side = %v, want %v", opposite.Side, types.OrderSideBuy)
		}

		// Step 2: Convert to complementary -> Sell YES @ 0.55
		complementary, err := ConvertToComplementaryOrder(opposite, "token-yes")
		if err != nil {
			t.Fatalf("ConvertToComplementaryOrder() error = %v", err)
		}

		if complementary.Side != types.OrderSideSell {
			t.Errorf("After complementary conversion, Side = %v, want %v", complementary.Side, types.OrderSideSell)
		}
		if complementary.TokenID != "token-yes" {
			t.Errorf("After complementary conversion, TokenID = %v, want token-yes", complementary.TokenID)
		}
		expectedPrice := decimal.NewFromFloat(0.55)
		if !complementary.Price.Equal(expectedPrice) {
			t.Errorf("After complementary conversion, Price = %v, want 0.55", complementary.Price)
		}

		// Verify: Both orders are SELL and prices sum to 1.0
		if original.Side != types.OrderSideSell || complementary.Side != types.OrderSideSell {
			t.Errorf("Result: both orders should be SELL, got %v and %v", original.Side, complementary.Side)
		}

		priceSum := original.Price.Add(complementary.Price)
		expectedSum := decimal.NewFromInt(1)
		if !priceSum.Equal(expectedSum) {
			t.Errorf("Prices should sum to 1.0, got %v", priceSum)
		}
	})
}

// TestConvertToOppositeSideOrderWithSpread tests the spread functionality for market making
func TestConvertToOppositeSideOrderWithSpread(t *testing.T) {
	tests := []struct {
		name           string
		order          *types.UserOrder
		spread         decimal.Decimal
		wantErr        bool
		validateResult func(*testing.T, *types.UserOrder)
	}{
		{
			name: "Buy YES @ 0.49 with spread 0.02 -> Sell YES @ 0.51",
			order: &types.UserOrder{
				TokenID:    "token-yes",
				Price:      decimal.NewFromFloat(0.49),
				Size:       decimal.NewFromInt(100),
				Side:       types.OrderSideBuy,
				FeeRateBps: 10,
			},
			spread:  decimal.NewFromFloat(0.02),
			wantErr: false,
			validateResult: func(t *testing.T, result *types.UserOrder) {
				if result.TokenID != "token-yes" {
					t.Errorf("TokenID = %v, want token-yes", result.TokenID)
				}
				expectedPrice := decimal.NewFromFloat(0.51)
				if !result.Price.Equal(expectedPrice) {
					t.Errorf("Price = %v, want %v", result.Price, expectedPrice)
				}
				if result.Side != types.OrderSideSell {
					t.Errorf("Side = %v, want %v", result.Side, types.OrderSideSell)
				}
			},
		},
		{
			name: "Sell YES @ 0.51 with spread 0.02 -> Buy YES @ 0.49",
			order: &types.UserOrder{
				TokenID:    "token-yes",
				Price:      decimal.NewFromFloat(0.51),
				Size:       decimal.NewFromInt(100),
				Side:       types.OrderSideSell,
				FeeRateBps: 10,
			},
			spread:  decimal.NewFromFloat(0.02),
			wantErr: false,
			validateResult: func(t *testing.T, result *types.UserOrder) {
				if result.TokenID != "token-yes" {
					t.Errorf("TokenID = %v, want token-yes", result.TokenID)
				}
				expectedPrice := decimal.NewFromFloat(0.49)
				if !result.Price.Equal(expectedPrice) {
					t.Errorf("Price = %v, want %v", result.Price, expectedPrice)
				}
				if result.Side != types.OrderSideBuy {
					t.Errorf("Side = %v, want %v", result.Side, types.OrderSideBuy)
				}
			},
		},
		{
			name: "Negative spread should return error",
			order: &types.UserOrder{
				TokenID:    "token-yes",
				Price:      decimal.NewFromFloat(0.5),
				Size:       decimal.NewFromInt(100),
				Side:       types.OrderSideBuy,
				FeeRateBps: 10,
			},
			spread:  decimal.NewFromFloat(-0.01),
			wantErr: true,
		},
		{
			name: "Spread causing price > 1.0 should return error",
			order: &types.UserOrder{
				TokenID:    "token-yes",
				Price:      decimal.NewFromFloat(0.95),
				Size:       decimal.NewFromInt(100),
				Side:       types.OrderSideBuy,
				FeeRateBps: 10,
			},
			spread:  decimal.NewFromFloat(0.10), // 0.95 + 0.10 = 1.05 > 1.0
			wantErr: true,
		},
		{
			name: "Spread causing price < 0.0 should return error",
			order: &types.UserOrder{
				TokenID:    "token-yes",
				Price:      decimal.NewFromFloat(0.05),
				Size:       decimal.NewFromInt(100),
				Side:       types.OrderSideSell,
				FeeRateBps: 10,
			},
			spread:  decimal.NewFromFloat(0.10), // 0.05 - 0.10 = -0.05 < 0.0
			wantErr: true,
		},
		{
			name: "Large spread within valid range",
			order: &types.UserOrder{
				TokenID:    "token-yes",
				Price:      decimal.NewFromFloat(0.3),
				Size:       decimal.NewFromInt(100),
				Side:       types.OrderSideBuy,
				FeeRateBps: 10,
			},
			spread:  decimal.NewFromFloat(0.5),
			wantErr: false,
			validateResult: func(t *testing.T, result *types.UserOrder) {
				expectedPrice := decimal.NewFromFloat(0.8)
				if !result.Price.Equal(expectedPrice) {
					t.Errorf("Price = %v, want %v", result.Price, expectedPrice)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ConvertToOppositeSideOrder(tt.order, tt.spread)

			if (err != nil) != tt.wantErr {
				t.Errorf("ConvertToOppositeSideOrder() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err == nil && tt.validateResult != nil {
				tt.validateResult(t, result)
			}
		})
	}
}

// TestMarketMakingWithSpread demonstrates the complete market making flow
func TestMarketMakingWithSpread(t *testing.T) {
	t.Run("Market making with spread creates profitable matching orders", func(t *testing.T) {
		// Original: Buy YES @ 0.49
		original := &types.UserOrder{
			TokenID:    "token-yes",
			Price:      decimal.NewFromFloat(0.49),
			Size:       decimal.NewFromInt(100),
			Side:       types.OrderSideBuy,
			FeeRateBps: 10,
		}

		spread := decimal.NewFromFloat(0.02)

		// Step 1: Convert to opposite side with spread -> Sell YES @ 0.51 (0.49 + 0.02)
		opposite, err := ConvertToOppositeSideOrder(original, spread)
		if err != nil {
			t.Fatalf("ConvertToOppositeSideOrder() error = %v", err)
		}

		if !opposite.Price.Equal(decimal.NewFromFloat(0.51)) {
			t.Errorf("After opposite conversion with spread, Price = %v, want 0.51", opposite.Price)
		}
		if opposite.Side != types.OrderSideSell {
			t.Errorf("After opposite conversion, Side = %v, want %v", opposite.Side, types.OrderSideSell)
		}

		// Step 2: Convert to complementary -> Buy NO @ 0.49 (1 - 0.51)
		complementary, err := ConvertToComplementaryOrder(opposite, "token-no")
		if err != nil {
			t.Fatalf("ConvertToComplementaryOrder() error = %v", err)
		}

		if !complementary.Price.Equal(decimal.NewFromFloat(0.49)) {
			t.Errorf("After complementary conversion, Price = %v, want 0.49", complementary.Price)
		}
		if complementary.Side != types.OrderSideBuy {
			t.Errorf("After complementary conversion, Side = %v, want %v", complementary.Side, types.OrderSideBuy)
		}

		// Verify: Both orders are BUY and cost sum is less than 1.0 (profitable)
		if original.Side != types.OrderSideBuy || complementary.Side != types.OrderSideBuy {
			t.Errorf("Result: both orders should be BUY, got %v and %v", original.Side, complementary.Side)
		}

		costSum := original.Price.Add(complementary.Price)
		// Cost: 0.49 + 0.49 = 0.98 < 1.0, profit = 0.02
		expectedCost := decimal.NewFromFloat(0.98)
		if !costSum.Equal(expectedCost) {
			t.Errorf("Cost sum = %v, want %v (profit = %v)", costSum, expectedCost, decimal.NewFromInt(1).Sub(costSum))
		}

		profit := decimal.NewFromInt(1).Sub(costSum)
		expectedProfit := spread
		if !profit.Equal(expectedProfit) {
			t.Errorf("Profit = %v, want %v", profit, expectedProfit)
		}
	})

	t.Run("Sell orders with spread also create profitable matching", func(t *testing.T) {
		// Original: Sell NO @ 0.45
		original := &types.UserOrder{
			TokenID:    "token-no",
			Price:      decimal.NewFromFloat(0.45),
			Size:       decimal.NewFromInt(100),
			Side:       types.OrderSideSell,
			FeeRateBps: 10,
		}

		spread := decimal.NewFromFloat(0.02)

		// Step 1: Convert to opposite side with spread -> Buy NO @ 0.43 (0.45 - 0.02)
		opposite, err := ConvertToOppositeSideOrder(original, spread)
		if err != nil {
			t.Fatalf("ConvertToOppositeSideOrder() error = %v", err)
		}

		if !opposite.Price.Equal(decimal.NewFromFloat(0.43)) {
			t.Errorf("After opposite conversion with spread, Price = %v, want 0.43", opposite.Price)
		}

		// Step 2: Convert to complementary -> Sell YES @ 0.57 (1 - 0.43)
		complementary, err := ConvertToComplementaryOrder(opposite, "token-yes")
		if err != nil {
			t.Fatalf("ConvertToComplementaryOrder() error = %v", err)
		}

		if !complementary.Price.Equal(decimal.NewFromFloat(0.57)) {
			t.Errorf("After complementary conversion, Price = %v, want 0.57", complementary.Price)
		}

		// Verify: Both orders are SELL and revenue sum is greater than 1.0 (profitable)
		if original.Side != types.OrderSideSell || complementary.Side != types.OrderSideSell {
			t.Errorf("Result: both orders should be SELL, got %v and %v", original.Side, complementary.Side)
		}

		revenueSum := original.Price.Add(complementary.Price)
		// Revenue: 0.45 + 0.57 = 1.02 > 1.0, profit = 0.02
		expectedRevenue := decimal.NewFromFloat(1.02)
		if !revenueSum.Equal(expectedRevenue) {
			t.Errorf("Revenue sum = %v, want %v (profit = %v)", revenueSum, expectedRevenue, revenueSum.Sub(decimal.NewFromInt(1)))
		}

		profit := revenueSum.Sub(decimal.NewFromInt(1))
		expectedProfit := spread
		if !profit.Equal(expectedProfit) {
			t.Errorf("Profit = %v, want %v", profit, expectedProfit)
		}
	})
}
