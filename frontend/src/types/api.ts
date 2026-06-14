// TypeScript mirror of the backend's JSON contract (internal/web/dto). Keeping
// these in sync is what lets the compiler catch shape mismatches before runtime.

export interface User {
  id: string;
  name: string;
  email: string;
  role: string;
  email_verified: boolean;
  avatar_url?: string;
  bio?: string;
}

export interface AuthResponse {
  access_token: string;
  refresh_token: string;
  expires_at: number;
  user: User;
}

export interface Fundamentals {
  pe: number;
  forward_pe: number;
  pb: number;
  ps: number;
  peg: number;
  ev_ebitda: number;
  earnings_yield: number;
  fcf_yield: number;
  dividend_yield: number;
  eps_ttm: number;
  bvps: number;
  roe: number;
  roic: number;
  gross_margin: number;
  operating_margin: number;
  net_margin: number;
  debt_to_equity: number;
  current_ratio: number;
  interest_coverage: number;
  fcf_positive: boolean;
  revenue_growth_yoy: number;
  eps_growth_yoy: number;
  revenue_cagr_3y: number;
  eps_growth_5y: number;
  eps_positive_years: number;
  dividend_years: number;
  beta: number;
  market_cap: number;
}

export interface StockSummary {
  symbol: string;
  name: string;
  sector: string;
  asset_class: string;
  current_price: number;
  fundamentals: Fundamentals;
}

export interface Candle {
  time: number; // unix seconds
  open: number;
  high: number;
  low: number;
  close: number;
  volume: number;
}

export interface StockDetail {
  symbol: string;
  name: string;
  sector: string;
  exchange: string;
  asset_class: string;
  current_price: number;
  fundamentals: Fundamentals;
  candles: Candle[];
}

export interface NewsItem {
  title: string;
  url: string;
  authors: string[];
  summary: string;
  source: string;
  banner_image: string;
}

export interface HoldingDTO {
  symbol: string;
  name: string;
  quantity: number;
  avg_price: number;
  current_price: number;
  cost_basis: number;
  market_value: number;
  unrealized_pl: number;
  unrealized_pl_pct: number;
  allocation_pct: number;
}

export interface PortfolioStats {
  cash: number;
  holdings_value: number;
  total_value: number;
  cost_basis: number;
  unrealized_pl: number;
  unrealized_pl_pct: number;
  roi: number;
  holdings: HoldingDTO[];
}

export interface TradeResponse {
  trade_id: string;
  symbol: string;
  side: string;
  quantity: number;
  price: number;
  total_value: number;
  executed_at: number;
}

export interface TradeHistoryItem {
  trade_id: string;
  symbol: string;
  side: string;
  quantity: number;
  price: number;
  total_value: number;
  status: string;
  executed_at: number;
}

export interface Investor {
  id: string;
  name: string;
  bio: string;
  strategy: string;
  roi: number;
  rank: number;
  followers: number;
  created_by?: string; // creator's name for user-built investors; empty for presets
}

export interface Trader {
  rank: number;
  name: string;
  avatar_url: string;
  value: number;
  roi: number;
}

export interface Allocation {
  id: string;
  investor_id: string;
  investor_name: string;
  strategy: string;
  capital: number;
  cash: number;
  market_value: number;
  return_pct: number;
  is_active: boolean;
}

export interface AllocationHolding {
  symbol: string;
  quantity: number;
  avg_price: number;
  current_price: number;
  market_value: number;
  unrealized_pl: number;
  unrealized_pl_pct: number;
}

export interface AllocationTrade {
  symbol: string;
  side: string;
  quantity: number;
  price: number;
  total_value: number;
  executed_at: number;
}

export interface AllocationDetail extends Allocation {
  holdings: AllocationHolding[];
  trades: AllocationTrade[];
}

export interface BacktestPoint {
  date: number;
  value: number;
}

export interface BacktestTrade {
  date: number;
  side: string;
  symbol: string;
  quantity: number;
  price: number;
  total_value: number;
}

export interface BacktestHolding {
  symbol: string;
  quantity: number;
  avg_price: number;
  last_price: number;
  market_value: number;
  unrealized_pl: number;
}

export interface BacktestResult {
  investor_id: string;
  investor_name: string;
  strategy: string;
  start_cash: number;
  final_value: number;
  end_cash: number;
  roi: number;
  max_drawdown: number;
  win_rate: number;
  trade_count: number;
  buy_count: number;
  sell_count: number;
  held_count: number;
  start_date: number;
  end_date: number;
  equity: BacktestPoint[];
  trades: BacktestTrade[];
  holdings: BacktestHolding[];
  note: string;
}

export interface FeedItem {
  investor_id: string;
  investor_name: string;
  symbol: string;
  side: string;
  quantity: number;
  price: number;
  executed_at: number;
}
