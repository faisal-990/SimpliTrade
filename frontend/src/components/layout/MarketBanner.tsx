import { useEffect, useState } from "react";
import { Activity, Clock } from "lucide-react";
import { useMarketStatus } from "@/hooks/queries";

// countdown formats seconds-until as "3h 12m" / "12m" / "<1m".
function countdown(untilUnix: number): string {
  const secs = Math.max(0, untilUnix - Math.floor(Date.now() / 1000));
  const h = Math.floor(secs / 3600);
  const m = Math.floor((secs % 3600) / 60);
  if (h > 0) return `${h}h ${m}m`;
  if (m > 0) return `${m}m`;
  return "<1m";
}

// MarketBanner is the always-visible signal that the live engine is (or isn't)
// running — green + pulsing while US markets are open, calm when closed. Makes
// it obvious "something is happening" at the open, and easy to spot if it isn't.
export function MarketBanner() {
  const { data } = useMarketStatus();
  const [, tick] = useState(0);

  // Re-render every 30s so the countdown stays live between refetches.
  useEffect(() => {
    const id = setInterval(() => tick((n) => n + 1), 30_000);
    return () => clearInterval(id);
  }, []);

  if (!data) return null;
  const left = data.next_change ? countdown(data.next_change) : null;

  if (data.open) {
    return (
      <div className="mb-5 flex items-center gap-2.5 rounded-xl border border-gain/30 bg-gain/10 px-3.5 py-2 text-sm text-gain">
        <span className="relative flex h-2.5 w-2.5">
          <span className="absolute inline-flex h-full w-full animate-ping rounded-full bg-gain opacity-60" />
          <span className="relative inline-flex h-2.5 w-2.5 rounded-full bg-gain" />
        </span>
        <Activity className="h-4 w-4" />
        <span className="font-medium text-foreground">Markets open</span>
        <span className="text-muted-foreground">— the engine is trading live{left && ` · closes in ${left}`}.</span>
      </div>
    );
  }

  return (
    <div className="mb-5 flex items-center gap-2.5 rounded-xl border bg-muted/40 px-3.5 py-2 text-sm text-muted-foreground">
      <span className="h-2.5 w-2.5 rounded-full bg-muted-foreground/40" />
      <Clock className="h-4 w-4" />
      <span className="font-medium text-foreground">Markets closed</span>
      {left && <span>— opens in {left}.</span>}
      <span className="ml-auto hidden text-xs sm:inline">US session 9:30 AM–4:00 PM ET</span>
    </div>
  );
}
