import { Link } from "react-router-dom";
import { PieChart, Pie, Cell, ResponsiveContainer, Tooltip } from "recharts";
import { Sparkles, Banknote, RotateCcw } from "lucide-react";
import {
  usePortfolioStats,
  useTradeHistory,
  useFollowing,
  useAllocations,
  useStopAllocation,
  useFeed,
  useSellAll,
  useResetAccount,
} from "@/hooks/queries";
import { Loading, ErrorState, EmptyState, Spinner } from "@/components/common/states";
import { Button } from "@/components/ui/button";
import { AllocationCard, type AggAlloc } from "@/components/portfolio/AllocationCard";
import { money, pct, pnlColor, qty, fromUnix } from "@/lib/format";
import { strategyLabel, initials } from "@/lib/investorMeta";
import { cn } from "@/lib/utils";
import type { HoldingDTO, Allocation, FeedItem } from "@/types/api";

const SLICE = ["var(--chart-1)", "var(--chart-2)", "var(--chart-4)", "var(--chart-5)", "var(--chart-3)"];

// Aggregate per-allocation rows into one card per investor: capital, cash and
// market value sum; return is value-weighted across the investor's slices.
function aggregateByInvestor(rows: Allocation[]): AggAlloc[] {
  const byId = new Map<string, AggAlloc>();
  for (const a of rows) {
    const cur = byId.get(a.investor_id);
    if (cur) {
      cur.capital += a.capital;
      cur.cash += a.cash;
      cur.market_value += a.market_value;
      cur.ids.push(a.id);
    } else {
      byId.set(a.investor_id, {
        investor_id: a.investor_id,
        investor_name: a.investor_name,
        strategy: a.strategy,
        capital: a.capital,
        cash: a.cash,
        market_value: a.market_value,
        return_pct: 0,
        ids: [a.id],
      });
    }
  }
  // value-weighted return as a FRACTION (the UI's pct() multiplies by 100).
  for (const v of byId.values()) {
    v.return_pct = v.capital > 0 ? v.market_value / v.capital - 1 : 0;
  }
  return [...byId.values()].sort((a, b) => b.market_value - a.market_value);
}

export default function Portfolio() {
  const stats = usePortfolioStats();
  const history = useTradeHistory();
  const following = useFollowing();
  const allocations = useAllocations();
  const feed = useFeed();
  const stop = useStopAllocation();
  const sellAll = useSellAll();
  const reset = useResetAccount();
  const activeAllocs = aggregateByInvestor((allocations.data ?? []).filter((a) => a.is_active));

  if (stats.isLoading) return <Loading label="Loading your portfolio…" />;
  if (stats.isError || !stats.data) return <ErrorState message={(stats.error as Error)?.message} />;

  const s = stats.data;
  const holdings = s.holdings ?? [];

  const stopInvestor = (a: AggAlloc) => a.ids.forEach((id) => stop.mutate(id));
  const onReset = () => {
    if (window.confirm("Reset your account? This liquidates every position and copy allocation and restores your starting balance.")) {
      reset.mutate();
    }
  };

  return (
    <div className="space-y-7">
      <header className="flex flex-wrap items-start justify-between gap-3">
        <div>
          <h1 className="text-3xl font-semibold tracking-tight">Portfolio</h1>
          <div className="hairline-gold mt-2 w-24" />
          <p className="mt-2 text-sm text-muted-foreground">Your simulated positions, valued at the latest prices.</p>
        </div>
        <div className="flex items-center gap-2">
          <Button
            variant="outline"
            size="sm"
            disabled={sellAll.isPending || !holdings.length}
            onClick={() => sellAll.mutate()}
          >
            {sellAll.isPending ? <Spinner /> : <Banknote className="h-4 w-4" />} Sell all
          </Button>
          <Button variant="ghost" size="sm" className="text-loss hover:text-loss" disabled={reset.isPending} onClick={onReset}>
            {reset.isPending ? <Spinner /> : <RotateCcw className="h-4 w-4" />} Reset
          </Button>
        </div>
      </header>

      {/* Stat cards */}
      <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
        <Stat label="Total value" value={money(s.total_value)} />
        <Stat label="Cash" value={money(s.cash)} />
        <Stat label="Unrealized P&L" value={money(s.unrealized_pl)} accent={pnlColor(s.unrealized_pl)} sub={pct(s.unrealized_pl_pct)} />
        <Stat label="ROI" value={pct(s.roi)} accent={pnlColor(s.roi)} />
      </div>

      {/* Active copy-trading allocations */}
      {activeAllocs.length > 0 && (
        <section className="rounded-xl border bg-card p-5">
          <h2 className="mb-3 text-sm font-semibold text-muted-foreground">Copy-trading allocations</h2>
          <div className="grid gap-3 sm:grid-cols-2 lg:grid-cols-3">
            {activeAllocs.map((a) => (
              <AllocationCard key={a.investor_id} a={a} stopping={stop.isPending} onStop={() => stopInvestor(a)} />
            ))}
          </div>
          <p className="mt-3 text-[11px] text-muted-foreground">Stopping liquidates the allocation at current prices and returns the cash to your balance. Multiple allocations to the same investor are summed here, and shown individually on the investor's page.</p>
        </section>
      )}

      {/* Investors you follow — entry point to capped copy-trading */}
      {!!following.data?.length && (
        <section className="rounded-xl border bg-card p-5">
          <div className="mb-3 flex items-center justify-between">
            <h2 className="text-sm font-semibold text-muted-foreground">Investors you follow</h2>
            <Link to="/app/investors" className="text-xs font-medium text-primary hover:underline">Browse all</Link>
          </div>
          <div className="grid gap-3 sm:grid-cols-2 lg:grid-cols-3">
            {following.data.map((inv) => (
              <div key={inv.id} className="flex items-center gap-3 rounded-xl border bg-background/50 p-3">
                <span className="flex h-9 w-9 items-center justify-center rounded-full bg-primary/12 text-xs font-semibold text-primary">
                  {initials(inv.name)}
                </span>
                <div className="min-w-0 flex-1">
                  <Link to={`/app/investors/${inv.id}`} className="block truncate text-sm font-medium hover:text-primary">{inv.name}</Link>
                  <p className="truncate text-xs text-muted-foreground">{strategyLabel(inv.strategy)}</p>
                </div>
                <span className={cn("text-sm font-semibold tabular-nums", pnlColor(inv.roi))}>{pct(inv.roi)}</span>
              </div>
            ))}
          </div>
          <div className="mt-4 flex items-center gap-2 rounded-lg bg-primary/8 px-3 py-2 text-xs text-muted-foreground">
            <Sparkles className="h-4 w-4 text-primary" />
            Open an investor to allocate a capped balance — their bot trades only that slice, then hit <span className="font-medium text-foreground">Simulate market</span> to watch it act.
            <Button asChild size="sm" variant="outline" className="ml-auto">
              <Link to="/app/investors">Choose an investor</Link>
            </Button>
          </div>
        </section>
      )}

      <div className="grid gap-7 lg:grid-cols-[1fr_1.6fr]">
        {/* Allocation */}
        <section className="rounded-xl border bg-card p-5">
          <h2 className="text-sm font-semibold text-muted-foreground">Allocation</h2>
          {holdings.length ? (
            <>
              <div className="mx-auto mt-2 h-52 w-52">
                <ResponsiveContainer width="100%" height="100%">
                  <PieChart>
                    <Pie data={holdings} dataKey="market_value" nameKey="symbol" innerRadius={58} outerRadius={88} paddingAngle={2} stroke="none">
                      {holdings.map((_, i) => (
                        <Cell key={i} fill={SLICE[i % SLICE.length]} />
                      ))}
                    </Pie>
                    <Tooltip formatter={(value) => money(Number(value))} contentStyle={{ borderRadius: 8, fontSize: 12 }} />
                  </PieChart>
                </ResponsiveContainer>
              </div>
              <ul className="mt-4 space-y-1.5">
                {holdings.map((h, i) => (
                  <li key={h.symbol} className="flex items-center justify-between text-sm">
                    <span className="flex items-center gap-2">
                      <span className="h-2.5 w-2.5 rounded-full" style={{ background: SLICE[i % SLICE.length] }} />
                      {h.symbol}
                    </span>
                    <span className="tabular-nums text-muted-foreground">{pct(h.allocation_pct, false)}</span>
                  </li>
                ))}
              </ul>
            </>
          ) : (
            <EmptyState title="No holdings yet" hint="Buy a stock from the market to get started." />
          )}
        </section>

        {/* Holdings table */}
        <section className="rounded-xl border bg-card p-5">
          <h2 className="mb-3 text-sm font-semibold text-muted-foreground">Holdings</h2>
          {holdings.length ? (
            <div className="overflow-x-auto">
              <table className="w-full text-sm">
                <thead>
                  <tr className="border-b text-left text-xs text-muted-foreground">
                    <th className="pb-2 font-medium">Symbol</th>
                    <th className="pb-2 text-right font-medium">Qty</th>
                    <th className="pb-2 text-right font-medium">Avg</th>
                    <th className="pb-2 text-right font-medium">Price</th>
                    <th className="pb-2 text-right font-medium">Value</th>
                    <th className="pb-2 text-right font-medium">P&L</th>
                  </tr>
                </thead>
                <tbody>
                  {holdings.map((h) => (
                    <HoldingRow key={h.symbol} h={h} />
                  ))}
                </tbody>
              </table>
            </div>
          ) : (
            <EmptyState title="Nothing held" />
          )}
        </section>
      </div>

      <div className="grid gap-7 lg:grid-cols-2">
        {/* Your trade history */}
        <section className="rounded-xl border bg-card p-5">
          <h2 className="mb-3 text-sm font-semibold text-muted-foreground">Your recent trades</h2>
          {history.isLoading ? (
            <Loading />
          ) : !history.data?.length ? (
            <EmptyState title="No trades yet" hint="Buy a stock or allocate to an investor." />
          ) : (
            <ul className="divide-y">
              {history.data.slice(0, 12).map((t) => (
                <li key={t.trade_id} className="py-2.5 text-sm">
                  <div className="flex items-center justify-between">
                    <span className="flex min-w-0 items-center gap-2">
                      <span className={cn("rounded px-1.5 py-0.5 text-[10px] font-semibold uppercase", t.side === "buy" ? "bg-gain/15 text-gain" : "bg-loss/15 text-loss")}>{t.side}</span>
                      <span className="font-medium">{t.symbol}</span>
                      <span className="truncate text-muted-foreground">{qty(t.quantity)} @ {money(t.price)}</span>
                      <span className="hidden text-xs text-muted-foreground sm:inline">· {money(t.total_value)}</span>
                    </span>
                    <span className="shrink-0 text-xs text-muted-foreground">{fromUnix(t.executed_at)}</span>
                  </div>
                  {t.reason && <p className="mt-0.5 text-[11px] italic text-muted-foreground">“{t.reason}”</p>}
                </li>
              ))}
            </ul>
          )}
        </section>

        {/* Centralized activity feed — live moves from investors you follow/copy */}
        <section className="rounded-xl border bg-card p-5">
          <h2 className="mb-3 text-sm font-semibold text-muted-foreground">Investor activity</h2>
          {feed.isLoading ? (
            <Loading />
          ) : !feed.data?.length ? (
            <EmptyState title="No activity yet" hint="Follow or allocate to investors to see their moves here." />
          ) : (
            <ul className="divide-y">
              {feed.data.slice(0, 12).map((f: FeedItem, i) => (
                <li key={`${f.investor_id}-${f.executed_at}-${i}`} className="py-2.5 text-sm">
                  <div className="flex items-center justify-between">
                    <span className="flex min-w-0 items-center gap-2">
                      <span className="flex h-6 w-6 shrink-0 items-center justify-center rounded-full bg-primary/12 text-[10px] font-semibold text-primary">
                        {initials(f.investor_name)}
                      </span>
                      <Link to={`/app/investors/${f.investor_id}`} className="shrink-0 font-medium hover:text-primary">{f.investor_name}</Link>
                      <span className={cn("rounded px-1.5 py-0.5 text-[10px] font-semibold uppercase", f.side === "buy" ? "bg-gain/15 text-gain" : "bg-loss/15 text-loss")}>{f.side}</span>
                      <span className="truncate text-muted-foreground">{qty(f.quantity)} {f.symbol} @ {money(f.price)}</span>
                    </span>
                    <span className="shrink-0 text-xs text-muted-foreground">{fromUnix(f.executed_at)}</span>
                  </div>
                  {f.reason && <p className="mt-0.5 pl-8 text-[11px] italic text-muted-foreground">“{f.reason}”</p>}
                </li>
              ))}
            </ul>
          )}
        </section>
      </div>
    </div>
  );
}

function Stat({ label, value, sub, accent }: { label: string; value: string; sub?: string; accent?: string }) {
  return (
    <div className="rounded-xl border bg-card p-4">
      <p className="text-xs text-muted-foreground">{label}</p>
      <p className={cn("font-display mt-1 text-2xl font-semibold tabular-nums", accent)}>{value}</p>
      {sub && <p className={cn("text-xs tabular-nums", accent)}>{sub}</p>}
    </div>
  );
}

function HoldingRow({ h }: { h: HoldingDTO }) {
  return (
    <tr className="border-b last:border-0">
      <td className="py-2.5">
        <Link to={`/app/stock/${h.symbol}`} className="font-medium hover:text-primary">{h.symbol}</Link>
      </td>
      <td className="py-2.5 text-right tabular-nums">{qty(h.quantity)}</td>
      <td className="py-2.5 text-right tabular-nums text-muted-foreground">{money(h.avg_price)}</td>
      <td className="py-2.5 text-right tabular-nums">{money(h.current_price)}</td>
      <td className="py-2.5 text-right tabular-nums">{money(h.market_value)}</td>
      <td className={cn("py-2.5 text-right tabular-nums", pnlColor(h.unrealized_pl))}>
        {money(h.unrealized_pl)} <span className="text-xs">({pct(h.unrealized_pl_pct)})</span>
      </td>
    </tr>
  );
}
