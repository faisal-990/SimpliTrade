import { useState } from "react";
import { Link, useNavigate } from "react-router-dom";
import { ArrowLeft, ArrowRight, Mail, Lock, KeyRound, Eye, EyeOff, MailCheck } from "lucide-react";
import { Brand } from "@/components/Brand";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Spinner } from "@/components/common/states";
import { useForgotPassword, useResetPassword } from "@/hooks/queries";
import { ApiError } from "@/lib/api";

type Step = "email" | "code";

export default function ForgotPassword() {
  const [step, setStep] = useState<Step>("email");
  const [email, setEmail] = useState("");
  const [code, setCode] = useState("");
  const [password, setPassword] = useState("");
  const [showPw, setShowPw] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const forgot = useForgotPassword();
  const reset = useResetPassword();
  const navigate = useNavigate();

  const requestCode = (e: React.FormEvent) => {
    e.preventDefault();
    setError(null);
    forgot.mutate(email.trim(), {
      onSuccess: () => setStep("code"),
      onError: (err) => setError(err instanceof ApiError ? err.message : "Something went wrong. Try again."),
    });
  };

  const submitReset = (e: React.FormEvent) => {
    e.preventDefault();
    setError(null);
    if (code.trim().length !== 6) {
      setError("Enter the 6-digit code from your email.");
      return;
    }
    if (password.length < 8) {
      setError("Password must be at least 8 characters.");
      return;
    }
    reset.mutate(
      { email: email.trim(), code: code.trim(), password },
      {
        onSuccess: () => navigate("/login", { replace: true, state: { reset: true } }),
        onError: (err) => setError(err instanceof ApiError ? err.message : "Could not reset password."),
      }
    );
  };

  return (
    <div className="flex min-h-svh items-center justify-center bg-background px-6 py-10">
      <div className="w-full max-w-sm">
        <div className="mb-8">
          <Brand />
        </div>

        {step === "email" ? (
          <>
            <div className="mb-8">
              <h2 className="text-2xl font-semibold tracking-tight">Forgot your password?</h2>
              <p className="mt-1.5 text-sm text-muted-foreground">
                Enter your email and we’ll send you a 6-digit code to reset it.
              </p>
            </div>

            {error && (
              <div className="mb-5 rounded-lg border border-loss/30 bg-loss/10 px-3 py-2 text-sm text-loss">{error}</div>
            )}

            <form className="space-y-5" onSubmit={requestCode} noValidate>
              <div className="space-y-2">
                <Label htmlFor="email">Email</Label>
                <div className="relative">
                  <Mail className="pointer-events-none absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
                  <Input
                    id="email"
                    type="email"
                    placeholder="you@example.com"
                    autoComplete="email"
                    required
                    className="h-11 pl-9"
                    value={email}
                    onChange={(e: React.ChangeEvent<HTMLInputElement>) => setEmail(e.target.value)}
                  />
                </div>
              </div>
              <Button type="submit" className="h-11 w-full text-[0.95rem]" disabled={forgot.isPending || !email}>
                {forgot.isPending ? <Spinner /> : <>Send reset code <ArrowRight className="h-4 w-4" /></>}
              </Button>
            </form>
          </>
        ) : (
          <>
            <div className="mb-6 flex items-center gap-2 rounded-lg bg-primary/8 px-3 py-2.5 text-sm text-muted-foreground">
              <MailCheck className="h-4 w-4 shrink-0 text-primary" />
              If an account exists for <span className="font-medium text-foreground">{email}</span>, a 6-digit code is on
              its way. Check your inbox (and spam).
            </div>
            <div className="mb-6">
              <h2 className="text-2xl font-semibold tracking-tight">Enter the code</h2>
              <p className="mt-1.5 text-sm text-muted-foreground">Then choose a new password. The code expires in 10 minutes.</p>
            </div>

            {error && (
              <div className="mb-5 rounded-lg border border-loss/30 bg-loss/10 px-3 py-2 text-sm text-loss">{error}</div>
            )}

            <form className="space-y-5" onSubmit={submitReset} noValidate>
              <div className="space-y-2">
                <Label htmlFor="code">6-digit code</Label>
                <div className="relative">
                  <KeyRound className="pointer-events-none absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
                  <Input
                    id="code"
                    inputMode="numeric"
                    maxLength={6}
                    placeholder="••••••"
                    required
                    className="h-11 pl-9 tracking-[0.4em]"
                    value={code}
                    onChange={(e: React.ChangeEvent<HTMLInputElement>) => setCode(e.target.value.replace(/\D/g, "").slice(0, 6))}
                  />
                </div>
              </div>

              <div className="space-y-2">
                <Label htmlFor="password">New password</Label>
                <div className="relative">
                  <Lock className="pointer-events-none absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
                  <Input
                    id="password"
                    type={showPw ? "text" : "password"}
                    placeholder="••••••••"
                    autoComplete="new-password"
                    required
                    className="h-11 pl-9 pr-10"
                    value={password}
                    onChange={(e: React.ChangeEvent<HTMLInputElement>) => setPassword(e.target.value)}
                  />
                  <button
                    type="button"
                    onClick={() => setShowPw((v) => !v)}
                    className="absolute right-3 top-1/2 -translate-y-1/2 text-muted-foreground transition-colors hover:text-foreground"
                    aria-label={showPw ? "Hide password" : "Show password"}
                  >
                    {showPw ? <EyeOff className="h-4 w-4" /> : <Eye className="h-4 w-4" />}
                  </button>
                </div>
              </div>

              <Button type="submit" className="h-11 w-full text-[0.95rem]" disabled={reset.isPending}>
                {reset.isPending ? <Spinner /> : "Reset password"}
              </Button>

              <button
                type="button"
                onClick={() => { setStep("email"); setError(null); setCode(""); }}
                className="text-xs font-medium text-muted-foreground hover:text-foreground"
              >
                Use a different email
              </button>
            </form>
          </>
        )}

        <Link
          to="/login"
          className="mt-8 flex items-center justify-center gap-1.5 text-sm text-muted-foreground hover:text-foreground"
        >
          <ArrowLeft className="h-4 w-4" /> Back to sign in
        </Link>
      </div>
    </div>
  );
}
