import { Link } from "react-router-dom";
import { useFeed } from "@/hooks/queries";
import { Loading, ErrorState, EmptyState } from "@/components/common/states";
import { money, qty, fromUnix } from "@/lib/format";
import { cn } from "@/lib/utils";

export default function Feed() {
  const { data, isLoading, isError, error } = useFeed();

  return (
    <div className="space-y-6">
      <header>
        <h1 className="text-2xl font-semibold tracking-tight">Feed</h1>
        <p className="mt-1 text-sm text-muted-foreground">Latest moves from the investors you follow.</p>
      </header>

      {isLoading ? (
        <Loading />
      ) : isError ? (
        <ErrorState message={(error as Error)?.message} />
      ) : !data?.length ? (
        <EmptyState
          title="Your feed is quiet"
          hint="Follow investors from the Investors tab to see their trades here."
        />
      ) : (
        <ul className="space-y-2.5">
          {data.map((item, i) => (
            <li key={i} className="flex items-center justify-between rounded-xl border bg-card px-4 py-3">
              <div className="flex items-center gap-3">
                <span className={cn("rounded px-1.5 py-0.5 text-[10px] font-semibold uppercase", item.side === "buy" ? "bg-gain/15 text-gain" : "bg-loss/15 text-loss")}>{item.side}</span>
                <div>
                  <p className="text-sm">
                    <span className="font-medium">{item.investor_name}</span>
                    <span className="text-muted-foreground"> {item.side === "buy" ? "bought" : "sold"} </span>
                    <Link to={`/app/stock/${item.symbol}`} className="font-medium hover:text-primary">{item.symbol}</Link>
                  </p>
                  <p className="text-xs text-muted-foreground">{qty(item.quantity)} @ {money(item.price)}</p>
                </div>
              </div>
              <span className="text-xs text-muted-foreground">{fromUnix(item.executed_at)}</span>
            </li>
          ))}
        </ul>
      )}
    </div>
  );
}
