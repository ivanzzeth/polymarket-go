# polymarket-go Development Decision

## Final Decision

Focus on **Polymarket-specific features** only, delegate generic quant features to BBGO framework.

## Core Modules

### 1. Smart Order Conversion (✅ Completed)
- `GetComplementaryTokenID` - Get complementary token ID
- `ConvertLimitOrder` - Convert limit orders
- `ConvertMarketOrder` - Convert market orders
- `GetConditionIDByTokenID` - Query condition ID
- **Value**: Solve position sync delay, enable immediate take-profit

### 2. Auto Management (✅ Completed)
- `WithAutoRedeem` - Auto-redeem resolved positions
- `WithAutoMerge` - Auto-merge complementary tokens (YES + NO → USDC)
- `Close()` - Graceful shutdown
- **Value**: Hands-free position management

### 3. Arbitrage Scanner (Planned)
- Dutch-book arbitrage detection
- Cross-market opportunities
- Gas cost optimization
- **Value**: Unique to Polymarket's complementary token mechanism

### 4. NegRisk Handler (Planned)
- NegRisk market detection
- Special redemption logic
- Risk assessment
- **Value**: Handle Polymarket's unique market type

### 5. BBGO Integration (Planned)
- Implement Exchange interface
- Implement Stream interface
- Type conversion layer
- **Value**: Reuse mature quant ecosystem

## What We Don't Build

Delegate to BBGO:
- ❌ Analytics - Performance analysis (win rate, profit/loss ratio, returns)
- ❌ RiskManager - Risk management
- ❌ Monitor - Real-time monitoring

**Reason**: BBGO already provides complete quant infrastructure for these generic features, avoid reinventing the wheel.

## Key Principles

1. **Focus on differentiation**: Build only Polymarket-specific features
2. **Reuse ecosystem**: Integrate with BBGO for generic features
3. **Production quality**: All code must be production-ready
4. **Simple is better**: Avoid over-engineering

## Implementation Status

### Phase 1: Smart Order Conversion (✅ Completed)
- Order conversion tools
- Auto-management features
- Examples and documentation
- Coding standards

### Phase 2: Arbitrage Scanner (Next)
- Architecture design
- Dutch-book detection
- Auto execution
- Gas optimization

### Phase 3: BBGO Integration (Future)
- Exchange interface
- Stream interface
- Strategy examples

## Success Criteria

- Smart order conversion accuracy > 95%
- Average cost savings > 15%
- Order optimization < 100ms
- Clear documentation with runnable examples
