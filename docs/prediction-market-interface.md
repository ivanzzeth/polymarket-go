# Generic Prediction Market Interface Design

## Overview

本文档定义了一个通用的二元市场（预测市场）交易接口抽象，使得：
1. **polymarket-go** 可以作为该接口的实现
2. 未来可以支持其他预测市场平台（Kalshi, PredictIt, Metaculus 等）
3. 与 **BBGO** 框架集成时有统一的适配层

**设计原则**：
- 抽象预测市场的核心概念，而非传统交易所概念
- 参考 BBGO 的 Exchange 接口设计，但针对预测市场特性调整
- 最小化必需接口，可选接口按需实现

---

## Core Domain Concepts

### 预测市场 vs 传统交易所

| 概念 | 传统交易所 | 预测市场 | 说明 |
|------|-----------|---------|------|
| 交易对象 | Currency Pair (BTC/USDT) | Outcome (YES/NO) | 预测市场交易的是事件结果 |
| 价格含义 | 资产价值 | 概率 [0, 1] | 价格代表结果发生的概率 |
| 结算方式 | 无到期 | 事件解决后结算 | 预测市场有明确的结束时间 |
| 互补关系 | 无 | YES + NO = 1 | 二元市场的核心特性 |
| 底层资产 | 加密货币 | Conditional Token | 基于 CTF 的条件代币 |

### 核心实体

```
┌─────────────────────────────────────────────────────────────────────┐
│                        Prediction Market                             │
├─────────────────────────────────────────────────────────────────────┤
│                                                                      │
│  ┌─────────────┐                                                    │
│  │   Market    │  A prediction market question                      │
│  │  (Event)    │  e.g., "Will X happen by date Y?"                  │
│  └──────┬──────┘                                                    │
│         │                                                            │
│         │ has multiple                                               │
│         ▼                                                            │
│  ┌─────────────┐                                                    │
│  │  Outcome    │  Possible results (YES/NO or multiple choices)     │
│  │  (Token)    │  Each outcome has a tradeable token                │
│  └──────┬──────┘                                                    │
│         │                                                            │
│         │ traded via                                                 │
│         ▼                                                            │
│  ┌─────────────┐     ┌─────────────┐     ┌─────────────┐           │
│  │  OrderBook  │     │   Order     │     │   Trade     │           │
│  │  (per token)│     │  (buy/sell) │     │  (executed) │           │
│  └─────────────┘     └─────────────┘     └─────────────┘           │
│                                                                      │
└─────────────────────────────────────────────────────────────────────┘
```

---

## Interface Design

### Core Interfaces (Required)

```go
// =============================================================================
// Exchange Interface - Main entry point
// =============================================================================

// BinaryPredictionExchange represents a binary prediction market exchange
// This is the main interface that platform implementations must satisfy
// "Binary" refers to markets with two mutually exclusive outcomes (YES/NO)
type BinaryPredictionExchange interface {
    // Identity
    Name() ExchangeName

    // Market Discovery
    QueryMarkets(ctx context.Context, filter *MarketFilter) ([]Market, error)
    QueryMarket(ctx context.Context, marketID string) (*Market, error)

    // Outcome/Token Discovery
    QueryOutcomes(ctx context.Context, marketID string) ([]Outcome, error)

    // Order Book
    QueryOrderBook(ctx context.Context, outcomeID string, depth int) (*OrderBook, error)

    // Order Management
    SubmitOrder(ctx context.Context, order SubmitOrder) (*Order, error)
    CancelOrder(ctx context.Context, orderID string) error
    QueryOpenOrders(ctx context.Context, filter *OrderFilter) ([]Order, error)
    QueryOrder(ctx context.Context, orderID string) (*Order, error)

    // Account
    QueryCollateralBalance(ctx context.Context) (*BalanceDetail, error)
    QueryPositionBalance(ctx context.Context, outcomeID string) (*BalanceDetail, error)
    QueryPositions(ctx context.Context) ([]Position, error)

    // Real-time Data
    NewStream() Stream
}

// =============================================================================
// Stream Interface - Real-time data subscription
// =============================================================================

// Stream provides real-time market data subscriptions
type Stream interface {
    // Lifecycle
    Connect(ctx context.Context) error
    Close() error

    // Subscriptions
    SubscribeOrderBook(outcomeIDs []string, callback func(OrderBookUpdate)) error
    SubscribeTrades(outcomeIDs []string, callback func(Trade)) error
    SubscribeOrders(callback func(OrderUpdate)) error
    SubscribeBalances(callback func(BalanceUpdate)) error

    // Connection events
    OnConnect(callback func())
    OnDisconnect(callback func(error))
}
```

### Optional Interfaces

```go
// =============================================================================
// Optional Interfaces - Implement as needed
// =============================================================================

// ComplementaryTokenSupport - For platforms with complementary token mechanics
// (e.g., Polymarket where YES + NO = 1)
type ComplementaryTokenSupport interface {
    // Get the complementary outcome for a given outcome
    // e.g., YES → NO, NO → YES
    GetComplementaryOutcome(ctx context.Context, outcomeID string) (*Outcome, error)

    // Convert order to complementary side
    // Buy YES @ 0.6 → Sell NO @ 0.4
    ConvertToComplementaryOrder(order *SubmitOrder) (*SubmitOrder, error)
}

// TokenOperations - For platforms supporting token minting/burning
type TokenOperations interface {
    // Split collateral into outcome tokens
    // 1 USDC → 1 YES + 1 NO
    Split(ctx context.Context, marketID string, amount decimal.Decimal) (txHash string, err error)

    // Merge outcome tokens back to collateral
    // 1 YES + 1 NO → 1 USDC
    Merge(ctx context.Context, marketID string, amount decimal.Decimal) (txHash string, err error)

    // Redeem winning tokens after market resolution
    Redeem(ctx context.Context, marketID string) (txHash string, err error)
}

// HistoricalData - For platforms providing historical data
type HistoricalData interface {
    // Query historical trades
    QueryTrades(ctx context.Context, outcomeID string, opts *TradeQueryOptions) ([]Trade, error)

    // Query historical OHLCV (if available)
    QueryKLines(ctx context.Context, outcomeID string, interval Interval, start, end time.Time) ([]KLine, error)
}

// MultiOutcomeSupport - For markets with more than 2 outcomes
type MultiOutcomeSupport interface {
    // Check if market is multi-outcome (not binary)
    IsMultiOutcome(marketID string) bool

    // Get all outcome probabilities (should sum to ~1.0)
    GetOutcomeProbabilities(ctx context.Context, marketID string) (map[string]decimal.Decimal, error)
}

// OnChainVerification - For blockchain-based platforms
type OnChainVerification interface {
    // Query balance directly from blockchain (source of truth)
    QueryCollateralBalanceOnChain(ctx context.Context, blockNumber *big.Int) (decimal.Decimal, error)
    QueryPositionBalanceOnChain(ctx context.Context, outcomeID string, blockNumber *big.Int) (decimal.Decimal, error)
}
```

---

## Data Types

### Market & Outcome Types

```go
// =============================================================================
// Market Types
// =============================================================================

type ExchangeName string

const (
    ExchangePolymarket ExchangeName = "polymarket"
    ExchangeKalshi     ExchangeName = "kalshi"
    ExchangePredictIt  ExchangeName = "predictit"
    ExchangeMetaculus  ExchangeName = "metaculus"
)

// MarketStatus represents the current state of a market
type MarketStatus string

const (
    MarketStatusOpen     MarketStatus = "open"      // Trading active
    MarketStatusClosed   MarketStatus = "closed"    // Trading stopped, awaiting resolution
    MarketStatusResolved MarketStatus = "resolved"  // Outcome determined
    MarketStatusVoided   MarketStatus = "voided"    // Market cancelled
)

// Market represents a prediction market (event/question)
type Market struct {
    ID            string
    Slug          string           // Human-readable identifier (e.g., "will-btc-reach-100k")
    Question      string           // The prediction question
    Description   string           // Detailed description
    Category      string           // Category (politics, sports, crypto, etc.)

    Status        MarketStatus
    CreatedAt     time.Time
    EndDate       *time.Time       // When trading ends
    ResolutionDate *time.Time      // When market was/will be resolved

    // Outcome information
    Outcomes      []Outcome        // All possible outcomes
    OutcomeType   OutcomeType      // Binary, MultiChoice, Scalar

    // Trading info
    Volume24h     decimal.Decimal  // 24h trading volume
    Liquidity     decimal.Decimal  // Total liquidity

    // Platform-specific metadata
    Metadata      map[string]interface{}
}

// OutcomeType indicates the type of outcomes in a market
type OutcomeType string

const (
    OutcomeTypeBinary      OutcomeType = "binary"       // YES/NO
    OutcomeTypeMultiChoice OutcomeType = "multi_choice" // Multiple exclusive outcomes
    OutcomeTypeScalar      OutcomeType = "scalar"       // Numeric range
)

// Outcome represents a tradeable outcome in a market
type Outcome struct {
    ID              string
    MarketID        string
    Name            string           // e.g., "Yes", "No", "Trump", "Biden"

    // Current pricing
    Price           decimal.Decimal  // Current probability/price [0, 1]

    // For binary markets: reference to the complementary outcome
    ComplementaryID *string          // nil for non-binary markets

    // Platform-specific token information
    TokenID         string           // Platform's internal token identifier

    // Resolution
    IsWinner        *bool            // nil if not resolved, true/false after resolution

    Metadata        map[string]interface{}
}

// MarketFilter for querying markets
type MarketFilter struct {
    Status     *MarketStatus
    Category   *string
    SearchTerm *string
    Limit      int
    Offset     int
}
```

### Order Types

```go
// =============================================================================
// Order Types
// =============================================================================

type OrderSide string

const (
    OrderSideBuy  OrderSide = "buy"
    OrderSideSell OrderSide = "sell"
)

type OrderType string

const (
    OrderTypeLimit  OrderType = "limit"
    OrderTypeMarket OrderType = "market"
)

type OrderStatus string

const (
    OrderStatusPending         OrderStatus = "pending"
    OrderStatusOpen            OrderStatus = "open"
    OrderStatusPartiallyFilled OrderStatus = "partially_filled"
    OrderStatusFilled          OrderStatus = "filled"
    OrderStatusCancelled       OrderStatus = "cancelled"
    OrderStatusRejected        OrderStatus = "rejected"
)

// SubmitOrder represents an order to be submitted
type SubmitOrder struct {
    OutcomeID  string           // Which outcome to trade
    Side       OrderSide        // Buy or Sell
    Type       OrderType        // Limit or Market

    // For limit orders
    Price      decimal.Decimal  // Price (probability) [0, 1]
    Size       decimal.Decimal  // Number of shares/contracts

    // For market orders
    Amount     *decimal.Decimal // Collateral amount to spend (for market buy)

    // Optional
    TimeInForce *TimeInForce
    Expiration  *time.Time
    ClientID    *string          // Client-provided order ID
}

type TimeInForce string

const (
    TimeInForceGTC TimeInForce = "gtc" // Good Till Cancelled
    TimeInForceIOC TimeInForce = "ioc" // Immediate Or Cancel
    TimeInForceFOK TimeInForce = "fok" // Fill Or Kill
)

// Order represents an order (submitted or historical)
type Order struct {
    ID          string
    ClientID    *string
    OutcomeID   string
    MarketID    string

    Side        OrderSide
    Type        OrderType
    Status      OrderStatus

    Price       decimal.Decimal
    Size        decimal.Decimal  // Original size
    FilledSize  decimal.Decimal  // Amount filled

    CreatedAt   time.Time
    UpdatedAt   time.Time

    // Execution info
    AvgFillPrice *decimal.Decimal // Average fill price (if partially/fully filled)
    Fee          decimal.Decimal  // Total fees paid
}

func (o *Order) RemainingSize() decimal.Decimal {
    return o.Size.Sub(o.FilledSize)
}

func (o *Order) IsDone() bool {
    return o.Status == OrderStatusFilled ||
           o.Status == OrderStatusCancelled ||
           o.Status == OrderStatusRejected
}

// OrderFilter for querying orders
type OrderFilter struct {
    OutcomeID *string
    MarketID  *string
    Status    *OrderStatus
    Side      *OrderSide
    Limit     int
}
```

### Order Book & Trade Types

```go
// =============================================================================
// Order Book Types
// =============================================================================

type PriceLevel struct {
    Price decimal.Decimal
    Size  decimal.Decimal
}

type OrderBook struct {
    OutcomeID  string
    Bids       []PriceLevel  // Sorted descending by price
    Asks       []PriceLevel  // Sorted ascending by price

    BestBid    *decimal.Decimal
    BestAsk    *decimal.Decimal
    Spread     *decimal.Decimal
    MidPrice   *decimal.Decimal

    Timestamp  time.Time
    Sequence   uint64
}

type OrderBookUpdate struct {
    OutcomeID string
    Bids      []PriceLevel  // Updated bid levels (Size=0 means remove)
    Asks      []PriceLevel  // Updated ask levels (Size=0 means remove)
    Timestamp time.Time
    Sequence  uint64
}

// =============================================================================
// Trade Types
// =============================================================================

type Trade struct {
    ID        string
    OutcomeID string
    MarketID  string

    Side      OrderSide        // Taker side
    Price     decimal.Decimal
    Size      decimal.Decimal

    MakerOrderID *string
    TakerOrderID *string

    Timestamp time.Time
}

type TradeQueryOptions struct {
    StartTime *time.Time
    EndTime   *time.Time
    Limit     int
}
```

### Balance & Position Types

```go
// =============================================================================
// Balance & Position Types
// =============================================================================

// BalanceDetail represents balance information with locked amounts
// Reuses existing type from polymarket-go
type BalanceDetail struct {
    TotalBalance     decimal.Decimal
    LockedBalance    decimal.Decimal  // Locked in open orders
    AvailableBalance decimal.Decimal  // Available for trading
}

// Position represents a position in an outcome
type Position struct {
    OutcomeID    string
    MarketID     string
    OutcomeName  string

    Size         decimal.Decimal  // Number of shares held
    AvgCost      decimal.Decimal  // Average cost basis
    CurrentPrice decimal.Decimal  // Current market price

    UnrealizedPnL decimal.Decimal // Unrealized profit/loss
    RealizedPnL   decimal.Decimal // Realized profit/loss
}

// BalanceUpdate for real-time balance changes
type BalanceUpdate struct {
    Type      BalanceUpdateType
    OutcomeID *string          // nil for collateral updates
    Balance   *BalanceDetail
    Timestamp time.Time
}

type BalanceUpdateType string

const (
    BalanceUpdateTypeCollateral BalanceUpdateType = "collateral"
    BalanceUpdateTypePosition   BalanceUpdateType = "position"
)

// OrderUpdate for real-time order changes
type OrderUpdate struct {
    Order     *Order
    Timestamp time.Time
}
```

---

## Platform Mappings

### Polymarket Mapping

```go
// Polymarket-specific mappings
//
// Polymarket Concept     → Generic Interface
// ────────────────────────────────────────────
// Market (Gamma API)     → Market
// Token (YES/NO)         → Outcome
// TokenID                → Outcome.TokenID
// ConditionID            → Market.ID (internal)
// USDC                   → Collateral
// Conditional Token      → Position Token
// CTF Split/Merge        → TokenOperations interface

type PolymarketExchange struct {
    client *polymarket.Client
}

func (e *PolymarketExchange) Name() ExchangeName {
    return ExchangePolymarket
}

// QueryMarkets maps to GammaClient.GetMarkets()
func (e *PolymarketExchange) QueryMarkets(ctx context.Context, filter *MarketFilter) ([]Market, error) {
    // Convert filter to Polymarket API params
    polymarketMarkets, err := e.client.GammaClient().GetMarkets(ctx, convertFilter(filter))
    if err != nil {
        return nil, err
    }

    // Convert to generic Market type
    return convertPolymarketMarkets(polymarketMarkets), nil
}

// QueryOrderBook maps to DataClient.GetOrderBook()
func (e *PolymarketExchange) QueryOrderBook(ctx context.Context, outcomeID string, depth int) (*OrderBook, error) {
    // outcomeID is the tokenID in Polymarket
    polymarketOB, err := e.client.DataClient().GetOrderBook(outcomeID)
    if err != nil {
        return nil, err
    }

    return convertOrderBook(polymarketOB), nil
}

// SubmitOrder maps to ClobClient.PostOrder()
func (e *PolymarketExchange) SubmitOrder(ctx context.Context, order SubmitOrder) (*Order, error) {
    // Convert to Polymarket order format
    polymarketOrder := convertToPolymarketOrder(order)

    result, err := e.client.ClobClient().PostOrder(polymarketOrder)
    if err != nil {
        return nil, err
    }

    return convertPolymarketOrder(result), nil
}

// Implement ComplementaryTokenSupport
func (e *PolymarketExchange) GetComplementaryOutcome(ctx context.Context, outcomeID string) (*Outcome, error) {
    complementaryTokenID, err := e.client.GetComplementaryTokenID(ctx, outcomeID)
    if err != nil {
        return nil, err
    }

    return &Outcome{
        ID:      complementaryTokenID,
        TokenID: complementaryTokenID,
    }, nil
}

// Implement TokenOperations
func (e *PolymarketExchange) Split(ctx context.Context, marketID string, amount decimal.Decimal) (string, error) {
    txHash, err := e.client.Split(ctx, marketID, amount)
    return txHash.Hex(), err
}

func (e *PolymarketExchange) Merge(ctx context.Context, marketID string, amount decimal.Decimal) (string, error) {
    txHash, err := e.client.Merge(ctx, marketID, amount)
    return txHash.Hex(), err
}

func (e *PolymarketExchange) Redeem(ctx context.Context, marketID string) (string, error) {
    txHash, err := e.client.Redeem(ctx, marketID)
    return txHash.Hex(), err
}

// Implement OnChainVerification
func (e *PolymarketExchange) QueryCollateralBalanceOnChain(ctx context.Context, blockNumber *big.Int) (decimal.Decimal, error) {
    return e.client.GetCollateralBalance(ctx, &polymarket.BalanceQueryOption{
        Source:      polymarket.DataSourceOnChain,
        BlockNumber: blockNumber,
    })
}
```

### Kalshi Mapping (Future)

```go
// Kalshi-specific mappings (conceptual)
//
// Kalshi Concept         → Generic Interface
// ────────────────────────────────────────────
// Event                  → Market
// Contract               → Outcome (binary: Yes/No)
// Contract Ticker        → Outcome.ID
// USD                    → Collateral
// Position               → Position

type KalshiExchange struct {
    // Kalshi API client
}

func (e *KalshiExchange) Name() ExchangeName {
    return ExchangeKalshi
}

// Note: Kalshi does NOT have complementary token mechanics
// Each position is independent, no Split/Merge operations
// Does NOT implement ComplementaryTokenSupport or TokenOperations
```

### PredictIt Mapping (Future)

```go
// PredictIt-specific mappings (conceptual)
//
// PredictIt Concept      → Generic Interface
// ────────────────────────────────────────────
// Market                 → Market
// Contract               → Outcome
// Share                  → Position unit
// USD                    → Collateral
//
// Note: PredictIt has $850 position limit per contract
// This should be reflected in platform-specific constraints
```

---

## BBGO Integration Layer

```go
// =============================================================================
// BBGO Exchange Adapter
// =============================================================================

// BBGOPredictionAdapter adapts BinaryPredictionExchange to BBGO's Exchange interface
type BBGOPredictionAdapter struct {
    exchange BinaryPredictionExchange
}

// NewBBGOAdapter creates a new BBGO-compatible adapter
func NewBBGOAdapter(exchange BinaryPredictionExchange) *BBGOPredictionAdapter {
    return &BBGOPredictionAdapter{exchange: exchange}
}

// BBGO Required Methods

func (a *BBGOPredictionAdapter) Name() types.ExchangeName {
    return types.ExchangeName(a.exchange.Name())
}

func (a *BBGOPredictionAdapter) QueryMarkets(ctx context.Context) (types.MarketMap, error) {
    markets, err := a.exchange.QueryMarkets(ctx, nil)
    if err != nil {
        return nil, err
    }

    marketMap := make(types.MarketMap)
    for _, m := range markets {
        for _, outcome := range m.Outcomes {
            // Create a BBGO "symbol" for each outcome
            // Format: MARKET_SLUG-OUTCOME (e.g., "btc-100k-YES")
            symbol := fmt.Sprintf("%s-%s", m.Slug, outcome.Name)

            marketMap[symbol] = types.Market{
                Symbol:      symbol,
                BaseCurrency:  outcome.Name,    // YES/NO
                QuoteCurrency: "USD",           // Collateral

                // Prediction market specific
                MinPrice:    fixedpoint.NewFromFloat(0.01),  // Min probability
                MaxPrice:    fixedpoint.NewFromFloat(0.99),  // Max probability
                TickSize:    fixedpoint.NewFromFloat(0.01),  // Price tick

                MinQuantity: fixedpoint.NewFromFloat(1),     // Min shares
                StepSize:    fixedpoint.NewFromFloat(1),     // Share step
            }
        }
    }

    return marketMap, nil
}

func (a *BBGOPredictionAdapter) QueryTickers(ctx context.Context, symbols ...string) (map[string]types.Ticker, error) {
    tickers := make(map[string]types.Ticker)

    for _, symbol := range symbols {
        marketSlug, outcomeName := parseSymbol(symbol)

        market, err := a.exchange.QueryMarket(ctx, marketSlug)
        if err != nil {
            continue
        }

        for _, outcome := range market.Outcomes {
            if outcome.Name == outcomeName {
                ob, _ := a.exchange.QueryOrderBook(ctx, outcome.ID, 1)

                tickers[symbol] = types.Ticker{
                    Symbol: symbol,
                    Last:   fixedpoint.NewFromDecimal(outcome.Price),
                    Bid:    fixedpoint.NewFromDecimal(*ob.BestBid),
                    Ask:    fixedpoint.NewFromDecimal(*ob.BestAsk),
                    Volume: fixedpoint.NewFromDecimal(market.Volume24h),
                    Time:   time.Now(),
                }
                break
            }
        }
    }

    return tickers, nil
}

func (a *BBGOPredictionAdapter) SubmitOrder(ctx context.Context, order types.SubmitOrder) (*types.Order, error) {
    // Convert BBGO order to generic order
    genericOrder := SubmitOrder{
        OutcomeID: order.Symbol, // Symbol is outcomeID in our mapping
        Side:      convertSide(order.Side),
        Type:      convertType(order.Type),
        Price:     order.Price.Decimal(),
        Size:      order.Quantity.Decimal(),
    }

    result, err := a.exchange.SubmitOrder(ctx, genericOrder)
    if err != nil {
        return nil, err
    }

    return convertToBBGOOrder(result), nil
}

func (a *BBGOPredictionAdapter) QueryOpenOrders(ctx context.Context, symbol string) ([]types.Order, error) {
    orders, err := a.exchange.QueryOpenOrders(ctx, &OrderFilter{
        OutcomeID: &symbol,
    })
    if err != nil {
        return nil, err
    }

    bbgoOrders := make([]types.Order, len(orders))
    for i, o := range orders {
        bbgoOrders[i] = *convertToBBGOOrder(&o)
    }

    return bbgoOrders, nil
}

func (a *BBGOPredictionAdapter) CancelOrders(ctx context.Context, orders ...types.Order) error {
    for _, order := range orders {
        if err := a.exchange.CancelOrder(ctx, order.OrderID); err != nil {
            return err
        }
    }
    return nil
}

func (a *BBGOPredictionAdapter) NewStream() types.Stream {
    return &BBGOStreamAdapter{
        stream: a.exchange.NewStream(),
    }
}

// QueryKLines - Not supported for most prediction markets
func (a *BBGOPredictionAdapter) QueryKLines(ctx context.Context, symbol string,
    interval types.Interval, options types.KLineQueryOptions) ([]types.KLine, error) {

    // Check if exchange supports historical data
    if hist, ok := a.exchange.(HistoricalData); ok {
        klines, err := hist.QueryKLines(ctx, symbol,
            convertInterval(interval),
            options.StartTime,
            options.EndTime)
        if err != nil {
            return nil, err
        }
        return convertToBBGOKLines(klines), nil
    }

    return nil, fmt.Errorf("historical KLine data not supported by %s", a.exchange.Name())
}
```

---

## Usage Examples

### Basic Usage

```go
// Create exchange instance
exchange := polymarket.NewExchange(client)

// Query markets
markets, _ := exchange.QueryMarkets(ctx, &MarketFilter{
    Status:   ptr(MarketStatusOpen),
    Category: ptr("crypto"),
    Limit:    10,
})

for _, market := range markets {
    fmt.Printf("Market: %s\n", market.Question)
    for _, outcome := range market.Outcomes {
        fmt.Printf("  %s: %.2f\n", outcome.Name, outcome.Price)
    }
}

// Place an order
order, _ := exchange.SubmitOrder(ctx, SubmitOrder{
    OutcomeID: markets[0].Outcomes[0].ID,  // Buy YES
    Side:      OrderSideBuy,
    Type:      OrderTypeLimit,
    Price:     decimal.NewFromFloat(0.55),
    Size:      decimal.NewFromFloat(100),
})

fmt.Printf("Order placed: %s\n", order.ID)
```

### With BBGO Integration

```go
// Wrap in BBGO adapter
bbgoExchange := NewBBGOAdapter(exchange)

// Use with BBGO strategies
session := &bbgo.ExchangeSession{
    Exchange: bbgoExchange,
    // ...
}

// Run BBGO strategy
gridStrategy := &grid.Strategy{
    Symbol:        "btc-100k-YES",
    UpperPrice:    fixedpoint.NewFromFloat(0.70),
    LowerPrice:    fixedpoint.NewFromFloat(0.30),
    GridNum:       10,
    Quantity:      fixedpoint.NewFromFloat(10),
}

gridStrategy.Run(ctx, session)
```

### Real-time Streaming

```go
stream := exchange.NewStream()

// Subscribe to order book updates
stream.SubscribeOrderBook([]string{outcomeID}, func(update OrderBookUpdate) {
    fmt.Printf("OrderBook update: bid=%s ask=%s\n",
        update.Bids[0].Price, update.Asks[0].Price)
})

// Subscribe to own orders
stream.SubscribeOrders(func(update OrderUpdate) {
    fmt.Printf("Order %s: %s\n", update.Order.ID, update.Order.Status)
})

// Connect
stream.Connect(ctx)
```

### Cross-Platform Arbitrage (Future)

```go
// Create multiple exchange instances
polymarket := polymarket.NewExchange(polyClient)
kalshi := kalshi.NewExchange(kalshiClient)

// Find same event on both platforms
polyMarket, _ := polymarket.QueryMarket(ctx, "btc-100k-2024")
kalshiMarket, _ := kalshi.QueryMarket(ctx, "btc-100k-2024")

// Compare prices
polyYesPrice := polyMarket.Outcomes[0].Price   // YES on Polymarket
kalshiYesPrice := kalshiMarket.Outcomes[0].Price // YES on Kalshi

if polyYesPrice.Sub(kalshiYesPrice).Abs().GreaterThan(decimal.NewFromFloat(0.05)) {
    fmt.Println("Arbitrage opportunity detected!")
    // Execute cross-platform arbitrage
}
```

---

## Platform Feature Matrix

| Feature | Polymarket | Kalshi | PredictIt |
|---------|------------|--------|-----------|
| **Core Trading** |
| QueryMarkets | ✅ | ✅ | ✅ |
| QueryOrderBook | ✅ | ✅ | ✅ |
| SubmitOrder | ✅ | ✅ | ✅ |
| CancelOrder | ✅ | ✅ | ✅ |
| **Real-time** |
| WebSocket Stream | ✅ | ✅ | ❌ |
| Order Updates | ✅ | ✅ | ❌ |
| **Special Features** |
| ComplementaryTokenSupport | ✅ | ❌ | ❌ |
| TokenOperations (Split/Merge) | ✅ | ❌ | ❌ |
| OnChainVerification | ✅ | ❌ | ❌ |
| **Historical Data** |
| QueryTrades | ✅ | ✅ | ✅ |
| QueryKLines (Candlesticks) | ✅ | ✅ | ❌ |
| **Constraints** |
| Position Limit | None | $25K/market | $850/contract |
| Min Order | ~$1 | $1 | $1 |
| Fees | 0% (gas only) | 0-7% | 10% profit fee |

---

## Polymarket Implementation Status

> **Status: ✅ READY** - 所有底层能力已实现，可以开始构建适配器层

### 接口实现映射详表

#### BinaryPredictionExchange (Core Interface)

| 接口方法 | polymarket-go 实现 | 说明 |
|---------|-------------------|------|
| `Name()` | 返回 `"polymarket"` | 常量 |
| `QueryMarkets()` | `GammaClient().GetMarkets(ctx, params)` | 支持过滤、分页 |
| `QueryMarket()` | `GammaClient().GetMarketByID(ctx, marketID, params)` | 按 ID 查询 |
| `QueryOutcomes()` | `GammaClient().GetMarketByID()` 返回的 `Market.Tokens` | Token 即 Outcome |
| `QueryOrderBook()` | `ClobClient().GetOrderBook(tokenID)` | 按 TokenID 查询 |
| `SubmitOrder()` | `ClobClient().CreateAndPostOrder(order, options, orderType)` | 限价单 |
| `SubmitOrder()` (market) | `ClobClient().CreateAndPostMarketOrder(order, options)` | 市价单 |
| `CancelOrder()` | `ClobClient().CancelOrder(orderID)` | 单个取消 |
| `CancelOrders()` | `ClobClient().CancelOrders(orderIDs)` | 批量取消 |
| `QueryOpenOrders()` | `ClobClient().GetOpenOrders(params, onlyFirstPage, cursor)` | 支持分页 |
| `QueryOrder()` | `ClobClient().GetOrder(orderID)` | 按 ID 查询 |
| `QueryCollateralBalance()` | `GetCollateralBalanceDetail(ctx)` | 返回 BalanceDetail |
| `QueryPositionBalance()` | `GetPositionBalanceDetail(ctx, tokenID)` | 返回 BalanceDetail |
| `QueryPositions()` | 需封装：遍历持仓 Token 调用 `GetPositionBalance` | - |
| `NewStream()` | `RealtimeDataClient()` | WebSocket 客户端 |

#### Stream Interface

| 接口方法 | polymarket-go 实现 | 说明 |
|---------|-------------------|------|
| `Connect()` | `RealtimeDataClient().Connect(ctx)` | 建立 WebSocket 连接 |
| `Close()` | `RealtimeDataClient().Close()` | 关闭连接 |
| `SubscribeOrderBook()` | `RealtimeDataClient().SubscribeBook(tokenID, callback)` | 订阅 Order Book |
| `SubscribeTrades()` | `RealtimeDataClient().SubscribeLastTradePrice(tokenID, callback)` | 订阅成交 |
| `SubscribeOrders()` | `RealtimeDataClient().SubscribeUser(callback)` | 订阅用户订单 |
| `OnConnect()` | `RealtimeDataClient().OnConnect(callback)` | 连接事件 |
| `OnDisconnect()` | `RealtimeDataClient().OnDisconnect(callback)` | 断连事件 |

#### ComplementaryTokenSupport (Optional)

| 接口方法 | polymarket-go 实现 | 说明 |
|---------|-------------------|------|
| `GetComplementaryOutcome()` | `GetComplementaryTokenID(ctx, tokenID)` | 获取互补 Token |
| `ConvertToComplementaryOrder()` | `ConvertLimitOrderToComplementary(ctx, order)` | 限价单转换 |
| `ConvertToComplementaryOrder()` | `ConvertMarketOrderToComplementary(ctx, order)` | 市价单转换 |

#### TokenOperations (Optional)

| 接口方法 | polymarket-go 实现 | 说明 |
|---------|-------------------|------|
| `Split()` | `Split(ctx, conditionID, amount)` | USDC → YES + NO |
| `Merge()` | `Merge(ctx, conditionID, amount)` | YES + NO → USDC |
| `Redeem()` | `Redeem(ctx, conditionID)` | 赎回获胜代币 |
| `RedeemNegRisk()` | `RedeemNegRisk(ctx, conditionID, amounts)` | NegRisk 市场赎回 |

#### HistoricalData (Optional)

| 接口方法 | polymarket-go 实现 | 说明 |
|---------|-------------------|------|
| `QueryTrades()` | `ClobClient().GetTrades(params, onlyFirstPage, cursor)` | 历史成交查询 |
| `QueryKLines()` | `GetCandlesticks(ctx, tokenID, interval, start, end, fillGaps)` | **新实现** OHLCV K线 |

#### OnChainVerification (Optional)

| 接口方法 | polymarket-go 实现 | 说明 |
|---------|-------------------|------|
| `QueryCollateralBalanceOnChain()` | `GetCollateralBalance(ctx, &BalanceQueryOption{Source: DataSourceOnChain, BlockNumber: blockNum})` | 链上 USDC 余额 |
| `QueryPositionBalanceOnChain()` | `GetPositionBalance(ctx, tokenID, &BalanceQueryOption{Source: DataSourceOnChain, BlockNumber: blockNum})` | 链上持仓余额 |

### 附加能力

| 功能 | polymarket-go 实现 | 说明 |
|------|-------------------|------|
| 订单转换（对手方） | `ConvertLimitOrderToOppositeSide(order, spread)` | Buy YES → Sell YES |
| 订单转换（匹配同边） | `ConvertLimitOrderToMatchingSameSide(ctx, order, spread)` | 通过互补 Token 匹配 |
| 自动赎回 | `startAutoRedeem(ctx, config)` | 自动赎回已结算市场 |
| 自动合并 | `startAutoMerge(ctx, config)` | 自动合并 YES+NO |
| Crypto 价格查询 | `GetCryptoPrice(ctx, symbol, start, end)` | BTC/ETH 15分钟价格 |
| ConditionID 查询 | `GetConditionIDByTokenID(ctx, tokenID)` | Token → Condition 映射 |

---

## Implementation Roadmap

### Phase 1: Core Polymarket Implementation ✅ COMPLETED
- [x] 底层能力：`BinaryPredictionExchange` 接口所需全部方法
- [x] 底层能力：`Stream` 接口（RealtimeDataClient）
- [x] 底层能力：`ComplementaryTokenSupport` 接口
- [x] 底层能力：`TokenOperations` 接口
- [x] 底层能力：`OnChainVerification` 接口
- [x] 底层能力：`HistoricalData.QueryKLines` (GetCandlesticks)

### Phase 2: BBGO Integration (Next)
- [ ] 实现 `PolymarketExchange` 适配器（封装 Client）
- [ ] 实现 `BBGOPredictionAdapter`
- [ ] 实现 `BBGOStreamAdapter`
- [ ] 测试 BBGO grid strategy
- [ ] 测试 BBGO market making strategy

### Phase 3: Additional Platforms (Future)
- [ ] Kalshi implementation
- [ ] PredictIt implementation
- [ ] Cross-platform arbitrage tooling

---

## References

- [BBGO Exchange Interface](https://github.com/c9s/bbgo/blob/main/pkg/types/exchange.go)
- [BBGO Adding New Exchange](https://github.com/c9s/bbgo/blob/main/doc/development/adding-new-exchange.md)
- [Polymarket CLOB API](https://docs.polymarket.com/)
- [Kalshi API](https://trading-api.readme.io/reference/getting-started)
- [bbgo-integration.md](./bbgo-integration.md)
- [gocryptotrader-integration.md](./gocryptotrader-integration.md)

---

**Last Updated:** 2025-11-30
**Status:** Phase 1 Complete - Ready for BBGO Integration
