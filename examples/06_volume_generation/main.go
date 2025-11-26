package main

import (
	"context"
	"fmt"
	"log"
	"math/big"
	"os"
	"strings"

	"github.com/shopspring/decimal"

	"github.com/ivanzzeth/polymarket-go-clob-client/types"
	"github.com/ivanzzeth/polymarket-go/examples/helper"
)

func main() {
	// Load .env file
	helper.LoadEnv()

	ctx := context.Background()

	// Create Polymarket client with signer
	client, err := helper.NewClientWithSigner(ctx)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	fmt.Println("=== Polymarket Volume Generation Example ===")
	fmt.Println()
	fmt.Println("This example demonstrates how to create matching orders on the same side")
	fmt.Println("by combining opposite side and complementary conversions.")
	fmt.Println()

	// Get token ID from environment or use example
	tokenID := os.Getenv("TEST_TOKEN_ID")
	if tokenID == "" {
		fmt.Println("TEST_TOKEN_ID environment variable not set, using example token ID")
		// Example token ID (this is just for demonstration, use a real one)
		tokenID = "26789704146335935253327636432210126325965975622942457950139452145379746946996"
	}

	// Example 1: Create two matching BUY orders
	fmt.Println(strings.Repeat("=", 60))
	fmt.Println("Example 1: Creating Two Matching BUY Orders")
	fmt.Println(strings.Repeat("=", 60))

	// Original order: Buy YES @ 0.49
	originalOrder := &types.UserOrder{
		TokenID:    tokenID,
		Price:      decimal.NewFromFloat(0.49),
		Size:       decimal.NewFromInt(100),
		Side:       types.OrderSideBuy,
		FeeRateBps: 10,
		Nonce:      big.NewInt(12345),
	}

	fmt.Printf("\n1️⃣  Original Order: %s YES @ %s (%s shares)\n",
		originalOrder.Side,
		originalOrder.Price.String(),
		originalOrder.Size.String(),
	)

	// Step 1: Convert to opposite side (BUY → SELL)
	oppositeOrder, err := client.ConvertLimitOrderToOppositeSide(originalOrder, decimal.Zero)
	if err != nil {
		log.Fatalf("Failed to convert to opposite side: %v", err)
	}

	fmt.Printf("\n2️⃣  After Opposite Side Conversion: %s YES @ %s (%s shares)\n",
		oppositeOrder.Side,
		oppositeOrder.Price.String(),
		oppositeOrder.Size.String(),
	)
	fmt.Println("   ⮑ Only changed BUY → SELL, same token and price")

	// Step 2: Convert to complementary (SELL YES @ 0.49 → BUY NO @ 0.51)
	complementaryOrder, err := client.ConvertLimitOrderToComplementary(ctx, oppositeOrder)
	if err != nil {
		log.Fatalf("Failed to convert to complementary: %v", err)
	}

	fmt.Printf("\n3️⃣  After Complementary Conversion: %s NO @ %s (%s shares)\n",
		complementaryOrder.Side,
		complementaryOrder.Price.String(),
		complementaryOrder.Size.String(),
	)
	fmt.Printf("   Token ID: %s\n", complementaryOrder.TokenID)
	fmt.Println("   ⮑ Changed SELL YES → BUY NO, price 0.49 → 0.51")

	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("📊 Result: Two BUY Orders That Can Match Each Other")
	fmt.Println(strings.Repeat("=", 60))
	fmt.Printf("Order A: %s YES @ %s\n", originalOrder.Side, originalOrder.Price.String())
	fmt.Printf("Order B: %s NO @ %s\n", complementaryOrder.Side, complementaryOrder.Price.String())
	fmt.Println("\n✅ These orders can match because:")
	fmt.Println("   • Both are BUY orders")
	fmt.Println("   • They are on complementary tokens (YES/NO)")
	fmt.Println("   • Prices sum to 1.0 (0.49 + 0.51 = 1.0)")
	fmt.Println("   • This triggers the CTF split operation")

	// Example 2: Create two matching SELL orders
	fmt.Println("\n\n" + strings.Repeat("=", 60))
	fmt.Println("Example 2: Creating Two Matching SELL Orders")
	fmt.Println(strings.Repeat("=", 60))

	// Original order: Sell NO @ 0.45
	originalSellOrder := &types.UserOrder{
		TokenID:    complementaryOrder.TokenID, // Use NO token
		Price:      decimal.NewFromFloat(0.45),
		Size:       decimal.NewFromInt(50),
		Side:       types.OrderSideSell,
		FeeRateBps: 10,
		Nonce:      big.NewInt(12346),
	}

	fmt.Printf("\n1️⃣  Original Order: %s NO @ %s (%s shares)\n",
		originalSellOrder.Side,
		originalSellOrder.Price.String(),
		originalSellOrder.Size.String(),
	)

	// Step 1: Convert to opposite side (SELL → BUY)
	oppositeSellOrder, err := client.ConvertLimitOrderToOppositeSide(originalSellOrder, decimal.Zero)
	if err != nil {
		log.Fatalf("Failed to convert to opposite side: %v", err)
	}

	fmt.Printf("\n2️⃣  After Opposite Side Conversion: %s NO @ %s (%s shares)\n",
		oppositeSellOrder.Side,
		oppositeSellOrder.Price.String(),
		oppositeSellOrder.Size.String(),
	)
	fmt.Println("   ⮑ Only changed SELL → BUY, same token and price")

	// Step 2: Convert to complementary (BUY NO @ 0.45 → SELL YES @ 0.55)
	complementarySellOrder, err := client.ConvertLimitOrderToComplementary(ctx, oppositeSellOrder)
	if err != nil {
		log.Fatalf("Failed to convert to complementary: %v", err)
	}

	fmt.Printf("\n3️⃣  After Complementary Conversion: %s YES @ %s (%s shares)\n",
		complementarySellOrder.Side,
		complementarySellOrder.Price.String(),
		complementarySellOrder.Size.String(),
	)
	fmt.Printf("   Token ID: %s\n", complementarySellOrder.TokenID)
	fmt.Println("   ⮑ Changed BUY NO → SELL YES, price 0.45 → 0.55")

	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("📊 Result: Two SELL Orders That Can Match Each Other")
	fmt.Println(strings.Repeat("=", 60))
	fmt.Printf("Order A: %s NO @ %s\n", originalSellOrder.Side, originalSellOrder.Price.String())
	fmt.Printf("Order B: %s YES @ %s\n", complementarySellOrder.Side, complementarySellOrder.Price.String())
	fmt.Println("\n✅ These orders can match because:")
	fmt.Println("   • Both are SELL orders")
	fmt.Println("   • They are on complementary tokens (NO/YES)")
	fmt.Println("   • Prices sum to 1.0 (0.45 + 0.55 = 1.0)")
	fmt.Println("   • This triggers the CTF merge operation")

	// Example 3: Market making with spread for guaranteed profit (using convenience method)
	fmt.Println("\n\n" + strings.Repeat("=", 60))
	fmt.Println("Example 3: Market Making with Spread (Guaranteed Profit)")
	fmt.Println(strings.Repeat("=", 60))
	fmt.Println("\nUsing ConvertLimitOrderToMatchingSameSide convenience method!")

	// Original order: Buy YES @ 0.48
	marketMakingOrder := &types.UserOrder{
		TokenID:    tokenID,
		Price:      decimal.NewFromFloat(0.48),
		Size:       decimal.NewFromInt(100),
		Side:       types.OrderSideBuy,
		FeeRateBps: 10,
		Nonce:      big.NewInt(12347),
	}

	spread := decimal.NewFromFloat(0.02)

	fmt.Printf("\n1️⃣  Original Order: %s YES @ %s (%s shares)\n",
		marketMakingOrder.Side,
		marketMakingOrder.Price.String(),
		marketMakingOrder.Size.String(),
	)
	fmt.Printf("   Spread: %s\n", spread.String())

	// Use the convenience method to create matching same-side order in ONE call!
	// This automatically:
	// - Converts to opposite side with spread (BUY → SELL @ 0.50)
	// - Converts to complementary token (SELL YES → BUY NO)
	// - Queries complementary token ID automatically
	matchingOrder, err := client.ConvertLimitOrderToMatchingSameSide(ctx, marketMakingOrder, spread)
	if err != nil {
		log.Fatalf("Failed to convert to matching same-side order: %v", err)
	}

	fmt.Printf("\n2️⃣  Matching Order (ONE call!): %s NO @ %s (%s shares)\n",
		matchingOrder.Side,
		matchingOrder.Price.String(),
		matchingOrder.Size.String(),
	)
	fmt.Printf("   Token ID: %s\n", matchingOrder.TokenID)
	fmt.Println("   ⮑ Automatically converted: BUY YES @ 0.48 → BUY NO @ 0.50")
	fmt.Println("   ⮑ How: 0.48 + 0.02 (spread) = 0.50, then 1 - 0.50 = 0.50")

	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("💰 Result: Two BUY Orders with GUARANTEED PROFIT")
	fmt.Println(strings.Repeat("=", 60))
	fmt.Printf("Order A: %s YES @ %s\n", marketMakingOrder.Side, marketMakingOrder.Price.String())
	fmt.Printf("Order B: %s NO @ %s\n", matchingOrder.Side, matchingOrder.Price.String())

	costA := marketMakingOrder.Price
	costB := matchingOrder.Price
	totalCost := costA.Add(costB)
	valueWhenMerged := decimal.NewFromInt(1)
	profit := valueWhenMerged.Sub(totalCost)

	fmt.Println("\n💡 Profit Calculation:")
	fmt.Printf("   • Cost A (Buy YES): %s USDC\n", costA.String())
	fmt.Printf("   • Cost B (Buy NO):  %s USDC\n", costB.String())
	fmt.Printf("   • Total Cost:       %s USDC\n", totalCost.String())
	fmt.Printf("   • Value after merge: %s USDC\n", valueWhenMerged.String())
	fmt.Printf("   • Guaranteed Profit: %s USDC ✅\n", profit.String())
	fmt.Println("\n✅ This is market making with guaranteed profit!")
	fmt.Println("   • Both orders are BUY orders on complementary tokens")
	fmt.Println("   • When both fill, you own YES + NO")
	fmt.Println("   • Merge them to get 1.0 USDC back")
	fmt.Printf("   • Profit equals the spread: %s USDC\n", spread.String())
	fmt.Println("\n🚀 Pro Tip: Use ConvertLimitOrderToMatchingSameSide() for one-step conversion!")

	// Use cases
	fmt.Println("\n\n" + strings.Repeat("=", 60))
	fmt.Println("💡 Use Cases")
	fmt.Println(strings.Repeat("=", 60))
	fmt.Println("\n1. Volume Generation:")
	fmt.Println("   • Create matching orders to generate trading volume")
	fmt.Println("   • Useful for market making and liquidity provision")
	fmt.Println("   • Can demonstrate market activity")

	fmt.Println("\n2. Testing & Development:")
	fmt.Println("   • Test order matching logic")
	fmt.Println("   • Verify CTF split/merge operations")
	fmt.Println("   • Debug trading strategies")

	fmt.Println("\n3. Synthetic Liquidity:")
	fmt.Println("   • Create liquidity on both sides simultaneously")
	fmt.Println("   • Balance inventory across YES/NO tokens")
	fmt.Println("   • Implement sophisticated market making strategies")

	fmt.Println("\n4. Arbitrage Execution:")
	fmt.Println("   • Execute complex multi-leg arbitrage trades")
	fmt.Println("   • Utilize CTF operations for capital efficiency")
	fmt.Println("   • Minimize external market exposure")

	fmt.Println("\n\n⚠️  Important Notes:")
	fmt.Println("   • These operations use real funds when executed")
	fmt.Println("   • Ensure you have sufficient collateral for CTF operations")
	fmt.Println("   • Consider gas fees and slippage in production")
	fmt.Println("   • Use appropriate risk management")

	fmt.Println("\n=== Example Complete ===")
}
