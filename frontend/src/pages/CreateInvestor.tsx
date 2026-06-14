import { useState } from "react";
import { Link, useNavigate } from "react-router-dom";
import { ChevronLeft, Sparkles, Wand2 } from "lucide-react";
import { useCreateInvestor, type CreateInvestorInput } from "@/hooks/queries";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Spinner } from "@/components/common/states";
import { ApiError } from "@/lib/api";
import { cn } from "@/lib/utils";

type Approach = CreateInvestorInput["approach"];

const APPROACHES: { id: Approach; label: string; blurb: string }[] = [
  { id: "value", label: "Value", blurb: "Buy cheap vs. intrinsic value — low P/E, low P/B." },
  { id: "quality", label: "Quality", blurb: "Own great businesses — high ROE & fat margins." },
  { id: "growth", label: "Growth", blurb: "Back fast growers — strong revenue & EPS growth." },
  { id: "momentum", label: "Momentum", blurb: "Ride winners — buy what's trending up." },
];

// Sensible starting points per approach (fractions where noted).
const DEFAULTS: CreateInvestorInput = {
  name: "",
  philosophy: "",
  approach: "value",
  max_positions: 15,
  pe_max: 15,
  pb_max: 1.5,
  roe_min: 0.15,
  operating_margin_min: 0.15,
  revenue_growth_min: 0.1,
  eps_growth_min: 0.1,
  return_6m_min: 0.1,
  stop_loss_pct: 0.2,
  take_profit_vs_intrinsic: 1.0,
  max_position_size: 0.1,
  cash_buffer_min: 0.05,
  position_sizing: "equal",
};

export default function CreateInvestor() {
  const create = useCreateInvestor();
  const navigate = useNavigate();
  const [form, setForm] = useState<CreateInvestorInput>(DEFAULTS);
  const [error, setError] = useState<string | null>(null);

  const set = <K extends keyof CreateInvestorInput>(k: K, v: CreateInvestorInput[K]) =>
    setForm((f) => ({ ...f, [k]: v }));

  const num = (k: keyof CreateInvestorInput) => (e: React.ChangeEvent<HTMLInputElement>) =>
    set(k, Number(e.target.value) as never);

  const submit = (e: React.FormEvent) => {
    e.preventDefault();
    setError(null);
    if (form.name.trim().length < 2) {
      setError("Give your investor a name.");
      return;
    }
    create.mutate(form, {
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

      {error && <div className="rounded-lg border border-loss/30 bg-loss/10 px-3 py-2 text-sm text-loss">{error}</div>}

      <form onSubmit={submit} className="space-y-6">
        {/* Identity */}
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
          </div>
        </Section>

        {/* Approach */}
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

        {/* Buy rules (approach-specific) */}
        <Section title="Buy rules">
          {form.approach === "value" && (
            <Grid>
              <Num label="Max P/E" value={form.pe_max} onChange={num("pe_max")} hint="lower = cheaper" />
              <Num label="Max P/B" value={form.pb_max} step="0.1" onChange={num("pb_max")} />
            </Grid>
          )}
          {form.approach === "quality" && (
            <Grid>
              <Num label="Min ROE" value={form.roe_min} step="0.01" onChange={num("roe_min")} hint="fraction, e.g. 0.15 = 15%" />
              <Num label="Min operating margin" value={form.operating_margin_min} step="0.01" onChange={num("operating_margin_min")} hint="fraction" />
            </Grid>
          )}
          {form.approach === "growth" && (
            <Grid>
              <Num label="Min revenue growth" value={form.revenue_growth_min} step="0.01" onChange={num("revenue_growth_min")} hint="YoY fraction" />
              <Num label="Min EPS growth" value={form.eps_growth_min} step="0.01" onChange={num("eps_growth_min")} hint="YoY fraction" />
            </Grid>
          )}
          {form.approach === "momentum" && (
            <Grid>
              <Num label="Min 6-month return" value={form.return_6m_min} step="0.01" onChange={num("return_6m_min")} hint="fraction, e.g. 0.10 = 10%" />
            </Grid>
          )}
        </Section>

        {/* Sell rules */}
        <Section title="Sell rules">
          <Grid>
            <Num label="Stop-loss" value={form.stop_loss_pct} step="0.01" onChange={num("stop_loss_pct")} hint="drop fraction (0 = none)" />
            <Num label="Take-profit ×intrinsic" value={form.take_profit_vs_intrinsic} step="0.1" onChange={num("take_profit_vs_intrinsic")} hint="0 = none" />
          </Grid>
        </Section>

        {/* Risk */}
        <Section title="Risk & sizing">
          <Grid>
            <Num label="Max positions" value={form.max_positions} step="1" onChange={num("max_positions")} />
            <Num label="Max position size" value={form.max_position_size} step="0.01" onChange={num("max_position_size")} hint="fraction of portfolio" />
            <Num label="Cash buffer" value={form.cash_buffer_min} step="0.01" onChange={num("cash_buffer_min")} hint="fraction held as cash" />
            <div className="space-y-2">
              <Label>Position sizing</Label>
              <select value={form.position_sizing} onChange={(e) => set("position_sizing", e.target.value as CreateInvestorInput["position_sizing"])}
                className="h-11 w-full rounded-md border bg-background px-2 text-sm">
                <option value="equal">Equal weight</option>
                <option value="conviction">By conviction</option>
              </select>
            </div>
          </Grid>
        </Section>

        <Button type="submit" className="h-11 w-full" disabled={create.isPending}>
          {create.isPending ? <Spinner /> : <><Sparkles className="h-4 w-4" /> Create investor</>}
        </Button>
        <p className="text-center text-[11px] text-muted-foreground">
          You can backtest it, allocate to it, and watch it trade — just like the presets.
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
  label: string; value: number | undefined; onChange: (e: React.ChangeEvent<HTMLInputElement>) => void; step?: string; hint?: string;
}) {
  return (
    <div className="space-y-1.5">
      <Label>{label}</Label>
      <Input type="number" step={step} value={value ?? 0} onChange={onChange} className="h-11" />
      {hint && <p className="text-[11px] text-muted-foreground">{hint}</p>}
    </div>
  );
}
