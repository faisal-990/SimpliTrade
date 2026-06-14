// Display metadata for investor strategies — labels + a coarse risk read derived
// from the strategy style (the backend doesn't expose risk_rating yet).

export const STRATEGY_LABEL: Record<string, string> = {
  deep_value: "Deep Value",
  quality_value: "Quality Value",
  quality_concentrated: "Quality · Concentrated",
  quality_growth: "Quality Growth",
  growth_quality: "Growth",
  garp: "GARP",
  disruptive_growth: "Disruptive Growth",
  high_conviction_value: "High Conviction",
  quant_value: "Quant · Magic Formula",
  absolute_value: "Absolute Value",
  global_contrarian_value: "Global Contrarian",
  focused_value: "Focused Value",
  activist_concentrated: "Activist",
  activist_value: "Activist Value",
  contrarian_deep_value: "Contrarian",
  distressed_contrarian: "Distressed",
  macro_risk_parity: "Macro · Risk Parity",
  macro_momentum: "Macro Momentum",
  trend_momentum: "Trend · Momentum",
};

export const strategyLabel = (style: string): string => STRATEGY_LABEL[style] ?? style;

type Risk = { label: string; level: 1 | 2 | 3 };

// Coarse risk tier per style → drives a small colored chip on cards.
const RISK: Record<string, Risk> = {
  deep_value: { label: "Conservative", level: 1 },
  absolute_value: { label: "Conservative", level: 1 },
  quality_value: { label: "Conservative", level: 1 },
  macro_risk_parity: { label: "Conservative", level: 1 },
  quant_value: { label: "Balanced", level: 2 },
  quality_concentrated: { label: "Balanced", level: 2 },
  quality_growth: { label: "Balanced", level: 2 },
  global_contrarian_value: { label: "Balanced", level: 2 },
  garp: { label: "Balanced", level: 2 },
  growth_quality: { label: "Balanced", level: 2 },
  distressed_contrarian: { label: "Balanced", level: 2 },
  focused_value: { label: "Aggressive", level: 3 },
  activist_concentrated: { label: "Aggressive", level: 3 },
  activist_value: { label: "Aggressive", level: 3 },
  contrarian_deep_value: { label: "Aggressive", level: 3 },
  high_conviction_value: { label: "Aggressive", level: 3 },
  disruptive_growth: { label: "Aggressive", level: 3 },
  macro_momentum: { label: "Aggressive", level: 3 },
  trend_momentum: { label: "Aggressive", level: 3 },
};

export const riskOf = (style: string): Risk => RISK[style] ?? { label: "Balanced", level: 2 };

export const riskChipClass = (level: 1 | 2 | 3): string =>
  level === 1
    ? "bg-gain/15 text-gain"
    : level === 2
      ? "bg-primary/15 text-primary"
      : "bg-loss/15 text-loss";

// Initials for an avatar tile.
export const initials = (name: string): string =>
  name.split(" ").map((p) => p[0]).slice(0, 2).join("").toUpperCase();
