import { useState } from "react";
import { Link, useNavigate, useLocation } from "react-router-dom";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { z } from "zod";
import { Eye, EyeOff, Mail, Lock, TrendingUp, Users, Wallet, ArrowRight } from "lucide-react";
import { Brand } from "@/components/Brand";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Spinner } from "@/components/common/states";
import { GoogleButton } from "@/components/auth/GoogleButton";
import { useAuth } from "@/auth/useAuth";
import { ApiError } from "@/lib/api";

const FEATURES = [
  { icon: TrendingUp, title: "Real-time market data", body: "Live prices and charts — trade the real market, risk-free." },
  { icon: Users, title: "Follow 20 legendary investors", body: "Graham, Buffett, Wood & more — see their strategies play out." },
  { icon: Wallet, title: "$100,000 to practice", body: "Build a portfolio, learn the ropes, keep every lesson." },
];

const schema = z.object({
  email: z.string().email("Enter a valid email"),
  password: z.string().min(8, "At least 8 characters"),
});
type Values = z.infer<typeof schema>;

export default function Login() {
  const [showPw, setShowPw] = useState(false);
  const [formError, setFormError] = useState<string | null>(null);
  const { login } = useAuth();
  const navigate = useNavigate();
  const location = useLocation();
  const justReset = (location.state as { reset?: boolean } | null)?.reset === true;

  const {
    register,
    handleSubmit,
    formState: { errors, isSubmitting },
  } = useForm<Values>({ resolver: zodResolver(schema) });

  const onSubmit = async (values: Values) => {
    setFormError(null);
    try {
      await login(values.email, values.password);
      navigate("/app/dashboard", { replace: true });
    } catch (e) {
      setFormError(e instanceof ApiError ? e.message : "Could not sign in. Try again.");
    }
  };

  return (
    <div className="grid min-h-svh lg:grid-cols-[1.05fr_1fr]">
      <aside className="relative hidden overflow-hidden lg:flex">
        <div className="absolute inset-0 bg-[radial-gradient(120%_120%_at_15%_0%,oklch(0.7_0.15_60)_0%,oklch(0.55_0.15_38)_45%,oklch(0.36_0.09_40)_100%)]" />
        <div className="absolute inset-0 bg-warm-grain opacity-60" />
        <div className="absolute -left-24 top-1/3 h-72 w-72 rounded-full bg-[oklch(0.85_0.12_75)] opacity-25 blur-3xl" />
        <div className="relative z-10 flex w-full flex-col justify-between p-10 text-[oklch(0.99_0.01_78)] xl:p-14">
          <Brand text="text-xl text-[oklch(0.99_0.01_78)]" mark="h-9 w-9 bg-white/15 backdrop-blur" />
          <div className="max-w-md">
            <h1 className="text-4xl font-semibold leading-tight tracking-tight xl:text-[2.6rem]">
              Invest like the legends.
              <br />
              <span className="text-[oklch(0.92_0.09_80)]">Risk-free.</span>
            </h1>
            <p className="mt-4 text-base leading-relaxed text-white/75">
              A calm, real-time trading simulator. Practice with real market data, follow famous investors, and learn
              what actually works — without risking a cent.
            </p>
            <ul className="mt-10 space-y-5">
              {FEATURES.map((f) => (
                <li key={f.title} className="flex gap-3.5">
                  <span className="mt-0.5 flex h-9 w-9 shrink-0 items-center justify-center rounded-lg bg-white/12 ring-1 ring-white/15 backdrop-blur">
                    <f.icon className="h-[18px] w-[18px]" />
                  </span>
                  <div>
                    <p className="font-medium leading-tight">{f.title}</p>
                    <p className="text-sm text-white/65">{f.body}</p>
                  </div>
                </li>
              ))}
            </ul>
          </div>
          <p className="text-sm text-white/55">
            “Price is what you pay; value is what you get.”<span className="text-white/40"> — Benjamin Graham</span>
          </p>
        </div>
      </aside>

      <main className="flex items-center justify-center px-6 py-10 sm:px-10">
        <div className="w-full max-w-sm">
          <div className="mb-10 lg:hidden">
            <Brand />
          </div>
          <div className="mb-8">
            <h2 className="text-2xl font-semibold tracking-tight">Welcome back</h2>
            <p className="mt-1.5 text-sm text-muted-foreground">Sign in to your simulated portfolio.</p>
          </div>

          {justReset && !formError && (
            <div className="mb-5 rounded-lg border border-gain/30 bg-gain/10 px-3 py-2 text-sm text-gain">
              Password updated — sign in with your new password.
            </div>
          )}

          {formError && (
            <div className="mb-5 rounded-lg border border-loss/30 bg-loss/10 px-3 py-2 text-sm text-loss">
              {formError}
            </div>
          )}

          <GoogleButton />
          <div className="my-6 flex items-center gap-3 text-xs text-muted-foreground">
            <span className="h-px flex-1 bg-border" /> or with email <span className="h-px flex-1 bg-border" />
          </div>

          <form className="space-y-5" onSubmit={handleSubmit(onSubmit)} noValidate>
            <div className="space-y-2">
              <Label htmlFor="email">Email</Label>
              <div className="relative">
                <Mail className="pointer-events-none absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
                <Input id="email" type="email" placeholder="you@example.com" autoComplete="email" className="h-11 pl-9" {...register("email")} />
              </div>
              {errors.email && <p className="text-xs text-loss">{errors.email.message}</p>}
            </div>

            <div className="space-y-2">
              <div className="flex items-center justify-between">
                <Label htmlFor="password">Password</Label>
                <Link to="/forgot-password" className="text-xs font-medium text-primary hover:underline">Forgot password?</Link>
              </div>
              <div className="relative">
                <Lock className="pointer-events-none absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
                <Input id="password" type={showPw ? "text" : "password"} placeholder="••••••••" autoComplete="current-password" className="h-11 pl-9 pr-10" {...register("password")} />
                <button type="button" onClick={() => setShowPw((v) => !v)} className="absolute right-3 top-1/2 -translate-y-1/2 text-muted-foreground transition-colors hover:text-foreground" aria-label={showPw ? "Hide password" : "Show password"}>
                  {showPw ? <EyeOff className="h-4 w-4" /> : <Eye className="h-4 w-4" />}
                </button>
              </div>
              {errors.password && <p className="text-xs text-loss">{errors.password.message}</p>}
            </div>

            <Button type="submit" className="h-11 w-full text-[0.95rem]" disabled={isSubmitting}>
              {isSubmitting ? <Spinner /> : <>Sign in <ArrowRight className="h-4 w-4" /></>}
            </Button>
          </form>

          <div className="my-7 flex items-center gap-3 text-xs text-muted-foreground">
            <span className="h-px flex-1 bg-border" /> new to SimpliTrade? <span className="h-px flex-1 bg-border" />
          </div>
          <Button asChild variant="outline" className="h-11 w-full text-[0.95rem]">
            <Link to="/signup">Create a free account</Link>
          </Button>
          <p className="mt-8 text-center text-xs leading-relaxed text-muted-foreground">
            Simulated trading for education. No real money, ever.
          </p>
        </div>
      </main>
    </div>
  );
}
