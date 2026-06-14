// Presentation helpers — consistent money / percent / date formatting + the
// gain/loss color convention used across the app.

const usd = new Intl.NumberFormat("en-US", { style: "currency", currency: "USD" });
const usd0 = new Intl.NumberFormat("en-US", { style: "currency", currency: "USD", maximumFractionDigits: 0 });

export const money = (n: number): string => usd.format(n ?? 0);
export const moneyWhole = (n: number): string => usd0.format(n ?? 0);

export const qty = (n: number): string =>
  (n ?? 0).toLocaleString("en-US", { maximumFractionDigits: 4 });

/** A fraction (0.1234) rendered as a signed percentage ("+12.34%"). */
export const pct = (fraction: number, signed = true): string => {
  const v = (fraction ?? 0) * 100;
  const s = `${v.toFixed(2)}%`;
  return signed && v > 0 ? `+${s}` : s;
};

/** Tailwind text color for a P&L value: warm green up, warm red down, muted flat. */
export const pnlColor = (n: number): string =>
  n > 0 ? "text-gain" : n < 0 ? "text-loss" : "text-muted-foreground";

export const marketCap = (n: number): string => {
  if (!n) return "—";
  if (n >= 1e12) return `$${(n / 1e12).toFixed(2)}T`;
  if (n >= 1e9) return `$${(n / 1e9).toFixed(2)}B`;
  if (n >= 1e6) return `$${(n / 1e6).toFixed(1)}M`;
  return moneyWhole(n);
};

export const fromUnix = (sec: number): string =>
  new Date((sec ?? 0) * 1000).toLocaleString("en-US", {
    month: "short",
    day: "numeric",
    hour: "numeric",
    minute: "2-digit",
  });

export const ratio = (n: number): string => (n ? n.toFixed(2) : "—");
