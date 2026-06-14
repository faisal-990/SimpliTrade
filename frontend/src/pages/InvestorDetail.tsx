import { useState } from "react";
import { Link, useParams } from "react-router-dom";
import { ChevronLeft, UserPlus, UserCheck, Quote, Check, Briefcase, LineChart as LineChartIcon } from "lucide-react";
import { LineChart, Line, XAxis, YAxis, Tooltip, ResponsiveContainer, CartesianGrid } from "recharts";
import {
  useInvestor,
  useInvestorTrades,
  useFollow,
  useUnfollow,
  useFollowing,
  useCreateAllocation,
  useAllocations,
  usePortfolioStats,
  useBacktest,
} from "@/hooks/queries";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Loading, ErrorState, EmptyState, Spinner } from "@/components/common/states";
import { money, pct, pnlColor, qty, fromUnix } from "@/lib/format";
import { strategyLabel, riskOf, riskChipClass, initials } from "@/lib/investorMeta";
import { profileFor } from "@/lib/investorProfiles";
import { ApiError } from "@/lib/api";
import { cn } from "@/lib/utils";

export default function InvestorDetail() {
  const { id = "" } = useParams();
  const investor = useInvestor(id);
  const trades = useInvestorTrades(id);
  const following = useFollowing();
  const follow = useFollow();
  const unfollow = useUnfollow();

  if (investor.isLoading) return <Loading />;
  if (investor.isError || !investor.data) return <ErrorState message={(investor.error as Error)?.message} />;

  const inv = investor.data;
  const profile = profileFor(inv.name);
  const risk = riskOf(inv.strategy);
  const isFollowing = (following.data ?? []).some((i) => i.id === id);
  const pending = follow.isPending || unfollow.isPending;
  const toggleFollow = () => (isFollowing ? unfollow.mutate(id) : follow.mutate(id));

  return (
    <div className="space-y-6">
      <Link to="/app/investors" className="inline-flex items-center gap-1 text-sm text-muted-foreground hover:text-foreground">
        <ChevronLeft className="h-4 w-4" /> Investors
      </Link>

      {/* Hero */}
      <header className="overflow-hidden rounded-2xl border bg-card">
        <div className="bg-[radial-gradient(120%_140%_at_0%_0%,oklch(0.7_0.13_60/0.18)_0%,transparent_55%)] p-6 sm:p-7">
          <div className="flex flex-wrap items-start justify-between gap-4">
            <div className="flex items-start gap-4">
              <span className="flex h-14 w-14 shrink-0 items-center justify-center rounded-2xl bg-primary/15 text-lg font-semibold text-primary">
                {initials(inv.name)}
              </span>
              <div>
                <h1 className="text-2xl font-semibold tracking-tight">{inv.name}</h1>
                {profile && <p className="text-sm text-muted-foreground">{profile.firm} · {profile.era}</p>}
                <div className="mt-2 flex flex-wrap items-center gap-1.5">
                  <span className="rounded bg-secondary px-2 py-0.5 text-xs font-medium text-secondary-foreground">{strategyLabel(inv.strategy)}</span>
                  <span className={cn("rounded px-2 py-0.5 text-xs font-medium", riskChipClass(risk.level))}>{risk.label}</span>
                  {inv.rank > 0 && inv.rank < 1_000_000 && (
                    <span className="rounded bg-muted px-2 py-0.5 text-xs font-medium text-muted-foreground">Rank #{inv.rank}</span>
                  )}
                  <span className="rounded bg-muted px-2 py-0.5 text-xs font-medium text-muted-foreground">
                    {inv.followers.toLocaleString()} {inv.followers === 1 ? "follower" : "followers"}
                  </span>
                </div>
              </div>
            </div>
            <div className="flex flex-col items-end gap-3">
              <div className="text-right">
                <p className={cn("text-3xl font-semibold tabular-nums leading-none", pnlColor(inv.roi))}>{pct(inv.roi)}</p>
                <p className="mt-1 text-xs text-muted-foreground">simulated ROI</p>
              </div>
              <Button onClick={toggleFollow} disabled={pending} variant={isFollowing ? "outline" : "default"}>
                {pending ? <Spinner /> : isFollowing ? <><UserCheck className="h-4 w-4" /> Following</> : <><UserPlus className="h-4 w-4" /> Follow</>}
              </Button>
            </div>
          </div>
          {profile && (
            <blockquote className="mt-5 flex gap-2 border-l-2 border-primary/40 pl-3 text-sm italic text-muted-foreground">
              <Quote className="h-4 w-4 shrink-0 text-primary/60" />
              {profile.quote}
            </blockquote>
          )}
        </div>
      </header>

      <div className="grid gap-6 lg:grid-cols-[1.6fr_1fr]">
        <div className="space-y-6">
          {/* Philosophy */}
          <section className="rounded-2xl border bg-card p-6">
            <h2 className="text-sm font-semibold text-muted-foreground">Philosophy</h2>
            <p className="mt-2 text-[15px] leading-relaxed">{profile?.summary ?? inv.bio}</p>
            {profile && (
              <>
                <h3 className="mt-5 text-sm font-semibold text-muted-foreground">How this strategy trades</h3>
                <ul className="mt-2 space-y-1.5">
                  {profile.approach.map((a) => (
                    <li key={a} className="flex gap-2 text-sm">
                      <Check className="mt-0.5 h-4 w-4 shrink-0 text-gain" /> {a}
                    </li>
                  ))}
                </ul>
                <p className="mt-5 text-xs text-muted-foreground">Known for · {profile.knownFor}</p>
              </>
            )}
          </section>

          {/* Recent moves */}
          <section>
            <h2 className="mb-3 text-sm font-semibold text-muted-foreground">Recent moves</h2>
            {trades.isLoading ? (
              <Loading />
            ) : !trades.data?.length ? (
              <EmptyState title="No trades yet" hint="This investor hasn’t placed a trade in the current market window." />
            ) : (
              <ul className="divide-y rounded-2xl border bg-card">
                {trades.data.slice(0, 20).map((t) => (
                  <li key={t.trade_id} className="flex items-center justify-between px-4 py-3 text-sm">
                    <span className="flex items-center gap-2">
                      <span className={cn("rounded px-1.5 py-0.5 text-[10px] font-semibold uppercase", t.side === "buy" ? "bg-gain/15 text-gain" : "bg-loss/15 text-loss")}>{t.side}</span>
                      <Link to={`/app/stock/${t.symbol}`} className="font-medium hover:text-primary">{t.symbol}</Link>
                      <span className="text-muted-foreground">{qty(t.quantity)} @ {money(t.price)}</span>
                    </span>
                    <span className="text-xs text-muted-foreground">{fromUnix(t.executed_at)}</span>
                  </li>
                ))}
              </ul>
            )}
          </section>
        </div>

        {/* Allocate / copy-trade */}
        <AllocatePanel investorId={id} investorName={inv.name} />
      </div>

      {/* Backtest */}
      <BacktestPanel investorId={id} investorName={inv.name} />
    </div>
  );
}

const DATE_FMT = new Intl.DateTimeFormat(undefined, { month: "short", day: "numeric" });
function shortDate(unix: number): string {
  return DATE_FMT.format(new Date(unix * 1000));
}

function BacktestPanel({ investorId, investorName }: { investorId: string; investorName: string }) {
  const backtest = useBacktest();
  const [days, setDays] = useState(180);
  const [cash, setCash] = useState(100000);
  const r = backtest.data;

  const run = () => backtest.mutate({ investorId, days, cash });

  return (
    <section className="rounded-2xl border bg-card p-6">
      <div className="flex flex-wrap items-center justify-between gap-3">
        <div className="flex items-center gap-2">
          <LineChartIcon className="h-4 w-4 text-primary" />
          <h2 className="text-sm font-semibold">Backtest this strategy</h2>
        </div>
        <div className="flex flex-wrap items-end gap-2">
          <label className="flex flex-col text-[11px] text-muted-foreground">
            Lookback
            <select
              value={days}
              onChange={(e) => setDays(Number(e.target.value))}
              className="mt-1 h-9 rounded-md border bg-background px-2 text-sm text-foreground"
            >
              <option value={90}>90 days</option>
              <option value={180}>180 days</option>
              <option value={360}>360 days</option>
            </select>
          </label>
          <label className="flex flex-col text-[11px] text-muted-foreground">
            Starting capital
            <Input
              type="number"
              min={1000}
              step="any"
              value={cash}
              onChange={(e: React.ChangeEvent<HTMLInputElement>) => setCash(Math.max(0, Number(e.target.value)))}
              className="mt-1 h-9 w-32"
            />
          </label>
          <Button className="h-9" disabled={backtest.isPending || cash <= 0} onClick={run}>
            {backtest.isPending ? <Spinner /> : "Run backtest"}
          </Button>
        </div>
      </div>

      <p className="mt-2 text-sm text-muted-foreground">
        Replays {investorName.split(" ")[0]}’s strategy day-by-day over historical prices, starting from cash, and shows
        how the portfolio would have grown.
      </p>

      {backtest.isError && (
        <p className="mt-4 rounded-lg border border-loss/30 bg-loss/10 px-3 py-2 text-sm text-loss">
          {(backtest.error as Error)?.message ?? "Backtest failed"}
        </p>
      )}

      {!r && !backtest.isPending && !backtest.isError && (
        <div className="mt-4">
          <EmptyState title="No backtest yet" hint="Pick a window and starting capital, then run it." />
        </div>
      )}

      {r && (
        <div className="mt-5 space-y-5">
          {/* Metrics */}
          <div className="grid grid-cols-2 gap-3 sm:grid-cols-4 lg:grid-cols-7">
            <Metric label="Final value" value={money(r.final_value)} />
            <Metric label="ROI" value={pct(r.roi)} accent={pnlColor(r.roi)} />
            <Metric label="Max drawdown" value={pct(-r.max_drawdown)} accent="text-loss" />
            <Metric label="Buys" value={String(r.buy_count)} accent="text-gain" />
            <Metric label="Sells" value={String(r.sell_count)} accent={r.sell_count > 0 ? "text-loss" : undefined} />
            <Metric label="Holding" value={String(r.held_count)} />
            <Metric label="Win rate" value={r.sell_count > 0 ? pct(r.win_rate, false) : "—"} />
          </div>

          {/* What it bought-and-held vs sold — directly answers "do bots ever hold/sell?" */}
          <p className="text-xs text-muted-foreground">
            Over this window {investorName.split(" ")[0]} made <strong className="text-foreground">{r.buy_count} buys</strong> and{" "}
            <strong className="text-foreground">{r.sell_count} sells</strong>, and is still <strong className="text-foreground">holding {r.held_count} positions</strong> at the end
            {r.sell_count === 0 && " — a buy-and-hold value strategy rarely sells unless price hits its target or the thesis breaks"}.
          </p>

          {/* Equity curve */}
          <div className="h-64 w-full">
            <ResponsiveContainer width="100%" height="100%">
              <LineChart data={r.equity} margin={{ top: 8, right: 8, bottom: 0, left: 8 }}>
                <CartesianGrid strokeDasharray="3 3" stroke="var(--border)" vertical={false} />
                <XAxis
                  dataKey="date"
                  tickFormatter={shortDate}
                  tick={{ fontSize: 11 }}
                  minTickGap={40}
                  stroke="var(--muted-foreground)"
                />
                <YAxis
                  domain={["auto", "auto"]}
                  tickFormatter={(v) => `$${Math.round(Number(v) / 1000)}k`}
                  tick={{ fontSize: 11 }}
                  width={48}
                  stroke="var(--muted-foreground)"
                />
                <Tooltip
                  formatter={(v) => [money(Number(v)), "Value"]}
                  labelFormatter={(l) => shortDate(Number(l))}
                  contentStyle={{ borderRadius: 8, fontSize: 12 }}
                />
                <Line type="monotone" dataKey="value" stroke="var(--primary)" strokeWidth={2} dot={false} />
              </LineChart>
            </ResponsiveContainer>
          </div>

          {/* Still holding (the "hold" decision made visible) */}
          {r.holdings.length > 0 && (
            <div>
              <h3 className="mb-2 text-xs font-semibold uppercase tracking-wide text-muted-foreground">
                Still holding at end ({r.holdings.length} positions)
              </h3>
              <ul className="divide-y rounded-xl border">
                {r.holdings.map((h) => (
                  <li key={h.symbol} className="flex items-center justify-between px-3 py-2 text-sm">
                    <span className="flex items-center gap-2">
                      <span className="rounded bg-muted px-1.5 py-0.5 text-[10px] font-semibold uppercase text-muted-foreground">hold</span>
                      <span className="font-medium">{h.symbol}</span>
                      <span className="text-muted-foreground">{qty(h.quantity)} @ {money(h.avg_price)}</span>
                    </span>
                    <span className="flex items-center gap-2 tabular-nums">
                      <span>{money(h.market_value)}</span>
                      <span className={cn("w-20 text-right text-xs", pnlColor(h.unrealized_pl))}>{money(h.unrealized_pl)}</span>
                    </span>
                  </li>
                ))}
              </ul>
            </div>
          )}

          {/* Trade log */}
          {r.trades.length > 0 && (
            <div>
              <h3 className="mb-2 text-xs font-semibold uppercase tracking-wide text-muted-foreground">
                Trade log ({r.trade_count} total — showing latest {Math.min(30, r.trades.length)})
              </h3>
              <ul className="max-h-64 divide-y overflow-y-auto rounded-xl border">
                {r.trades.slice(0, 30).map((t, i) => (
                  <li key={`${t.symbol}-${t.date}-${i}`} className="flex items-center justify-between px-3 py-2 text-sm">
                    <span className="flex items-center gap-2">
                      <span className={cn("rounded px-1.5 py-0.5 text-[10px] font-semibold uppercase", t.side === "buy" ? "bg-gain/15 text-gain" : "bg-loss/15 text-loss")}>{t.side}</span>
                      <span className="font-medium">{t.symbol}</span>
                      <span className="text-muted-foreground">{qty(t.quantity)} @ {money(t.price)}</span>
                    </span>
                    <span className="text-xs text-muted-foreground">{shortDate(t.date)}</span>
                  </li>
                ))}
              </ul>
            </div>
          )}

          {r.note && <p className="text-[11px] leading-relaxed text-muted-foreground">⚠ {r.note}</p>}
        </div>
      )}
    </section>
  );
}

function Metric({ label, value, accent }: { label: string; value: string; accent?: string }) {
  return (
    <div className="rounded-xl border bg-background/50 p-3">
      <p className="text-[11px] text-muted-foreground">{label}</p>
      <p className={cn("mt-0.5 text-lg font-semibold tabular-nums", accent)}>{value}</p>
    </div>
  );
}

function AllocatePanel({ investorId, investorName }: { investorId: string; investorName: string }) {
  const stats = usePortfolioStats();
  const allocations = useAllocations();
  const create = useCreateAllocation();
  const [amount, setAmount] = useState(10000);
  const [err, setErr] = useState<string | null>(null);

  const cash = stats.data?.cash ?? 0;
  const mine = (allocations.data ?? []).filter((a) => a.investor_id === investorId && a.is_active);

  const submit = () => {
    setErr(null);
    create.mutate(
      { investor_id: investorId, capital: amount },
      { onError: (e) => setErr(e instanceof ApiError ? e.message : "Could not allocate") }
    );
  };

  return (
    <aside className="h-fit space-y-4 rounded-2xl border bg-card p-6">
      <div className="flex items-center gap-2">
        <Briefcase className="h-4 w-4 text-primary" />
        <h2 className="text-sm font-semibold">Copy this investor</h2>
      </div>
      <p className="text-sm text-muted-foreground">
        Allocate a capped slice of your cash. {investorName.split(" ")[0]}’s bot trades <strong>only that amount</strong> for
        you — your main balance is never touched beyond it.
      </p>

      <div className="space-y-2">
        <label htmlFor="amt" className="text-xs font-medium text-muted-foreground">Amount to allocate</label>
        <Input
          id="amt"
          type="number"
          min={0}
          step="any"
          value={amount}
          onChange={(e: React.ChangeEvent<HTMLInputElement>) => setAmount(Math.max(0, Number(e.target.value)))}
          className="h-11"
        />
        <p className="text-xs text-muted-foreground">Available cash: {money(cash)}</p>
      </div>

      <Button className="h-11 w-full" disabled={create.isPending || amount <= 0 || amount > cash} onClick={submit}>
        {create.isPending ? <Spinner /> : `Allocate ${money(amount)} & start`}
      </Button>
      {err && <p className="rounded-lg border border-loss/30 bg-loss/10 px-3 py-2 text-sm text-loss">{err}</p>}

      {mine.length > 0 && (
        <div className="rounded-lg border bg-background/50 p-3">
          <p className="text-xs font-medium text-muted-foreground">Active allocations to {investorName.split(" ")[0]}</p>
          {mine.map((a) => (
            <div key={a.id} className="mt-2 flex items-center justify-between text-sm">
              <span>{money(a.market_value)}</span>
              <span className={cn("tabular-nums", pnlColor(a.return_pct))}>{pct(a.return_pct)}</span>
            </div>
          ))}
          <p className="mt-2 text-[11px] text-muted-foreground">Manage these in your Portfolio.</p>
        </div>
      )}

      <p className="text-[11px] leading-relaxed text-muted-foreground">
        The bot evaluates and trades your allocation on each market tick. Run the engine (or the admin “open market”
        control) to see trades appear.
      </p>
    </aside>
  );
}
