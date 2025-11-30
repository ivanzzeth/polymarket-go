package polymarket

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/ivanzzeth/polymarket-go/constants"
	"github.com/shopspring/decimal"
)

// CryptoSymbol represents supported cryptocurrency symbols
type CryptoSymbol string

const (
	CryptoSymbolBTC CryptoSymbol = "BTC"
	CryptoSymbolETH CryptoSymbol = "ETH"
)

type cryptoPriceVariant string

const (
	cryptoPriceVariantFifteen cryptoPriceVariant = "fifteen"
)

// CryptoPriceResponse represents the response from the crypto price API
type CryptoPriceResponse struct {
	OpenPrice  decimal.Decimal `json:"openPrice"`
	ClosePrice decimal.Decimal `json:"closePrice"`
}

type cryptoPriceQuery struct {
	Symbol         CryptoSymbol
	EventStartTime time.Time
	EndDate        time.Time
	Variant        cryptoPriceVariant
}

// GetCryptoPrice queries the open/close price for a cryptocurrency with explicit time window
func (c *Client) GetCryptoPrice(ctx context.Context, symbol CryptoSymbol, startTime, endTime time.Time) (*CryptoPriceResponse, error) {
	return getCryptoPrice(ctx, c.httpClient, cryptoPriceQuery{
		Symbol:         symbol,
		EventStartTime: startTime,
		EndDate:        endTime,
		Variant:        cryptoPriceVariantFifteen,
	})
}

// GetCryptoPrice15MinWindow queries the 15-minute window price starting from the given time
// The startTime will be rounded down to the nearest 15-minute interval
func (c *Client) GetCryptoPrice15MinWindow(ctx context.Context, symbol CryptoSymbol, startTime time.Time) (*CryptoPriceResponse, error) {
	start, end := get15MinWindow(startTime)
	return c.GetCryptoPrice(ctx, symbol, start, end)
}

func get15MinWindow(t time.Time) (startTime, endTime time.Time) {
	t = t.UTC()
	minute := t.Minute()
	roundedMinute := (minute / 15) * 15
	startTime = time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), roundedMinute, 0, 0, time.UTC)
	endTime = startTime.Add(15 * time.Minute)
	return startTime, endTime
}

func getCryptoPrice(ctx context.Context, httpClient *http.Client, query cryptoPriceQuery) (*CryptoPriceResponse, error) {
	if query.Symbol == "" {
		return nil, fmt.Errorf("symbol is required")
	}

	if query.Variant == "" {
		query.Variant = cryptoPriceVariantFifteen
	}

	params := url.Values{}
	params.Set("symbol", string(query.Symbol))
	params.Set("eventStartTime", query.EventStartTime.UTC().Format("2006-01-02T15:04:05.000Z"))
	params.Set("endDate", query.EndDate.UTC().Format("2006-01-02T15:04:05.000Z"))
	params.Set("variant", string(query.Variant))

	fullURL := fmt.Sprintf("%s?%s", constants.CryptoPriceAPIURL, params.Encode())

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fullURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Accept", "application/json")

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status %d", resp.StatusCode)
	}

	var result CryptoPriceResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}
