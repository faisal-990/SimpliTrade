import { useState, type ChangeEvent } from "react";
import { Link, useParams } from "react-router-dom";
import { ChevronLeft, Check } from "lucide-react";
import { useStock, useTrade } from "@/hooks/queries";
import { AdvancedChart, tvSymbol } from "@/components/market/TradingView";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Loading, ErrorState, Spinner } from "@/components/common/states";
import { money, marketCap, pct, ratio } from "@/lib/format";
import { ApiError } from "@/lib/api";
import type { Fundamentals } from "@/types/api";

export default function StockDetail() {
  const { symbol = "" } = useParams();
  const { data, isLoading, isError, error } = useStock(symbol);

  if (isLoading) return <Loading label={`Loading ${symbol}…`} />;
  if (isError || !data) return <ErrorState message={(error as Error)?.message} />;

  const f = data.fundamentals;
  return (
    <div className="space-y-6">
      <Link to="/app/dashboard" className="inline-flex items-center gap-1 text-sm text-muted-foreground hover:text-foreground">
        <ChevronLeft className="h-4 w-4" /> Market
      </Link>

      <header className="flex flex-wrap items-end justify-between gap-3">
        <div>
          <div className="flex items-center gap-2">
            <h1 className="text-2xl font-semibold tracking-tight">{data.symbol}</h1>
            <span className="rounded bg-muted px-2 py-0.5 text-xs font-medium text-muted-foreground">{data.sector}</span>
          </div>
          <p className="mt-0.5 text-sm text-muted-foreground">{data.name} · {data.exchange}</p>
        </div>
        <p className="text-2xl font-semibold tabular-nums">{money(data.current_price)}</p>
      </header>

      <div className="grid gap-6 lg:grid-cols-[1.7fr_1fr]">
        <div className="space-y-6">
          <AdvancedChart symbol={tvSymbol(data.symbol, data.exchange)} />
          <Fundamentals f={f} />
        </div>
        <TradePanel symbol={data.symbol} price={data.current_price} />
      </div>
    </div>
  );
}

function Metric({ label, value }: { label: string; value: string }) {
  return (
    <div className="rounded-lg border bg-card p-3">
      <p className="text-xs text-muted-foreground">{label}</p>
      <p className="mt-0.5 font-medium tabular-nums">{value}</p>
    </div>
  );
}

function Fundamentals({ f }: { f: Fundamentals }) {
  const allZero = !f || (f.pe === 0 && f.pb === 0 && f.roe === 0 && f.market_cap === 0);
  return (
    <section>
      <h2 className="mb-3 text-sm font-semibold text-muted-foreground">Fundamentals</h2>
      {allZero && (
        <p className="mb-3 rounded-lg border bg-muted/40 px-3 py-2 text-xs text-muted-foreground">
          Fundamentals aren’t available for this symbol on the current data plan — momentum-based views still work.
        </p>
      )}
      <div className="grid grid-cols-2 gap-3 sm:grid-cols-3">
        <Metric label="Market cap" value={marketCap(f?.market_cap)} />
        <Metric label="P/E" value={ratio(f?.pe)} />
        <Metric label="P/B" value={ratio(f?.pb)} />
        <Metric label="ROE" value={f?.roe ? pct(f.roe, false) : "—"} />
        <Metric label="Gross margin" value={f?.gross_margin ? pct(f.gross_margin, false) : "—"} />
        <Metric label="Dividend yield" value={f?.dividend_yield ? pct(f.dividend_yield, false) : "—"} />
        <Metric label="Debt / equity" value={ratio(f?.debt_to_equity)} />
        <Metric label="Beta" value={ratio(f?.beta)} />
        <Metric label="EPS (ttm)" value={f?.eps_ttm ? money(f.eps_ttm) : "—"} />
      </div>
    </section>
  );
}

function TradePanel({ symbol, price }: { symbol: string; price: number }) {
  const [quantity, setQuantity] = useState(1);
  const [done, setDone] = useState<string | null>(null);
  const [err, setErr] = useState<string | null>(null);
  const buy = useTrade("buy");
  const sell = useTrade("sell");
  const pending = buy.isPending || sell.isPending;

  const submit = (side: "buy" | "sell") => {
    setDone(null);
    setErr(null);
    const m = side === "buy" ? buy : sell;
    m.mutate(
      { symbol, quantity },
      {
        onSuccess: (t) => setDone(`${side === "buy" ? "Bought" : "Sold"} ${t.quantity} ${symbol} @ ${money(t.price)}`),
        onError: (e) => setErr(e instanceof ApiError ? e.message : "Trade failed"),
      }
    );
  };

  const estimate = price * (quantity || 0);

  return (
    <aside className="h-fit rounded-xl border bg-card p-5">
      <h2 className="text-sm font-semibold">Trade {symbol}</h2>
      <p className="mt-0.5 text-xs text-muted-foreground">Simulated order at the latest price.</p>

      <div className="mt-4 space-y-2">
        <Label htmlFor="qty">Quantity</Label>
        <Input
          id="qty"
          type="number"
          min={0}
          step="any"
          value={quantity}
          onChange={(e: ChangeEvent<HTMLInputElement>) => setQuantity(Math.max(0, Number(e.target.value)))}
          className="h-11"
        />
      </div>

      <div className="mt-4 flex items-center justify-between rounded-lg bg-muted/50 px-3 py-2 text-sm">
        <span className="text-muted-foreground">Estimated total</span>
        <span className="font-semibold tabular-nums">{money(estimate)}</span>
      </div>

      <div className="mt-4 grid grid-cols-2 gap-3">
        <Button onClick={() => submit("buy")} disabled={pending || quantity <= 0} className="h-11">
          {buy.isPending ? <Spinner /> : "Buy"}
        </Button>
        <Button onClick={() => submit("sell")} disabled={pending || quantity <= 0} variant="outline" className="h-11">
          {sell.isPending ? <Spinner /> : "Sell"}
        </Button>
      </div>

      {done && (
        <p className="mt-4 flex items-center gap-1.5 rounded-lg border border-gain/30 bg-gain/10 px-3 py-2 text-sm text-gain">
          <Check className="h-4 w-4" /> {done}
        </p>
      )}
      {err && <p className="mt-4 rounded-lg border border-loss/30 bg-loss/10 px-3 py-2 text-sm text-loss">{err}</p>}
    </aside>
  );
}
