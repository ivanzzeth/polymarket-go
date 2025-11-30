package main

import (
	"context"
	"fmt"
	"log"
	"time"

	polymarket "github.com/ivanzzeth/polymarket-go"
	"github.com/ivanzzeth/polymarket-go/examples/helper"
)

func main() {
	helper.LoadEnv()

	ctx := context.Background()

	client, err := helper.NewClientWithSigner(ctx)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	defer client.Close()

	fmt.Println("=== Crypto Price Query Example ===")

	// Test with non-aligned time - library should auto-align to 15min window
	// 15:07:30 should align to 15:00:00 - 15:15:00 window
	// Expected: openPrice: 90651.53173357679, closePrice: 90785.51457241473
	startTime := time.Date(2025, 11, 29, 15, 7, 30, 0, time.UTC)

	fmt.Printf("Query: BTC 15min window starting at %s\n\n", startTime.Format(time.RFC3339))

	price, err := client.GetCryptoPrice15MinWindow(ctx, polymarket.CryptoSymbolBTC, startTime)
	if err != nil {
		log.Fatalf("Failed to get crypto price: %v", err)
	}

	fmt.Println("=== Results ===")
	fmt.Printf("Open Price:  %s USD\n", price.OpenPrice.String())
	fmt.Printf("Close Price: %s USD\n", price.ClosePrice.String())

	// Verify against expected values
	expectedOpen := "90651.53173357679"
	expectedClose := "90785.51457241473"

	fmt.Println("\n=== Verification ===")
	if price.OpenPrice.String() == expectedOpen && price.ClosePrice.String() == expectedClose {
		fmt.Println("✓ All prices match expected values!")
	} else {
		fmt.Printf("✗ Mismatch: got open=%s close=%s, expected open=%s close=%s\n",
			price.OpenPrice.String(), price.ClosePrice.String(), expectedOpen, expectedClose)
	}

	fmt.Println("\n=== Example Complete ===")
}
