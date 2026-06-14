import { useState } from "react";
import { Link } from "react-router-dom";
import { useQueries } from "@tanstack/react-query";
import { ChevronDown, ChevronUp } from "lucide-react";
import { api } from "@/lib/api";
import { qk } from "@/hooks/queries";
import { Button } from "@/components/ui/button";
import { Spinner } from "@/components/common/states";
import { money, pct, pnlColor, qty, fromUnix } from "@/lib/format";
import { strategyLabel } from "@/lib/investorMeta";
import { cn } from "@/lib/utils";
import type { AllocationDetail, AllocationHolding, AllocationTrade } from "@/types/api";

export interface AggAlloc {
  investor_id: string;
  investor_name: string;
  strategy: string;
  capital: number;
  cash: number;
  market_value: number;
  return_pct: number;
  ids: string[];
}

// Merge holdings from several copy sub-accounts (same investor) by symbol.
function mergeHoldings(details: AllocationDetail[]): AllocationHolding[] {
  const bySymbol = new Map<string, AllocationHolding>();
  for (const d of details) {
    for (const h of d.holdings ?? []) {
      const cur = bySymbol.get(h.symbol);
      if (cur) {
        const totalQty = cur.quantity + h.quantity;
        cur.avg_price = totalQty > 0 ? (cur.avg_price * cur.quantity + h.avg_price * h.quantity) / totalQty : 0;
        cur.quantity = totalQty;
        cur.market_value += h.market_value;
        cur.current_price = h.current_price;
        cur.unrealized_pl = (cur.current_price - cur.avg_price) * cur.quantity;
        cur.unrealized_pl_pct = cur.avg_price > 0 ? cur.current_price / cur.avg_price - 1 : 0; // fraction
      } else {
        bySymbol.set(h.symbol, { ...h });
      }
    }
  }
  return [...bySymbol.values()].sort((a, b) => b.market_value - a.market_value);
}

function mergeTrades(details: AllocationDetail[]): AllocationTrade[] {
  return details
    .flatMap((d) => d.trades ?? [])
    .sort((a, b) => b.executed_at - a.executed_at)
    .slice(0, 20);
}

export function AllocationCard({ a, onStop, stopping }: { a: AggAlloc; onStop: () => void; stopping: boolean }) {
  const [open, setOpen] = useState(false);

  // Fetch each sub-account's activity only once expanded.
  const results = useQueries({
    queries: a.ids.map((id) => ({
      queryKey: qk.allocation(id),
      queryFn: () => api.get<AllocationDetail>(`/allocations/${id}`),
      enabled: open,
    })),
  });
  const loading = open && results.some((r) => r.isLoading);
  const details = results.map((r) => r.data).filter(Boolean) as AllocationDetail[];
  const holdings = mergeHoldings(details);
  const trades = mergeTrades(details);

  return (
    <div className="rounded-xl border bg-background/50 p-4">
      <div className="flex items-start justify-between">
        <div className="min-w-0">
          <Link to={`/app/investors/${a.investor_id}`} className="block truncate font-medium hover:text-primary">{a.investor_name}</Link>
          <p className="truncate text-xs text-muted-foreground">
            {strategyLabel(a.strategy)}
            {a.ids.length > 1 && <span className="ml-1 text-primary">· {a.ids.length} allocations</span>}
          </p>
        </div>
        <span className={cn("text-sm font-semibold tabular-nums", pnlColor(a.return_pct))}>{pct(a.return_pct)}</span>
      </div>

      <div className="mt-3 flex items-end justify-between">
        <div>
          <p className="text-[11px] text-muted-foreground">Value</p>
          <p className="font-semibold tabular-nums">{money(a.market_value)}</p>
          <p className="text-[11px] text-muted-foreground">of {money(a.capital)} allocated · {money(a.cash)} cash</p>
        </div>
        <Button size="sm" variant="outline" disabled={stopping} onClick={onStop}>
          {a.ids.length > 1 ? "Stop all" : "Stop"}
        </Button>
      </div>

      <button
        onClick={() => setOpen((v) => !v)}
        className="mt-3 flex w-full items-center justify-center gap-1 rounded-lg border border-dashed py-1.5 text-xs font-medium text-muted-foreground hover:bg-accent hover:text-accent-foreground"
      >
        {open ? <ChevronUp className="h-3.5 w-3.5" /> : <ChevronDown className="h-3.5 w-3.5" />}
        {open ? "Hide bot activity" : "See what the bot did"}
      </button>

      {open && (
        <div className="mt-3 space-y-3 border-t pt-3">
          {loading ? (
            <div className="flex justify-center py-3"><Spinner /></div>
          ) : (
            <>
              <div>
                <p className="mb-1.5 text-[11px] font-semibold uppercase tracking-wide text-muted-foreground">Positions</p>
                {holdings.length ? (
                  <ul className="space-y-1">
                    {holdings.map((h) => (
                      <li key={h.symbol} className="flex items-center justify-between text-xs">
                        <span className="flex items-center gap-1.5">
                          <Link to={`/app/stock/${h.symbol}`} className="font-medium hover:text-primary">{h.symbol}</Link>
                          <span className="text-muted-foreground">{qty(h.quantity)} @ {money(h.avg_price)}</span>
                        </span>
                        <span className="flex items-center gap-2 tabular-nums">
                          <span>{money(h.market_value)}</span>
                          <span className={cn("w-12 text-right", pnlColor(h.unrealized_pl))}>{pct(h.unrealized_pl_pct)}</span>
                        </span>
                      </li>
                    ))}
                  </ul>
                ) : (
                  <p className="text-xs text-muted-foreground">No open positions — the bot is holding cash, waiting for a setup that fits its strategy.</p>
                )}
              </div>

              <div>
                <p className="mb-1.5 text-[11px] font-semibold uppercase tracking-wide text-muted-foreground">Recent orders</p>
                {trades.length ? (
                  <ul className="space-y-1">
                    {trades.map((t, i) => (
                      <li key={`${t.symbol}-${t.executed_at}-${i}`} className="flex items-center justify-between text-xs">
                        <span className="flex items-center gap-1.5">
                          <span className={cn("rounded px-1 py-0.5 text-[9px] font-semibold uppercase", t.side === "buy" ? "bg-gain/15 text-gain" : "bg-loss/15 text-loss")}>{t.side}</span>
                          <span className="font-medium">{t.symbol}</span>
                          <span className="text-muted-foreground">{qty(t.quantity)} @ {money(t.price)}</span>
                        </span>
                        <span className="text-muted-foreground">{fromUnix(t.executed_at)}</span>
                      </li>
                    ))}
                  </ul>
                ) : (
                  <p className="text-xs text-muted-foreground">No orders yet. Hit <span className="font-medium text-foreground">Simulate market</span> to run a cycle and let the bot trade this capital.</p>
                )}
              </div>
            </>
          )}
        </div>
      )}
    </div>
  );
}
