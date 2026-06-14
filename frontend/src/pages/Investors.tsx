import { Link } from "react-router-dom";
import { Trophy } from "lucide-react";
import { useLeaderboard, useFollowing, useFollow, useUnfollow } from "@/hooks/queries";
import { InvestorCard } from "@/components/investors/InvestorCard";
import { Loading, ErrorState, EmptyState } from "@/components/common/states";
import { pct, pnlColor } from "@/lib/format";
import { cn } from "@/lib/utils";

export default function Investors() {
  const { data, isLoading, isError, error } = useLeaderboard();
  const following = useFollowing();
  const follow = useFollow();
  const unfollow = useUnfollow();

  const followedIds = new Set((following.data ?? []).map((i) => i.id));
  const pendingId = follow.isPending ? follow.variables : unfollow.isPending ? unfollow.variables : undefined;

  const toggle = (id: string) => {
    if (followedIds.has(id)) unfollow.mutate(id);
    else follow.mutate(id);
  };

  return (
    <div className="space-y-6">
      <header>
        <h1 className="text-2xl font-semibold tracking-tight">Investors</h1>
        <p className="mt-1 text-sm text-muted-foreground">
          Famous strategies, running live on real data. Follow the ones you believe in.
        </p>
      </header>

      {isLoading ? (
        <Loading />
      ) : isError ? (
        <ErrorState message={(error as Error)?.message} />
      ) : !data?.length ? (
        <EmptyState title="No investors yet" hint="Start the engine to provision the bot investors." />
      ) : (
        <div className="grid gap-6 lg:grid-cols-[1fr_280px]">
          {/* Cards */}
          <div className="grid gap-4 sm:grid-cols-2 xl:grid-cols-2">
            {data.map((inv) => (
              <InvestorCard
                key={inv.id}
                investor={inv}
                following={followedIds.has(inv.id)}
                busy={pendingId === inv.id}
                onToggleFollow={toggle}
              />
            ))}
          </div>

          {/* Leaderboard side panel */}
          <aside className="lg:sticky lg:top-7 h-fit rounded-2xl border bg-card p-4">
            <div className="mb-3 flex items-center gap-2">
              <Trophy className="h-4 w-4 text-primary" />
              <h2 className="text-sm font-semibold">Leaderboard</h2>
            </div>
            <ol className="space-y-1">
              {[...data]
                .sort((a, b) => a.rank - b.rank)
                .slice(0, 10)
                .map((inv) => (
                  <li key={inv.id}>
                    <Link
                      to={`/app/investors/${inv.id}`}
                      className="flex items-center gap-2.5 rounded-lg px-2 py-1.5 text-sm hover:bg-accent/50"
                    >
                      <span className={cn("w-5 text-center text-xs font-semibold", inv.rank === 1 ? "text-primary" : "text-muted-foreground")}>
                        {inv.rank}
                      </span>
                      <span className="min-w-0 flex-1 truncate">{inv.name}</span>
                      <span className={cn("tabular-nums text-xs font-medium", pnlColor(inv.roi))}>{pct(inv.roi)}</span>
                    </Link>
                  </li>
                ))}
            </ol>
          </aside>
        </div>
      )}
    </div>
  );
}
