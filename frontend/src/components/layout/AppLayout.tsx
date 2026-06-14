import { useState } from "react";
import { NavLink, Outlet, useNavigate } from "react-router-dom";
import {
  LayoutDashboard,
  Briefcase,
  Trophy,
  Rss,
  LogOut,
  Menu,
  X,
  PanelLeftClose,
  PanelLeftOpen,
  Zap,
  CircleUser,
} from "lucide-react";
import { Brand, LogoMark } from "@/components/Brand";
import { Button } from "@/components/ui/button";
import { Spinner } from "@/components/common/states";
import { useAuth } from "@/auth/useAuth";
import { useSimulateMarket } from "@/hooks/queries";
import { cn } from "@/lib/utils";

// SimulateButton triggers one engine cycle (bots + your allocations trade).
function SimulateButton({ collapsed }: { collapsed?: boolean }) {
  const sim = useSimulateMarket();
  return (
    <Button
      size="sm"
      className={cn("w-full", collapsed && "px-0")}
      disabled={sim.isPending}
      onClick={() => sim.mutate()}
      title="Run one market cycle"
    >
      {sim.isPending ? <Spinner /> : <Zap className="h-4 w-4" />}
      {!collapsed && (sim.isPending ? "Simulating…" : "Simulate market")}
    </Button>
  );
}

const NAV = [
  { to: "/app/dashboard", label: "Dashboard", icon: LayoutDashboard },
  { to: "/app/portfolio", label: "Portfolio", icon: Briefcase },
  { to: "/app/investors", label: "Investors", icon: Trophy },
  { to: "/app/feed", label: "Feed", icon: Rss },
  { to: "/app/profile", label: "Profile", icon: CircleUser },
];

const COLLAPSE_KEY = "simplitrade.sidebar.collapsed";

function initials(name?: string): string {
  if (!name) return "U";
  return name.split(" ").map((p) => p[0]).slice(0, 2).join("").toUpperCase();
}

function NavItems({ collapsed, onNavigate }: { collapsed?: boolean; onNavigate?: () => void }) {
  return (
    <nav className="flex flex-col gap-1">
      {NAV.map(({ to, label, icon: Icon }) => (
        <NavLink
          key={to}
          to={to}
          onClick={onNavigate}
          title={collapsed ? label : undefined}
          className={({ isActive }) =>
            cn(
              "flex items-center gap-3 rounded-lg px-3 py-2 text-sm font-medium transition-colors",
              collapsed && "justify-center px-2",
              isActive
                ? "bg-primary/10 text-primary"
                : "text-muted-foreground hover:bg-accent hover:text-accent-foreground"
            )
          }
        >
          <Icon className="h-[18px] w-[18px] shrink-0" />
          {!collapsed && label}
        </NavLink>
      ))}
    </nav>
  );
}

export function AppLayout() {
  const { user, logout } = useAuth();
  const navigate = useNavigate();
  const [mobileOpen, setMobileOpen] = useState(false);
  const [collapsed, setCollapsed] = useState(() => localStorage.getItem(COLLAPSE_KEY) === "1");

  const toggleCollapse = () => {
    setCollapsed((v) => {
      const next = !v;
      localStorage.setItem(COLLAPSE_KEY, next ? "1" : "0");
      return next;
    });
  };

  const handleLogout = async () => {
    await logout();
    navigate("/login", { replace: true });
  };

  return (
    <div className="min-h-svh bg-background">
      {/* Sidebar (desktop) */}
      <aside
        className={cn(
          "fixed inset-y-0 left-0 hidden flex-col border-r bg-card/60 px-3 py-5 transition-[width] duration-200 lg:flex",
          collapsed ? "w-16" : "w-64"
        )}
      >
        <div className={cn("flex items-center", collapsed ? "justify-center" : "justify-between px-2")}>
          {collapsed ? <LogoMark className="h-8 w-8" /> : <Brand />}
        </div>

        <div className="mt-8 flex-1">
          <NavItems collapsed={collapsed} />
        </div>

        <div className="mb-3">
          <SimulateButton collapsed={collapsed} />
        </div>

        <button
          onClick={toggleCollapse}
          className="mb-2 flex items-center justify-center gap-2 rounded-lg px-2 py-1.5 text-xs text-muted-foreground hover:bg-accent hover:text-accent-foreground"
          title={collapsed ? "Expand" : "Collapse"}
        >
          {collapsed ? <PanelLeftOpen className="h-4 w-4" /> : <><PanelLeftClose className="h-4 w-4" /> Collapse</>}
        </button>

        {!collapsed && <UserCard name={user?.name} email={user?.email} avatarUrl={user?.avatar_url} onLogout={handleLogout} />}
        {collapsed && (
          <button onClick={handleLogout} title="Sign out" className="flex h-9 items-center justify-center rounded-lg text-muted-foreground hover:bg-accent">
            <LogOut className="h-4 w-4" />
          </button>
        )}
      </aside>

      {/* Mobile top bar */}
      <div className="flex items-center justify-between border-b px-4 py-3 lg:hidden">
        <Brand />
        <Button variant="ghost" size="icon" onClick={() => setMobileOpen((v) => !v)} aria-label="Menu">
          {mobileOpen ? <X className="h-5 w-5" /> : <Menu className="h-5 w-5" />}
        </Button>
      </div>
      {mobileOpen && (
        <div className="border-b bg-card px-4 py-3 lg:hidden">
          <NavItems onNavigate={() => setMobileOpen(false)} />
          <div className="mt-3">
            <SimulateButton />
          </div>
          <div className="mt-3">
            <UserCard name={user?.name} email={user?.email} avatarUrl={user?.avatar_url} onLogout={handleLogout} />
          </div>
        </div>
      )}

      {/* Content */}
      <main className={cn("transition-[padding] duration-200", collapsed ? "lg:pl-16" : "lg:pl-64")}>
        <div className="mx-auto max-w-7xl px-5 py-7 sm:px-8">
          <Outlet />
        </div>
      </main>
    </div>
  );
}

function UserCard({ name, email, avatarUrl, onLogout }: { name?: string; email?: string; avatarUrl?: string; onLogout: () => void }) {
  return (
    <div className="rounded-xl border bg-background/60 p-3">
      <NavLink to="/app/profile" className="flex items-center gap-3 rounded-lg hover:opacity-90">
        {avatarUrl ? (
          <img src={avatarUrl} alt={name ?? "User"} className="h-9 w-9 rounded-full object-cover" referrerPolicy="no-referrer" />
        ) : (
          <span className="flex h-9 w-9 items-center justify-center rounded-full bg-primary/15 text-sm font-semibold text-primary">
            {initials(name)}
          </span>
        )}
        <div className="min-w-0 flex-1">
          <p className="truncate text-sm font-medium">{name ?? "User"}</p>
          <p className="truncate text-xs text-muted-foreground">{email}</p>
        </div>
      </NavLink>
      <Button variant="ghost" size="sm" className="mt-2 w-full justify-start text-muted-foreground" onClick={onLogout}>
        <LogOut className="h-4 w-4" /> Sign out
      </Button>
    </div>
  );
}
