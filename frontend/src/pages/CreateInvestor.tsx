import { useState } from "react";
import { Link, useNavigate } from "react-router-dom";
import { ChevronLeft, Sparkles, Wand2, LineChart } from "lucide-react";
import { useCreateInvestor, type CreateInvestorInput } from "@/hooks/queries";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Spinner } from "@/components/common/states";
import { ApiError } from "@/lib/api";
import { cn } from "@/lib/utils";

type Approach = CreateInvestorInput["approach"];

const APPROACHES: { id: Approach; label: string; blurb: string }[] = [
  { id: "value", label: "Value", blurb: "Buy cheap vs. intrinsic value — only names trading at low P/E and P/B. Patient, contrarian." },
  { id: "quality", label: "Quality", blurb: "Own great businesses — high return on equity and fat margins. Buy and hold compounders." },
  { id: "growth", label: "Growth", blurb: "Back fast growers — strong revenue and earnings growth. Pays up for expansion." },
  { id: "momentum", label: "Momentum", blurb: "Ride winners — buy what's already trending up. Trades more, cuts losers fast." },
];

// The form holds values in human units: percentage fields as whole percents
// (15 = 15%), ratio/count fields as-is. Converted to fractions on submit.
interface FormState {
  name: string;
  philosophy: string;
  approach: Approach;
  max_positions: number;
  pe_max: number; // ratio
  pb_max: number; // ratio
  roe_min: number; // %
  operating_margin_min: number; // %
  revenue_growth_min: number; // %
  eps_growth_min: number; // %
  return_6m_min: number; // %
  stop_loss_pct: number; // %
  take_profit_vs_intrinsic: number; // multiple
  max_position_size: number; // %
  cash_buffer_min: number; // %
  position_sizing: "equal" | "conviction";
}

const DEFAULTS: FormState = {
  name: "",
  philosophy: "",
  approach: "value",
  max_positions: 15,
  pe_max: 15,
  pb_max: 1.5,
  roe_min: 15,
  operating_margin_min: 15,
  revenue_growth_min: 10,
  eps_growth_min: 10,
  return_6m_min: 10,
  stop_loss_pct: 20,
  take_profit_vs_intrinsic: 1.0,
  max_position_size: 10,
  cash_buffer_min: 5,
  position_sizing: "equal",
};

export default function CreateInvestor() {
  const create = useCreateInvestor();
  const navigate = useNavigate();
  const [form, setForm] = useState<FormState>(DEFAULTS);
  const [error, setError] = useState<string | null>(null);

  const set = <K extends keyof FormState>(k: K, v: FormState[K]) => setForm((f) => ({ ...f, [k]: v }));
  const num = (k: keyof FormState) => (e: React.ChangeEvent<HTMLInputElement>) => set(k, Number(e.target.value) as never);

  const submit = (e: React.FormEvent) => {
    e.preventDefault();
    setError(null);
    if (form.name.trim().length < 2) {
      setError("Give your investor a name.");
      return;
    }
    const pct = (v: number) => v / 100; // percent → fraction for the API
    const payload: CreateInvestorInput = {
      name: form.name.trim(),
      philosophy: form.philosophy.trim(),
      approach: form.approach,
      max_positions: form.max_positions,
      pe_max: form.pe_max,
      pb_max: form.pb_max,
      roe_min: pct(form.roe_min),
      operating_margin_min: pct(form.operating_margin_min),
      revenue_growth_min: pct(form.revenue_growth_min),
      eps_growth_min: pct(form.eps_growth_min),
      return_6m_min: pct(form.return_6m_min),
      stop_loss_pct: pct(form.stop_loss_pct),
      take_profit_vs_intrinsic: form.take_profit_vs_intrinsic,
      max_position_size: pct(form.max_position_size),
      cash_buffer_min: pct(form.cash_buffer_min),
      position_sizing: form.position_sizing,
    };
    create.mutate(payload, {
      onSuccess: (inv) => navigate(`/app/investors/${inv.id}`, { replace: true }),
      onError: (err) => setError(err instanceof ApiError ? err.message : "Could not create investor"),
    });
  };

  return (
    <div className="mx-auto max-w-2xl space-y-6">
      <Link to="/app/investors" className="inline-flex items-center gap-1 text-sm text-muted-foreground hover:text-foreground">
        <ChevronLeft className="h-4 w-4" /> Investors
      </Link>

      <header className="flex items-center gap-3">
        <span className="flex h-11 w-11 items-center justify-center rounded-2xl bg-primary/15 text-primary">
          <Wand2 className="h-5 w-5" />
        </span>
        <div>
          <h1 className="text-2xl font-semibold tracking-tight">Build your own investor</h1>
          <p className="text-sm text-muted-foreground">Define a strategy and it trades on the same engine as the legends.</p>
        </div>
      </header>

      <div className="flex items-start gap-2 rounded-xl bg-primary/8 px-3 py-2.5 text-xs text-muted-foreground">
        <LineChart className="mt-0.5 h-4 w-4 shrink-0 text-primary" />
        Every box has a hint explaining what it does. After you create it, open its page to <strong className="text-foreground">backtest</strong> it
        over historical prices, then allocate funds and watch it trade. Nothing here risks real money.
      </div>

      {error && <div className="rounded-lg border border-loss/30 bg-loss/10 px-3 py-2 text-sm text-loss">{error}</div>}

      <form onSubmit={submit} className="space-y-6">
        <Section title="Identity">
          <div className="space-y-2">
            <Label htmlFor="name">Name</Label>
            <Input id="name" value={form.name} maxLength={100} placeholder="e.g. Steady Compounder"
              onChange={(e: React.ChangeEvent<HTMLInputElement>) => set("name", e.target.value)} className="h-11" />
          </div>
          <div className="space-y-2">
            <Label htmlFor="philo">Philosophy</Label>
            <textarea id="philo" value={form.philosophy} rows={2} maxLength={500}
              placeholder="One line on how this investor thinks…"
              onChange={(e: React.ChangeEvent<HTMLTextAreaElement>) => set("philosophy", e.target.value)}
              className="w-full rounded-md border bg-background px-3 py-2 text-sm outline-none focus-visible:ring-2 focus-visible:ring-ring" />
            <p className="text-[11px] text-muted-foreground">Shown on the investor's page. Just flavor — it doesn't affect trading.</p>
          </div>
        </Section>

        <Section title="Approach">
          <div className="grid grid-cols-2 gap-2.5 sm:grid-cols-4">
            {APPROACHES.map((a) => (
              <button key={a.id} type="button" onClick={() => set("approach", a.id)}
                className={cn("rounded-xl border p-3 text-left transition-colors",
                  form.approach === a.id ? "border-primary bg-primary/8" : "hover:bg-accent")}>
                <p className="text-sm font-semibold">{a.label}</p>
              </button>
            ))}
          </div>
          <p className="text-xs text-muted-foreground">{APPROACHES.find((a) => a.id === form.approach)?.blurb}</p>
        </Section>

        <Section title="Buy rules — what it's willing to buy">
          {form.approach === "value" && (
            <Grid>
              <Num label="Max P/E" value={form.pe_max} onChange={num("pe_max")} hint="Only buy below this price-to-earnings. Lower = cheaper & stricter. Classic value: 10–15." />
              <Num label="Max P/B" value={form.pb_max} step="0.1" onChange={num("pb_max")} hint="Price-to-book ceiling. Below 1.5 is classic deep value. Lower = stricter." />
            </Grid>
          )}
          {form.approach === "quality" && (
            <Grid>
              <Num label="Min ROE %" value={form.roe_min} step="1" onChange={num("roe_min")} hint="Return on equity floor. Higher = only elite, profitable businesses. 15–20% is strong." />
              <Num label="Min operating margin %" value={form.operating_margin_min} step="1" onChange={num("operating_margin_min")} hint="Profitability floor. Higher = pickier about fat margins. 15%+ is healthy." />
            </Grid>
          )}
          {form.approach === "growth" && (
            <Grid>
              <Num label="Min revenue growth %" value={form.revenue_growth_min} step="1" onChange={num("revenue_growth_min")} hint="Year-over-year sales growth floor. Higher = only fast growers. 10–20% is solid." />
              <Num label="Min EPS growth %" value={form.eps_growth_min} step="1" onChange={num("eps_growth_min")} hint="Year-over-year earnings growth floor. Higher = stricter." />
            </Grid>
          )}
          {form.approach === "momentum" && (
            <Grid>
              <Num label="Min 6-month return %" value={form.return_6m_min} step="1" onChange={num("return_6m_min")} hint="Only buy names already up at least this much over 6 months. Higher = chase stronger trends." />
            </Grid>
          )}
        </Section>

        <Section title="Sell rules — when it exits">
          <Grid>
            <Num label="Stop-loss %" value={form.stop_loss_pct} step="1" onChange={num("stop_loss_pct")} hint="Sell if a position falls this far below your cost. 20% = cut losers at −20%. 0 = never auto-stop." />
            <Num label="Take-profit (× fair value)" value={form.take_profit_vs_intrinsic} step="0.1" onChange={num("take_profit_vs_intrinsic")} hint="Sell when price reaches this multiple of estimated fair value. 1.0 = sell at fair value; 1.2 = let it run 20% past. 0 = hold." />
          </Grid>
        </Section>

        <Section title="Risk & sizing — how it spreads the money">
          <Grid>
            <Num label="Max positions" value={form.max_positions} step="1" onChange={num("max_positions")} hint="How many names to hold at once. More = more diversified, smaller bets each." />
            <Num label="Max position size %" value={form.max_position_size} step="1" onChange={num("max_position_size")} hint="Cap on any single holding. 10% = no stock can exceed 10% of the portfolio." />
            <Num label="Cash buffer %" value={form.cash_buffer_min} step="1" onChange={num("cash_buffer_min")} hint="Minimum kept in cash, never invested. 5% = always hold ≥5% dry powder." />
            <div className="space-y-1.5">
              <Label>Position sizing</Label>
              <select value={form.position_sizing} onChange={(e) => set("position_sizing", e.target.value as FormState["position_sizing"])}
                className="h-11 w-full rounded-md border bg-background px-2 text-sm">
                <option value="equal">Equal weight</option>
                <option value="conviction">By conviction</option>
              </select>
              <p className="text-[11px] text-muted-foreground">Equal = same size each. Conviction = bigger bets on higher-scoring names.</p>
            </div>
          </Grid>
        </Section>

        <Button type="submit" className="h-11 w-full" disabled={create.isPending}>
          {create.isPending ? <Spinner /> : <><Sparkles className="h-4 w-4" /> Create investor</>}
        </Button>
        <p className="text-center text-[11px] text-muted-foreground">
          Next step on its page: <strong className="text-foreground">Run backtest</strong> to see how it would've performed, then allocate funds.
        </p>
      </form>
    </div>
  );
}

function Section({ title, children }: { title: string; children: React.ReactNode }) {
  return (
    <section className="space-y-3 rounded-2xl border bg-card p-5">
      <h2 className="text-sm font-semibold text-muted-foreground">{title}</h2>
      {children}
    </section>
  );
}

function Grid({ children }: { children: React.ReactNode }) {
  return <div className="grid gap-3 sm:grid-cols-2">{children}</div>;
}

function Num({ label, value, onChange, step = "any", hint }: {
  label: string; value: number; onChange: (e: React.ChangeEvent<HTMLInputElement>) => void; step?: string; hint?: string;
}) {
  return (
    <div className="space-y-1.5">
      <Label>{label}</Label>
      <Input type="number" step={step} value={value} onChange={onChange} className="h-11" />
      {hint && <p className="text-[11px] leading-relaxed text-muted-foreground">{hint}</p>}
    </div>
  );
}
