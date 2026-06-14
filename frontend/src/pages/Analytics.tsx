import { LineChart, Line, XAxis, YAxis, Tooltip, ResponsiveContainer, CartesianGrid, Legend } from "recharts";
import { useAnalytics } from "@/hooks/queries";
import { Loading, ErrorState, EmptyState } from "@/components/common/states";
import { money, pct, pnlColor } from "@/lib/format";
import { cn } from "@/lib/utils";

const DATE_FMT = new Intl.DateTimeFormat(undefined, { month: "short", day: "numeric" });
const shortDate = (unix: number) => DATE_FMT.format(new Date(unix * 1000));
const SLICE = ["var(--chart-1)", "var(--chart-2)", "var(--chart-4)", "var(--chart-5)", "var(--chart-3)"];

export default function Analytics() {
  const { data: a, isLoading, isError, error } = useAnalytics();

  if (isLoading) return <Loading label="Crunching your performance…" />;
  if (isError || !a) return <ErrorState message={(error as Error)?.message} />;

  // Rebase portfolio + benchmark to % return from the start, so they compare
  // fairly regardless of absolute scale. Both series share the same timeline.
  const bench = new Map(a.benchmark.map((p) => [p.date, p.value]));
  const benchStart = a.benchmark[0]?.value ?? a.start_value;
  const chart = a.equity.map((p) => ({
    date: p.date,
    portfolio: (p.value / a.start_value - 1) * 100,
    market: bench.has(p.date) ? ((bench.get(p.date) as number) / benchStart - 1) * 100 : null,
  }));

  const traded = a.trade_count > 0 || a.sectors.length > 0;

  return (
    <div className="space-y-7">
      <header>
        <h1 className="text-2xl font-semibold tracking-tight">Analytics</h1>
        <p className="mt-1 text-sm text-muted-foreground">How your portfolio has performed — return, risk, and where you're exposed.</p>
      </header>

      {/* Metrics */}
      <div className="grid grid-cols-2 gap-3 sm:grid-cols-3 lg:grid-cols-6">
        <Metric label="Total value" value={money(a.current_value)} />
        <Metric label="ROI" value={pct(a.roi)} accent={pnlColor(a.roi)} />
        <Metric label="Max drawdown" value={pct(-a.max_drawdown)} accent="text-loss" />
        <Metric label="Volatility (ann.)" value={pct(a.volatility, false)} />
        <Metric label="Sharpe" value={a.sharpe.toFixed(2)} accent={pnlColor(a.sharpe)} />
        <Metric label="Win rate" value={a.trade_count > 0 ? pct(a.win_rate, false) : "—"} />
      </div>

      {!traded && (
        <EmptyState title="Not much to show yet" hint="Buy a few stocks (or allocate to investors) and your performance will build up here." />
      )}

      {/* Equity vs market */}
      <section className="rounded-2xl border bg-card p-5">
        <div className="mb-3 flex items-center justify-between">
          <h2 className="text-sm font-semibold text-muted-foreground">Return vs. {a.benchmark_name}</h2>
          <span className="text-xs text-muted-foreground">% return since start</span>
        </div>
        <div className="h-72 w-full">
          <ResponsiveContainer width="100%" height="100%">
            <LineChart data={chart} margin={{ top: 8, right: 8, bottom: 0, left: 8 }}>
              <CartesianGrid strokeDasharray="3 3" stroke="var(--border)" vertical={false} />
              <XAxis dataKey="date" tickFormatter={shortDate} tick={{ fontSize: 11 }} minTickGap={40} stroke="var(--muted-foreground)" />
              <YAxis tickFormatter={(v) => `${Number(v).toFixed(0)}%`} tick={{ fontSize: 11 }} width={44} stroke="var(--muted-foreground)" />
              <Tooltip
                formatter={(v, name) => [v == null ? "—" : `${Number(v).toFixed(2)}%`, name === "portfolio" ? "Your portfolio" : "Market"]}
                labelFormatter={(l) => shortDate(Number(l))}
                contentStyle={{ borderRadius: 8, fontSize: 12 }}
              />
              <Legend formatter={(v) => (v === "portfolio" ? "Your portfolio" : "Market")} wrapperStyle={{ fontSize: 12 }} />
              <Line type="monotone" dataKey="portfolio" stroke="var(--primary)" strokeWidth={2} dot={false} />
              <Line type="monotone" dataKey="market" stroke="var(--muted-foreground)" strokeWidth={1.5} strokeDasharray="4 3" dot={false} connectNulls />
            </LineChart>
          </ResponsiveContainer>
        </div>
        <p className="mt-2 text-[11px] text-muted-foreground">
          Best day {pct(a.best_day)} · worst day {pct(a.worst_day)}. Your line is flat while you hold cash.
        </p>
      </section>

      {/* Sector exposure */}
      <section className="rounded-2xl border bg-card p-5">
        <h2 className="mb-3 text-sm font-semibold text-muted-foreground">Sector exposure</h2>
        {a.sectors.length ? (
          <ul className="space-y-2.5">
            {a.sectors.map((s, i) => (
              <li key={s.sector}>
                <div className="mb-1 flex items-center justify-between text-sm">
                  <span className="flex items-center gap-2">
                    <span className="h-2.5 w-2.5 rounded-full" style={{ background: SLICE[i % SLICE.length] }} />
                    {s.sector}
                  </span>
                  <span className="tabular-nums text-muted-foreground">{money(s.value)} · {pct(s.pct, false)}</span>
                </div>
                <div className="h-2 w-full overflow-hidden rounded-full bg-muted">
                  <div className="h-full rounded-full" style={{ width: `${Math.min(100, s.pct * 100)}%`, background: SLICE[i % SLICE.length] }} />
                </div>
              </li>
            ))}
          </ul>
        ) : (
          <EmptyState title="No holdings" hint="Your sector mix appears once you own stocks." />
        )}
      </section>
    </div>
  );
}

function Metric({ label, value, accent }: { label: string; value: string; accent?: string }) {
  return (
    <div className="rounded-xl border bg-card p-4">
      <p className="text-xs text-muted-foreground">{label}</p>
      <p className={cn("mt-1 text-lg font-semibold tabular-nums", accent)}>{value}</p>
    </div>
  );
}
