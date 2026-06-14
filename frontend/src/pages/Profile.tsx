import { useState } from "react";
import { BadgeCheck, Save } from "lucide-react";
import { useAuth } from "@/auth/useAuth";
import { useUpdateProfile } from "@/hooks/queries";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Spinner } from "@/components/common/states";
import { ApiError } from "@/lib/api";
import { initials } from "@/lib/investorMeta";
import { cn } from "@/lib/utils";

export default function Profile() {
  const { user, setUser } = useAuth();
  const update = useUpdateProfile();
  const [name, setName] = useState(user?.name ?? "");
  const [bio, setBio] = useState(user?.bio ?? "");
  const [error, setError] = useState<string | null>(null);
  const [saved, setSaved] = useState(false);

  if (!user) return null;

  const dirty = name.trim() !== user.name || (bio ?? "") !== (user.bio ?? "");

  const save = (e: React.FormEvent) => {
    e.preventDefault();
    setError(null);
    setSaved(false);
    update.mutate(
      { name: name.trim(), bio },
      {
        onSuccess: (updated) => {
          setUser(updated);
          setSaved(true);
        },
        onError: (err) => setError(err instanceof ApiError ? err.message : "Could not save profile"),
      }
    );
  };

  return (
    <div className="mx-auto max-w-2xl space-y-7">
      <header>
        <h1 className="text-2xl font-semibold tracking-tight">Profile</h1>
        <p className="mt-1 text-sm text-muted-foreground">Your account and how you appear in SimpliTrade.</p>
      </header>

      {/* Identity card */}
      <section className="flex items-center gap-4 rounded-2xl border bg-card p-6">
        {user.avatar_url ? (
          <img src={user.avatar_url} alt={user.name} className="h-16 w-16 rounded-full object-cover ring-1 ring-border" referrerPolicy="no-referrer" />
        ) : (
          <span className="flex h-16 w-16 items-center justify-center rounded-full bg-primary/15 text-xl font-semibold text-primary">
            {initials(user.name)}
          </span>
        )}
        <div className="min-w-0">
          <p className="truncate text-lg font-semibold">{user.name}</p>
          <p className="flex items-center gap-1.5 truncate text-sm text-muted-foreground">
            {user.email}
            {user.email_verified && <BadgeCheck className="h-4 w-4 text-gain" aria-label="verified" />}
          </p>
        </div>
      </section>

      {/* About me (editable) */}
      <form onSubmit={save} className="space-y-5 rounded-2xl border bg-card p-6">
        <h2 className="text-sm font-semibold text-muted-foreground">About me</h2>

        {error && <div className="rounded-lg border border-loss/30 bg-loss/10 px-3 py-2 text-sm text-loss">{error}</div>}
        {saved && !dirty && (
          <div className="rounded-lg border border-gain/30 bg-gain/10 px-3 py-2 text-sm text-gain">Profile saved.</div>
        )}

        <div className="space-y-2">
          <Label htmlFor="name">Display name</Label>
          <Input id="name" value={name} onChange={(e: React.ChangeEvent<HTMLInputElement>) => setName(e.target.value)} className="h-11" maxLength={100} />
        </div>

        <div className="space-y-2">
          <Label htmlFor="bio">Bio</Label>
          <textarea
            id="bio"
            value={bio}
            onChange={(e: React.ChangeEvent<HTMLTextAreaElement>) => setBio(e.target.value.slice(0, 500))}
            rows={4}
            placeholder="A few words about your investing style…"
            className="w-full rounded-md border bg-background px-3 py-2 text-sm outline-none ring-offset-background placeholder:text-muted-foreground focus-visible:ring-2 focus-visible:ring-ring"
          />
          <p className="text-right text-[11px] text-muted-foreground">{bio.length}/500</p>
        </div>

        <Button type="submit" className={cn("h-11", !dirty && "opacity-60")} disabled={update.isPending || !dirty || !name.trim()}>
          {update.isPending ? <Spinner /> : <><Save className="h-4 w-4" /> Save changes</>}
        </Button>
      </form>
    </div>
  );
}
