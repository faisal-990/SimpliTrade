# SimpliTrade — Backend Build Plan

Dependency-ordered. Each tier ships on its own feature branch → PR → review →
merge. **Done bar before any commit:** `go build`, `go vet`, `go test -race`
green (`make check`).

## Tenets
1. Strict layering `controller → service → repository → db`.
2. Interface seams where the product changes: `Broker`, `MarketData`, `Strategy`.
3. Pure `Decide()`; side effects only in Executor/Broker.
4. Twin Towers share only Postgres.
5. Money lives on `Account` (mode `sim`|`live`) so the real-money toggle is a data change.
6. **Free-API discipline:** only Tower 2 calls external APIs; cache + batch +
   rate-limit/backoff + circuit-breaker; fundamentals fetched once, quotes cheap.
7. Build against mocks/fakes; wire real Postgres + real provider near the end.

## Testing
- Pure logic (`Decide`, broker math, pricing) → table-driven unit tests.
- Repositories → integration vs ephemeral Postgres (testcontainers, T10).
- Services → mocked repos (`go.uber.org/mock`).
- Handlers/flows → `net/http/httptest`.

## Roadmap

| Tier | Deliverable | Gate |
|---|---|---|
| **T0 Foundations** ✅ | typed `config`; `httpx` error/response envelope; slog logger; `Account` + `RefreshToken` models (money moved off `User`); storage refactor (config-driven, returns errors); test+CI harness; repo hygiene | builds from config; `make check` green |
| **T1 Auth + Account** | signup (persist + bcrypt); login; JWT carries userID+accountID+role; refresh rotation; `/me`; reset; default sim account ($100k) on signup | signup→login→/me→refresh E2E |
| **T2 Market data** | `MarketData` interface + `FakeProvider`; Stock/StockPrice repos; seed ~100-symbol universe | fake provider populates DB |
| **T3 Broker + Trading** | `Broker` + `SimulatedBroker` (atomic txn); trade buy/sell/history; idempotency | math correct; funds/oversell rejected; race-safe |
| **T4 Portfolio** | holdings valuation; P&L; ROI; allocation | `/portfolio/stats` matches fixtures |
| **T5 Strategy engine** | v2 `StrategyConfig` + loader; `Strategy` evaluators; pure `Decide()`; pricing math; risk filter | table-driven Decide() per paradigm |
| **T6 Daemon/runner** | slow + fast lanes; seed 20 bot investors; Performance recompute; scheduler + graceful shutdown | bots trade on fake data; restart resumes from DB |
| **T7 Social** | leaderboard; investor profile; trade feed; follow/unfollow; aggregated feed | follow → see trades E2E |
| **T8 Dashboard** | fundamentals; candle graph; news (fix seed path) | endpoints return correct shapes |
| **T9 Real provider** | Finnhub/AlphaVantage behind `MarketData`; rate-limit, cache, retry; config select | integration test w/ real key; fake still works |
| **T10 Sellable** | golang-migrate + testcontainers; Stripe billing + entitlements; metrics/health; rate limits; OpenAPI; Docker/compose; `LiveBroker` behind `Account.Mode` + KYC gate | subscribe → tier gates a feature; metrics green |

## Status
- ✅ Strategies: v2 schema + 20 investor profiles (`internal/engine/strategies/`).
- ✅ T0 Foundations.
- ✅ T1 Auth + Account — signup/login/refresh-rotation/logout/me, JWT (userID+accountID+role), default sim account; E2E through the real stack with an in-memory repo.
- ▶ Next: T2 Market data (FakeProvider first).
