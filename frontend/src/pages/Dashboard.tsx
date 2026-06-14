import { Link } from "react-router-dom";
import { ExternalLink, ArrowUpRight } from "lucide-react";
import { useNews, useStocks } from "@/hooks/queries";
import { TickerTape } from "@/components/market/TradingView";
import { Loading, ErrorState, EmptyState } from "@/components/common/states";
import { money, marketCap } from "@/lib/format";
import type { StockSummary } from "@/types/api";

const TAPE = [
  { proName: "NASDAQ:AAPL", title: "Apple" },
  { proName: "NASDAQ:MSFT", title: "Microsoft" },
  { proName: "NASDAQ:NVDA", title: "NVIDIA" },
  { proName: "NASDAQ:AMD", title: "AMD" },
  { proName: "NYSE:ORCL", title: "Oracle" },
  { proName: "SP:SPX", title: "S&P 500" },
];

export default function Dashboard() {
  const stocks = useStocks();
  const news = useNews();

  return (
    <div className="space-y-7">
      <header>
        <h1 className="text-2xl font-semibold tracking-tight">Market</h1>
        <p className="mt-1 text-sm text-muted-foreground">Live prices and headlines. Pick a stock to chart and trade it.</p>
      </header>

      <div className="rounded-xl border bg-card p-1">
        <TickerTape symbols={TAPE} />
      </div>

      <div className="grid gap-7 lg:grid-cols-[1.6fr_1fr]">
        {/* Stock universe */}
        <section>
          <h2 className="mb-3 text-sm font-semibold text-muted-foreground">Stocks</h2>
          {stocks.isLoading ? (
            <Loading />
          ) : stocks.isError ? (
            <ErrorState message={(stocks.error as Error)?.message} />
          ) : !stocks.data?.length ? (
            <EmptyState title="No stocks yet" hint="Seed the universe to populate the market." />
          ) : (
            <div className="grid gap-3 sm:grid-cols-2">
              {stocks.data.map((s) => (
                <StockCard key={s.symbol} stock={s} />
              ))}
            </div>
          )}
        </section>

        {/* News */}
        <section>
          <h2 className="mb-3 text-sm font-semibold text-muted-foreground">Market news</h2>
          {news.isLoading ? (
            <Loading />
          ) : news.isError ? (
            <ErrorState message={(news.error as Error)?.message} />
          ) : !news.data?.length ? (
            <EmptyState title="No news right now" />
          ) : (
            <ul className="space-y-3">
              {news.data.slice(0, 8).map((n, i) => (
                <li key={i} className="rounded-xl border bg-card p-3.5">
                  <a href={n.url} target="_blank" rel="noreferrer" className="group flex items-start justify-between gap-3">
                    <div>
                      <p className="text-sm font-medium leading-snug group-hover:text-primary">{n.title}</p>
                      <p className="mt-1 text-xs text-muted-foreground">{n.source}</p>
                    </div>
                    <ExternalLink className="mt-0.5 h-4 w-4 shrink-0 text-muted-foreground group-hover:text-primary" />
                  </a>
                </li>
              ))}
            </ul>
          )}
        </section>
      </div>
    </div>
  );
}

function StockCard({ stock }: { stock: StockSummary }) {
  return (
    <Link
      to={`/app/stock/${stock.symbol}`}
      className="group flex items-center justify-between rounded-xl border bg-card p-4 transition-colors hover:border-primary/40 hover:bg-accent/40"
    >
      <div className="min-w-0">
        <div className="flex items-center gap-2">
          <span className="font-semibold">{stock.symbol}</span>
          <span className="rounded bg-muted px-1.5 py-0.5 text-[10px] font-medium text-muted-foreground">{stock.sector}</span>
        </div>
        <p className="mt-0.5 truncate text-xs text-muted-foreground">{stock.name}</p>
        <p className="mt-1 text-[11px] text-muted-foreground">Mkt cap {marketCap(stock.fundamentals?.market_cap)}</p>
      </div>
      <div className="text-right">
        <p className="font-semibold tabular-nums">{money(stock.current_price)}</p>
        <ArrowUpRight className="ml-auto mt-1 h-4 w-4 text-muted-foreground group-hover:text-primary" />
      </div>
    </Link>
  );
}
