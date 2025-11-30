package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

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

	fmt.Println("=== Candlestick (OHLCV) Query Example ===")

	tokenID := os.Getenv("TEST_TOKEN_ID")
	if tokenID == "" {
		log.Fatal("TEST_TOKEN_ID environment variable is required")
	}

	// Query last 10 days of data with 2h candles
	endTime := time.Now().UTC()
	startTime := endTime.Add(-240 * time.Hour)
	interval := 2 * time.Hour

	fmt.Printf("Token ID: %s\n", tokenID)
	fmt.Printf("Time Range: %s to %s\n", startTime.Format(time.RFC3339), endTime.Format(time.RFC3339))
	fmt.Printf("Interval: %v\n\n", interval)

	candles, err := client.GetCandlesticks(ctx, tokenID, interval, startTime, endTime, true)
	if err != nil {
		log.Fatalf("Failed to get candlesticks: %v", err)
	}

	fmt.Printf("=== Results: %d candles ===\n\n", len(candles))
	fmt.Println("Time                 | O      | H      | L      | C      | Volume   | Trades")
	fmt.Println("---------------------|--------|--------|--------|--------|----------|-------")
	for _, candle := range candles {
		fmt.Printf("%-20s | %6s | %6s | %6s | %6s | %8s | %5d\n",
			candle.StartTime.Local().Format("2006-01-02 15:04:05"),
			candle.Open.StringFixed(4),
			candle.High.StringFixed(4),
			candle.Low.StringFixed(4),
			candle.Close.StringFixed(4),
			candle.Volume.StringFixed(2),
			candle.NumTrades)
	}

	fmt.Println("\n=== Example Complete ===")
}
