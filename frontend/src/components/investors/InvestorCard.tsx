import { Link } from "react-router-dom";
import { UserPlus, UserCheck, ArrowUpRight, Users, Trash2 } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Spinner } from "@/components/common/states";
import { pct, pnlColor } from "@/lib/format";
import { strategyLabel, riskOf, riskChipClass, initials } from "@/lib/investorMeta";
import { cn } from "@/lib/utils";
import type { Investor } from "@/types/api";

interface Props {
  investor: Investor;
  following: boolean;
  busy?: boolean;
  onToggleFollow: (id: string) => void;
  // When provided (only for investors the user created), shows a delete action.
  onDelete?: (id: string) => void;
  deleting?: boolean;
}

// Rich investor card: avatar, name, strategy + risk chips, rank, big ROI, bio
// snippet, follow toggle, and a link into the full profile.
export function InvestorCard({ investor, following, busy, onToggleFollow, onDelete, deleting }: Props) {
  const risk = riskOf(investor.strategy);
  return (
    <div className="group flex flex-col rounded-2xl border bg-card p-5 transition-colors hover:border-primary/40">
      <div className="flex items-start gap-3">
        <span className="flex h-11 w-11 shrink-0 items-center justify-center rounded-full bg-primary/12 text-sm font-semibold text-primary">
          {initials(investor.name)}
        </span>
        <div className="min-w-0 flex-1">
          <div className="flex items-center gap-2">
            <Link to={`/app/investors/${investor.id}`} className="truncate font-semibold hover:text-primary">
              {investor.name}
            </Link>
            {investor.rank > 0 && investor.rank < 1_000_000 && (
              <span className="rounded bg-muted px-1.5 py-0.5 text-[10px] font-semibold text-muted-foreground">#{investor.rank}</span>
            )}
          </div>
          <div className="mt-1 flex flex-wrap items-center gap-1.5">
            <span className="rounded bg-secondary px-1.5 py-0.5 text-[10px] font-medium text-secondary-foreground">{strategyLabel(investor.strategy)}</span>
            <span className={cn("rounded px-1.5 py-0.5 text-[10px] font-medium", riskChipClass(risk.level))}>{risk.label}</span>
          </div>
        </div>
        <div className="text-right">
          <p className={cn("text-lg font-semibold tabular-nums leading-none", pnlColor(investor.roi))}>{pct(investor.roi)}</p>
          <p className="mt-1 text-[10px] uppercase tracking-wide text-muted-foreground">ROI</p>
        </div>
      </div>

      <p className="mt-3 line-clamp-2 text-sm leading-relaxed text-muted-foreground">{investor.bio}</p>

      <p className="mt-2 flex items-center gap-1.5 text-xs text-muted-foreground">
        <Users className="h-3.5 w-3.5" />
        {investor.followers.toLocaleString()} {investor.followers === 1 ? "follower" : "followers"}
      </p>

      <div className="mt-4 flex items-center gap-2">
        <Button
          size="sm"
          variant={following ? "outline" : "default"}
          className="flex-1"
          disabled={busy}
          onClick={() => onToggleFollow(investor.id)}
        >
          {busy ? <Spinner /> : following ? <><UserCheck className="h-4 w-4" /> Following</> : <><UserPlus className="h-4 w-4" /> Follow</>}
        </Button>
        <Button asChild size="sm" variant="ghost">
          <Link to={`/app/investors/${investor.id}`}>
            View <ArrowUpRight className="h-4 w-4" />
          </Link>
        </Button>
        {onDelete && (
          <Button
            size="sm"
            variant="ghost"
            className="text-muted-foreground hover:text-loss"
            disabled={deleting}
            title="Delete this investor"
            onClick={() => onDelete(investor.id)}
          >
            {deleting ? <Spinner /> : <Trash2 className="h-4 w-4" />}
          </Button>
        )}
      </div>
    </div>
  );
}
