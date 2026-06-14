import { Link } from "react-router-dom";
import { Trophy, Wand2, Users as UsersIcon } from "lucide-react";
import { useLeaderboard, useFollowing, useFollow, useUnfollow, useMyInvestors, useDeleteInvestor, useTraders } from "@/hooks/queries";
import { InvestorCard } from "@/components/investors/InvestorCard";
import { Loading, ErrorState, EmptyState } from "@/components/common/states";
import { Button } from "@/components/ui/button";
import { initials } from "@/lib/investorMeta";
import { pct, pnlColor } from "@/lib/format";
import { cn } from "@/lib/utils";

export default function Investors() {
  const { data, isLoading, isError, error } = useLeaderboard();
  const following = useFollowing();
  const follow = useFollow();
  const unfollow = useUnfollow();
  const mine = useMyInvestors();
  const traders = useTraders();
  const del = useDeleteInvestor();
  const removeMine = (id: string) => {
    if (window.confirm("Delete this investor? Any allocations to it are closed and the cash returned to you. This can't be undone.")) {
      del.mutate(id, {
        onError: (e) => window.alert(e instanceof Error ? e.message : "Could not delete investor"),
      });
    }
  };

  const followedIds = new Set((following.data ?? []).map((i) => i.id));
  const pendingId = follow.isPending ? follow.variables : unfollow.isPending ? unfollow.variables : undefined;

  const toggle = (id: string) => {
    if (followedIds.has(id)) unfollow.mutate(id);
    else follow.mutate(id);
  };

  return (
    <div className="space-y-6">
      <header className="flex flex-wrap items-start justify-between gap-3">
        <div>
          <h1 className="text-2xl font-semibold tracking-tight">Investors</h1>
          <p className="mt-1 text-sm text-muted-foreground">
            Famous strategies, running live on real data. Follow the ones you believe in — or build your own.
          </p>
        </div>
        <Button asChild>
          <Link to="/app/investors/new"><Wand2 className="h-4 w-4" /> Build your own</Link>
        </Button>
      </header>

      {/* Investors the user created */}
      {!!mine.data?.length && (
        <section className="rounded-2xl border bg-card p-5">
          <h2 className="mb-3 text-sm font-semibold text-muted-foreground">Your investors</h2>
          <div className="grid gap-4 sm:grid-cols-2">
            {mine.data.map((inv) => (
              <InvestorCard
                key={inv.id}
                investor={inv}
                following={followedIds.has(inv.id)}
                busy={pendingId === inv.id}
                onToggleFollow={toggle}
                onDelete={removeMine}
                deleting={del.isPending && del.variables === inv.id}
              />
            ))}
          </div>
        </section>
      )}

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

          {/* Side panels */}
          <div className="space-y-4 lg:sticky lg:top-7 h-fit">
            {/* Investor leaderboard */}
            <aside className="rounded-2xl border bg-card p-4">
              <div className="mb-3 flex items-center gap-2">
                <Trophy className="h-4 w-4 text-primary" />
                <h2 className="text-sm font-semibold">Investor leaderboard</h2>
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

            {/* Real-user leaderboard */}
            {!!traders.data?.length && (
              <aside className="rounded-2xl border bg-card p-4">
                <div className="mb-3 flex items-center gap-2">
                  <UsersIcon className="h-4 w-4 text-primary" />
                  <h2 className="text-sm font-semibold">Top traders</h2>
                </div>
                <ol className="space-y-1">
                  {traders.data.slice(0, 10).map((t) => (
                    <li key={t.rank} className="flex items-center gap-2.5 rounded-lg px-2 py-1.5 text-sm">
                      <span className={cn("w-5 text-center text-xs font-semibold", t.rank === 1 ? "text-primary" : "text-muted-foreground")}>{t.rank}</span>
                      {t.avatar_url ? (
                        <img src={t.avatar_url} alt={t.name} className="h-5 w-5 rounded-full object-cover" referrerPolicy="no-referrer" />
                      ) : (
                        <span className="flex h-5 w-5 items-center justify-center rounded-full bg-primary/12 text-[9px] font-semibold text-primary">{initials(t.name)}</span>
                      )}
                      <span className="min-w-0 flex-1 truncate">{t.name}</span>
                      <span className={cn("tabular-nums text-xs font-medium", pnlColor(t.roi))}>{pct(t.roi)}</span>
                    </li>
                  ))}
                </ol>
                <p className="mt-2 text-[11px] text-muted-foreground">Real users ranked by portfolio return.</p>
              </aside>
            )}
          </div>
        </div>
      )}
    </div>
  );
}
