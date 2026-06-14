# Strategy Schema v2 — Multi-Paradigm Investor Profiles

Each file under `internal/engine/strategies/<investor>.yml` describes ONE real-world
investor as a deterministic strategy the engine can execute. The engine's `Decide()`
reads only the blocks relevant to the declared `identity.style` — a value investor
has no `technical` block, a momentum trader has no `valuation` block.

> v1 note: `benjamin.yml` and `loose.yml` are the legacy v1 fixtures the old daemon
> parsed. They are superseded by `graham.yml` (v2) and will be removed once the
> Phase-5 engine adopts this schema.

## Why blocks are optional
Every gate/score block is OPTIONAL. A strategy only declares the blocks its paradigm
uses. The engine resolves which evaluator to run from `identity.style`:

| style                     | gate blocks used                       | decision basis              |
|---------------------------|----------------------------------------|-----------------------------|
| deep_value / defensive    | valuation, financial_safety, stability | price vs intrinsic + safety |
| quality_value / quality_* | valuation, quality, stability          | moat + fair price           |
| garp                      | valuation(peg), growth, quality        | growth at reasonable price  |
| growth / disruptive_growth| growth, quality(margins)               | revenue growth + TAM        |
| quant_value (magic formula)| valuation(earnings_yield), quality(roic)| mechanical factor rank     |
| contrarian / distressed   | valuation, stability                   | max pessimism + MOS         |
| activist                  | valuation, quality                     | undervalued + catalyst      |
| macro_risk_parity         | allocation                             | asset-class targets         |
| macro_momentum            | momentum, allocation                   | top-down + trend            |
| trend_momentum            | technical, momentum                    | pure price action           |

## Canonical structure
```yaml
schema_version: 2

identity:
  id:        <kebab-slug>          # stable key
  name:      <Full Name>
  firm:      <firm / fund>
  era:       <active years>
  style:     <style enum above>
  philosophy:<one-line thesis>
  enabled:   true

profile:                           # USER-FACING — drives selection & pricing tier
  risk_rating:          <1..10>    # 1 = capital preservation, 10 = max aggression
  volatility:           low|medium|high|extreme
  time_horizon:         short|medium|long
  typical_holding_days: <int>
  tagline:              <quote>

universe:                          # what it is allowed to trade
  asset_classes:  [equity|etf|bond|commodity|crypto]
  market_cap_min: <num|null>       # USD
  market_cap_max: <num|null>
  sectors_include:[]               # empty = all
  sectors_exclude:[]
  geographies:    [US|GLOBAL|...]
  min_positions:  <int>
  max_positions:  <int>

buy:
  gates:                           # HARD pass/fail — must pass ALL present blocks
    valuation:        {pe_max, forward_pe_max, pb_max, ps_max, peg_max,
                       ev_ebitda_max, earnings_yield_min, fcf_yield_min,
                       dividend_yield_min, use_graham_number}
    financial_safety: {current_ratio_min, debt_to_equity_max, interest_coverage_min,
                       net_current_asset_rule, fcf_positive}
    quality:          {roe_min, roic_min, gross_margin_min, operating_margin_min,
                       net_margin_min, require_consistent_margins}
    growth:           {revenue_growth_yoy_min, eps_growth_yoy_min,
                       revenue_cagr_3y_min, eps_growth_5y_min}
    stability:        {eps_positive_years_min, dividend_years_min, beta_max}
    technical:        {price_above_sma_days, rsi_min, rsi_max,
                       near_52w_high_pct, breakout_required}
    momentum:         {return_3m_min, return_6m_min, return_12m_min,
                       relative_strength_min}
  margin_of_safety_min: <0..1|null># buy only if price <= intrinsic * (1 - mos)
  scoring:                         # weights (~sum 1.0) to RANK survivors
    <factor>: <weight>             # factor ∈ valuation|quality|growth|intrinsic|
                                   #          momentum|earnings_yield|stability

sell:
  take_profit_vs_intrinsic: <num|null>  # sell when price >= intrinsic * x
  stop_loss_pct:            <num|null>   # hard stop, e.g. 0.15 = -15%
  trailing_stop_pct:        <num|null>
  thesis_break:             {roe_below, debt_to_equity_above, eps_downtrend_years,
                             revenue_growth_below, dividend_cut, rsi_below}
  max_position_size:        <0..1>

risk:
  position_sizing: equal|conviction|volatility_parity|pyramid
  max_position_size: <0..1>
  cash_buffer_min:   <0..1>
  rebalance:         none|monthly|quarterly|annual|dynamic
  leverage_max:      <num>          # 1.0 = no leverage

allocation:                         # ONLY for macro_* styles (asset-class targets)
  equity: <0..1>
  bonds:  <0..1>
  gold:   <0..1>
  commodities: <0..1>
  cash:   <0..1>
```

## Metric vocabulary the engine must support
The provider/marketdata layer must supply these per symbol so gates can evaluate:
- **Valuation:** pe, forward_pe, pb, ps, peg, ev_ebitda, earnings_yield (EBIT/EV),
  fcf_yield, dividend_yield, eps_ttm, bvps
- **Quality:** roe, roic, gross_margin, operating_margin, net_margin, debt_to_equity,
  current_ratio, interest_coverage, fcf_positive
- **Growth:** revenue_growth_yoy, eps_growth_yoy, revenue_cagr_3y, eps_growth_5y
- **Stability:** eps_positive_years, dividend_years, beta
- **Technical/Momentum:** sma50, sma200, rsi14, return_1m/3m/6m/12m, high_52w,
  relative_strength, volatility_30d
- **Size:** market_cap, sector, geography

Anything a strategy references but the provider can't supply yet → that gate is
skipped with a logged warning (fail-open per-metric, never crash the engine).
