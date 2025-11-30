package polymarket

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/ivanzzeth/polymarket-go-clob-client/types"
	"github.com/ivanzzeth/polymarket-go/constants"
	"github.com/shopspring/decimal"
)

// Candlestick represents a candlestick bar with OHLCV data
type Candlestick struct {
	StartTime time.Time       `json:"startTime"` // Start time of the interval
	EndTime   time.Time       `json:"endTime"`   // End time of the interval
	Interval  time.Duration   `json:"interval"`  // Time interval duration
	Open      decimal.Decimal `json:"open"`      // Opening price
	High      decimal.Decimal `json:"high"`      // Highest price
	Low       decimal.Decimal `json:"low"`       // Lowest price
	Close     decimal.Decimal `json:"close"`     // Closing price
	Volume    decimal.Decimal `json:"volume"`    // Total volume (in shares)
	NumTrades int             `json:"numTrades"` // Number of trades in this bar
}

type pricePoint struct {
	Timestamp int64
	Price     decimal.Decimal
	Size      decimal.Decimal
	Side      types.OrderSide
}

type orderFilledEvent struct {
	ID                string `json:"id"`
	Timestamp         string `json:"timestamp"`
	Maker             string `json:"maker"`
	MakerAssetID      string `json:"makerAssetId"`
	MakerAmountFilled string `json:"makerAmountFilled"`
	Taker             string `json:"taker"`
	TakerAssetID      string `json:"takerAssetId"`
	TakerAmountFilled string `json:"takerAmountFilled"`
	TransactionHash   string `json:"transactionHash"`
}

type graphQLResponse struct {
	Data struct {
		OrderFilledEvents []orderFilledEvent `json:"orderFilledEvents"`
	} `json:"data"`
	Errors []struct {
		Message string `json:"message"`
	} `json:"errors"`
}

const getOrderFilledEventsQuery = `
query GetOrderFilledEvents(
  $tokenId: String!
  $startTimestamp: BigInt!
  $endTimestamp: BigInt!
  $first: Int!
  $skip: Int!
) {
  orderFilledEvents(
    where: {
      or: [
        { makerAssetId: $tokenId, timestamp_gte: $startTimestamp, timestamp_lte: $endTimestamp }
        { takerAssetId: $tokenId, timestamp_gte: $startTimestamp, timestamp_lte: $endTimestamp }
      ]
    }
    orderBy: timestamp
    orderDirection: asc
    first: $first
    skip: $skip
  ) {
    id
    timestamp
    maker
    makerAssetId
    makerAmountFilled
    taker
    takerAssetId
    takerAmountFilled
    transactionHash
  }
}
`

// GetCandlesticks fetches OHLC candlestick data for a token from on-chain data
// If fillGaps is true, missing intervals will be filled with the previous close price
func (c *Client) GetCandlesticks(ctx context.Context, tokenID string, interval time.Duration, startTime, endTime time.Time, fillGaps bool) ([]Candlestick, error) {
	if interval <= 0 {
		return nil, fmt.Errorf("interval must be positive")
	}
	if endTime.Before(startTime) {
		return nil, fmt.Errorf("endTime must be after startTime")
	}

	events, err := fetchAllOrderFilledEvents(ctx, c.httpClient, tokenID, startTime, endTime)
	if err != nil {
		return nil, err
	}

	if len(events) == 0 {
		return []Candlestick{}, nil
	}

	pricePoints := convertEventsToPrice(events, tokenID)
	candlesticks := aggregateToCandlesticks(pricePoints, interval)

	if fillGaps && len(candlesticks) > 0 {
		candlesticks = fillCandlestickGaps(candlesticks, interval, startTime, endTime)
	}

	return candlesticks, nil
}

func fetchAllOrderFilledEvents(ctx context.Context, httpClient *http.Client, tokenID string, startTime, endTime time.Time) ([]orderFilledEvent, error) {
	var allEvents []orderFilledEvent
	tokenIDLower := strings.ToLower(tokenID)
	startTimestamp := startTime.Unix()
	endTimestamp := endTime.Unix()

	const batchSize = 1000
	skip := 0

	for {
		events, err := fetchOrderFilledEvents(ctx, httpClient, tokenIDLower, startTimestamp, endTimestamp, batchSize, skip)
		if err != nil {
			return nil, err
		}

		allEvents = append(allEvents, events...)

		if len(events) < batchSize {
			break
		}
		skip += batchSize
	}

	return allEvents, nil
}

func fetchOrderFilledEvents(ctx context.Context, httpClient *http.Client, tokenID string, startTimestamp, endTimestamp int64, first, skip int) ([]orderFilledEvent, error) {
	variables := map[string]interface{}{
		"tokenId":        tokenID,
		"startTimestamp": strconv.FormatInt(startTimestamp, 10),
		"endTimestamp":   strconv.FormatInt(endTimestamp, 10),
		"first":          first,
		"skip":           skip,
	}

	requestBody := map[string]interface{}{
		"query":     getOrderFilledEventsQuery,
		"variables": variables,
	}

	bodyBytes, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, constants.GoldskySubgraphURL, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status %d", resp.StatusCode)
	}

	var result graphQLResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if len(result.Errors) > 0 {
		return nil, fmt.Errorf("GraphQL error: %s", result.Errors[0].Message)
	}

	return result.Data.OrderFilledEvents, nil
}

func convertEventsToPrice(events []orderFilledEvent, tokenID string) []pricePoint {
	tokenIDLower := strings.ToLower(tokenID)
	var points []pricePoint

	for _, event := range events {
		timestamp, _ := strconv.ParseInt(event.Timestamp, 10, 64)
		timestampMs := timestamp * 1000

		makerAssetID := strings.ToLower(event.MakerAssetID)
		takerAssetID := strings.ToLower(event.TakerAssetID)

		makerAmount, _ := decimal.NewFromString(event.MakerAmountFilled)
		takerAmount, _ := decimal.NewFromString(event.TakerAmountFilled)

		var price, size decimal.Decimal
		var side types.OrderSide

		if makerAssetID == tokenIDLower {
			// Maker sells token
			// makerAmountFilled = tokens, takerAmountFilled = USDC
			if !makerAmount.IsZero() {
				price = takerAmount.Div(makerAmount)
				size = makerAmount.Div(decimal.NewFromInt(1e6))
				side = types.OrderSideSell
			}
		} else if takerAssetID == tokenIDLower {
			// Taker buys token
			// makerAmountFilled = USDC, takerAmountFilled = tokens
			if !takerAmount.IsZero() {
				price = makerAmount.Div(takerAmount)
				size = takerAmount.Div(decimal.NewFromInt(1e6))
				side = types.OrderSideBuy
			}
		} else {
			continue
		}

		points = append(points, pricePoint{
			Timestamp: timestampMs,
			Price:     price,
			Size:      size,
			Side:      side,
		})
	}

	return points
}

func aggregateToCandlesticks(points []pricePoint, interval time.Duration) []Candlestick {
	if len(points) == 0 {
		return []Candlestick{}
	}

	intervalMs := interval.Milliseconds()
	candlesMap := make(map[int64]*Candlestick)

	for _, point := range points {
		bucketTime := (point.Timestamp / intervalMs) * intervalMs
		bucketTimeSec := bucketTime / 1000
		intervalStart := time.Unix(bucketTimeSec, 0).UTC()

		if candle, exists := candlesMap[bucketTimeSec]; exists {
			if point.Price.GreaterThan(candle.High) {
				candle.High = point.Price
			}
			if point.Price.LessThan(candle.Low) {
				candle.Low = point.Price
			}
			candle.Close = point.Price
			candle.Volume = candle.Volume.Add(point.Size)
			candle.NumTrades++
		} else {
			candlesMap[bucketTimeSec] = &Candlestick{
				StartTime: intervalStart,
				EndTime:   intervalStart.Add(interval),
				Interval:  interval,
				Open:      point.Price,
				High:      point.Price,
				Low:       point.Price,
				Close:     point.Price,
				Volume:    point.Size,
				NumTrades: 1,
			}
		}
	}

	result := make([]Candlestick, 0, len(candlesMap))
	for _, candle := range candlesMap {
		result = append(result, *candle)
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].StartTime.Before(result[j].StartTime)
	})

	return result
}

func fillCandlestickGaps(candlesticks []Candlestick, interval time.Duration, startTime, endTime time.Time) []Candlestick {
	if len(candlesticks) == 0 {
		return candlesticks
	}

	// Create a map for quick lookup
	candleMap := make(map[int64]Candlestick)
	for _, c := range candlesticks {
		candleMap[c.StartTime.Unix()] = c
	}

	// Align startTime and endTime to interval boundaries
	intervalSec := int64(interval.Seconds())
	alignedStart := (startTime.Unix() / intervalSec) * intervalSec
	alignedEnd := (endTime.Unix() / intervalSec) * intervalSec

	// Use first candle's close as initial previous close
	prevClose := candlesticks[0].Close

	var result []Candlestick
	for t := alignedStart; t <= alignedEnd; t += intervalSec {
		if candle, exists := candleMap[t]; exists {
			result = append(result, candle)
			prevClose = candle.Close
		} else {
			// Fill gap with previous close price (flat candle / 一字线)
			intervalStart := time.Unix(t, 0).UTC()
			result = append(result, Candlestick{
				StartTime: intervalStart,
				EndTime:   intervalStart.Add(interval),
				Interval:  interval,
				Open:      prevClose,
				High:      prevClose,
				Low:       prevClose,
				Close:     prevClose,
				Volume:    decimal.Zero,
				NumTrades: 0,
			})
		}
	}

	return result
}
