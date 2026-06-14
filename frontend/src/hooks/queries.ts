import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { api } from "@/lib/api";
import type {
  Allocation,
  AllocationDetail,
  BacktestResult,
  FeedItem,
  Investor,
  NewsItem,
  PortfolioStats,
  StockDetail,
  StockSummary,
  TradeHistoryItem,
  TradeResponse,
  Trader,
  User,
} from "@/types/api";

// Centralized query keys — single source of truth for caching/invalidation.
export const qk = {
  stocks: ["stocks"] as const,
  stock: (s: string) => ["stock", s] as const,
  news: ["news"] as const,
  portfolioStats: ["portfolio", "stats"] as const,
  holdings: ["portfolio", "holdings"] as const,
  tradeHistory: ["trade", "history"] as const,
  leaderboard: ["investors"] as const,
  investor: (id: string) => ["investor", id] as const,
  investorTrades: (id: string) => ["investor", id, "trades"] as const,
  following: ["following"] as const,
  feed: ["feed"] as const,
  allocations: ["allocations"] as const,
  allocation: (id: string) => ["allocation", id] as const,
};

// --- market / dashboard ---
export const useStocks = () =>
  useQuery({ queryKey: qk.stocks, queryFn: () => api.get<StockSummary[]>("/dashboard/fundamentals?limit=100") });

export const useStock = (symbol: string) =>
  useQuery({ queryKey: qk.stock(symbol), queryFn: () => api.get<StockDetail>(`/dashboard/graph/${symbol}`), enabled: !!symbol });

// News refreshes on a 10-minute cadence. staleTime matches the interval so we
// don't refetch more often than that (and a real, rate-limited provider behind
// the endpoint stays within budget). Today the endpoint serves an embedded seed,
// so headlines won't change until a live news source is wired in server-side.
const NEWS_REFRESH_MS = 30 * 60 * 1000;
export const useNews = () =>
  useQuery({
    queryKey: qk.news,
    queryFn: () => api.get<NewsItem[]>("/dashboard/news"),
    refetchInterval: NEWS_REFRESH_MS,
    staleTime: NEWS_REFRESH_MS,
    refetchOnWindowFocus: false,
  });

// --- portfolio ---
export const usePortfolioStats = () =>
  useQuery({ queryKey: qk.portfolioStats, queryFn: () => api.get<PortfolioStats>("/portfolio/stats") });

export const useHoldings = () =>
  useQuery({ queryKey: qk.holdings, queryFn: () => api.get<PortfolioStats["holdings"]>("/portfolio/") });

export const useTradeHistory = () =>
  useQuery({ queryKey: qk.tradeHistory, queryFn: () => api.get<TradeHistoryItem[]>("/trade/history") });

// --- social ---
export const useLeaderboard = () =>
  useQuery({ queryKey: qk.leaderboard, queryFn: () => api.get<Investor[]>("/investor/") });

export const useInvestor = (id: string) =>
  useQuery({ queryKey: qk.investor(id), queryFn: () => api.get<Investor>(`/investor/${id}`), enabled: !!id });

export const useInvestorTrades = (id: string) =>
  useQuery({ queryKey: qk.investorTrades(id), queryFn: () => api.get<TradeHistoryItem[]>(`/investor/${id}/trades`), enabled: !!id });

// useBacktest replays an investor's strategy over historical prices on demand.
export function useBacktest() {
  return useMutation({
    mutationFn: ({ investorId, days, cash }: { investorId: string; days: number; cash: number }) =>
      api.get<BacktestResult>(`/investor/${investorId}/backtest?days=${days}&cash=${cash}`),
  });
}

export const useFollowing = () =>
  useQuery({ queryKey: qk.following, queryFn: () => api.get<Investor[]>("/following") });

export const useFeed = () => useQuery({ queryKey: qk.feed, queryFn: () => api.get<FeedItem[]>("/feed") });

// --- social: real-user leaderboard ---
export const useTraders = () =>
  useQuery({ queryKey: ["traders"] as const, queryFn: () => api.get<Trader[]>("/traders") });

// --- mutations ---
interface TradeInput {
  symbol: string;
  quantity: number;
}

export function useTrade(side: "buy" | "sell") {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (input: TradeInput) => api.post<TradeResponse>(`/trade/${side}`, input),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: qk.portfolioStats });
      qc.invalidateQueries({ queryKey: qk.holdings });
      qc.invalidateQueries({ queryKey: qk.tradeHistory });
    },
  });
}

// --- custom ("build your own") investors ---
export interface CreateInvestorInput {
  name: string;
  philosophy: string;
  approach: "value" | "quality" | "growth" | "momentum";
  max_positions: number;
  pe_max?: number;
  pb_max?: number;
  roe_min?: number;
  operating_margin_min?: number;
  revenue_growth_min?: number;
  eps_growth_min?: number;
  return_6m_min?: number;
  stop_loss_pct?: number;
  take_profit_vs_intrinsic?: number;
  max_position_size: number;
  cash_buffer_min: number;
  position_sizing: "equal" | "conviction";
}

export const qkMyInvestors = ["custom-investors"] as const;

export const useMyInvestors = () =>
  useQuery({ queryKey: qkMyInvestors, queryFn: () => api.get<Investor[]>("/custom-investors/") });

export function useCreateInvestor() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (input: CreateInvestorInput) => api.post<Investor>("/custom-investors/", input),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: qkMyInvestors });
      qc.invalidateQueries({ queryKey: qk.leaderboard });
    },
  });
}

export function useDeleteInvestor() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => api.del(`/custom-investors/${id}`),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: qkMyInvestors });
      qc.invalidateQueries({ queryKey: qk.leaderboard });
      qc.invalidateQueries({ queryKey: qk.following });
    },
  });
}

// --- profile ---
export function useUpdateProfile() {
  return useMutation({
    mutationFn: (input: { name: string; bio: string }) => api.put<User>("/auth/me", input),
  });
}

// --- password reset (public, OTP via email) ---
export function useForgotPassword() {
  return useMutation({
    mutationFn: (email: string) => api.post("/auth/forgot-password", { email }),
  });
}

export function useResetPassword() {
  return useMutation({
    mutationFn: (input: { email: string; code: string; password: string }) =>
      api.post("/auth/reset-password", input),
  });
}

// --- dev / admin controls ---
export function useSimulateMarket() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: () => api.post("/admin/simulate"),
    onSuccess: () => qc.invalidateQueries(), // a cycle touches everything
  });
}

export function useResetAccount() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: () => api.post("/admin/reset"),
    onSuccess: () => qc.invalidateQueries(),
  });
}

export function useSellAll() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: () => api.post<{ sold: number }>("/trade/sell-all"),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: qk.portfolioStats });
      qc.invalidateQueries({ queryKey: qk.holdings });
      qc.invalidateQueries({ queryKey: qk.tradeHistory });
    },
  });
}

function invalidateSocial(qc: ReturnType<typeof useQueryClient>, investorId?: string) {
  qc.invalidateQueries({ queryKey: qk.following });
  qc.invalidateQueries({ queryKey: qk.feed });
  qc.invalidateQueries({ queryKey: qk.leaderboard }); // follower counts
  if (investorId) qc.invalidateQueries({ queryKey: qk.investor(investorId) });
}

// --- copy-trading allocations ---
export const useAllocations = () =>
  useQuery({ queryKey: qk.allocations, queryFn: () => api.get<Allocation[]>("/allocations/") });

// useAllocationDetail loads what the bot did with one allocation's capital:
// its holdings + recent trades. Pass enabled=false to defer until expanded.
export const useAllocationDetail = (id: string, enabled = true) =>
  useQuery({
    queryKey: qk.allocation(id),
    queryFn: () => api.get<AllocationDetail>(`/allocations/${id}`),
    enabled: enabled && !!id,
  });

export function useCreateAllocation() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (input: { investor_id: string; capital: number }) =>
      api.post<Allocation>("/allocations/", input),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: qk.allocations });
      qc.invalidateQueries({ queryKey: qk.portfolioStats });
    },
  });
}

export function useStopAllocation() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => api.del(`/allocations/${id}`),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: qk.allocations });
      qc.invalidateQueries({ queryKey: qk.portfolioStats });
    },
  });
}

export function useFollow() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => api.post(`/investor/${id}/follow`),
    onSuccess: (_data, id) => invalidateSocial(qc, id),
  });
}

export function useUnfollow() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => api.del(`/investor/${id}/follow`),
    onSuccess: (_data, id) => invalidateSocial(qc, id),
  });
}
