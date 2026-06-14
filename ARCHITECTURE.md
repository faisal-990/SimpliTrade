# SimpliTrade — Architecture

A risk-free trading simulator: trade virtual money on **real-time market data**,
following the principles of famous investors (Graham, Buffett, Wood, …) encoded
as strategies. Built so a future real↔fake-money toggle is a config flip, not a
rewrite.

## Twin Towers

Two independent processes share one Postgres database. They never call each
other — they coordinate through the DB, so either can restart independently.

```
        FRONTEND (React + Vite + Tailwind + shadcn)
                      │ REST/JSON (Bearer JWT)
                      ▼
┌──────────── TOWER 1 · API SERVER (Gin) ────────────┐
│  Router → Middlewares (CORS · Logger · Auth)        │
│  Controllers → Services → Repositories → GORM       │
│                    │                                │
│                    └── Broker iface (Sim | Live) ◄─ real/fake toggle
└───────────────┬───────────────────────┬────────────┘
                │ reads                  │ reads/writes
                ▼                        ▼
            ┌────────────────── PostgreSQL ──────────────────┐
            │ users · accounts · refresh_tokens · investors · │
            │ follows · stocks · stock_prices · trades ·      │
            │ holdings · performance                          │
            └───────────────────────────▲────────────────────┘
                                         │ writes prices + bot trades
┌──────────── TOWER 2 · ENGINE / MARKET DAEMON ───────┐
│  Price poller → Strategy engine → Bot investors      │
│         │ pulls real-time                            │
└─────────┼────────────────────────────────────────────┘
          ▼
   EXTERNAL MARKET DATA API (real quotes + fundamentals)
```

**Only Tower 2 calls the external API.** Tower 1 always reads our DB, so user
traffic never consumes the (free-tier) API quota.

## Layering (Tower 1)

Strict, one-directional: `controller → service → repository → db`.
- **Controllers** parse/validate DTOs, pull identity from the JWT context, and
  render responses via `internal/web/httpx` (one success + one error envelope).
- **Services** hold business rules and orchestrate transactions; depend on repo
  *interfaces* (mockable).
- **Repositories** are pure GORM queries, no business logic.

Interface seams where the product is expected to change:
`Broker` (sim↔live money), `MarketData` (provider swap), `Strategy` (add
investors). Each is an interface from day one.

## The decision engine (Tower 2)

The engine is a **pure decision function** — it applies an investor's
already-encoded principles to the current world; it does not invent strategy:

```
Decide(market snapshot, strategy, portfolio) → []Intent   (pure, no side effects)
        Executor → Broker turns an Intent into a Trade     (side effects here)
```

Three inputs: **real-time prices/fundamentals**, the **selected strategy**
(a YAML in `internal/engine/strategies/`), and the bot's **portfolio status**
(cash + holdings, for sizing and rebalancing).

### Two clocks
Inputs change at different rates, so the engine runs two schedules:
- **Slow lane** (pre-market / nightly): pull expensive *fundamentals*,
  precompute price-independent thresholds (Graham number, intrinsic value, buy
  price), persist to DB + cache. This is the rate-limited work, done off-peak.
- **Fast lane** (market hours): poll cheap *quotes*, compare live price to the
  precomputed thresholds (the 100-goroutine fan-out), produce intents, execute.

On restart mid-day the fast lane rehydrates thresholds + portfolios from the DB
— no expensive refetch. This split is also the free-API quota strategy.

## The real↔fake money seam

`Account.Mode` (`sim` | `live`) selects a `Broker`:
- `SimulatedBroker` (built first): fills at DB price; in one transaction debits
  `Account.Balance`, upserts `Holding`, inserts `Trade`.
- `LiveBroker` (future): wraps a real brokerage behind the same interface.

`brokerFor(account.Mode)` is the only thing the toggle touches — controllers and
business rules are unaffected. Users and engine bots trade through the **same**
Broker path.

## Strategies as data

Investors are data, not code. Each `internal/engine/strategies/<name>.yml`
(schema v2, see `strategies/SCHEMA.md`) declares a `style` that routes to an
evaluator, plus optional metric blocks (valuation/quality/growth/technical/
momentum/allocation) and a user-facing `profile` (risk 1–10). Adding the 21st
investor = dropping one file.

See `BACKEND_PLAN.md` for the build roadmap and `strategies/SCHEMA.md` for the
strategy schema.
