import { useState } from "react";
import { Link, useNavigate } from "react-router-dom";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { z } from "zod";
import { Eye, EyeOff, ArrowRight } from "lucide-react";
import { Brand } from "@/components/Brand";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Spinner } from "@/components/common/states";
import { GoogleButton } from "@/components/auth/GoogleButton";
import { useAuth } from "@/auth/useAuth";
import { ApiError } from "@/lib/api";

const schema = z.object({
  name: z.string().min(2, "Tell us your name").max(100),
  email: z.string().email("Enter a valid email"),
  password: z.string().min(8, "At least 8 characters").max(72),
});
type Values = z.infer<typeof schema>;

export default function Signup() {
  const [showPw, setShowPw] = useState(false);
  const [formError, setFormError] = useState<string | null>(null);
  const { signup } = useAuth();
  const navigate = useNavigate();
  const {
    register,
    handleSubmit,
    formState: { errors, isSubmitting },
  } = useForm<Values>({ resolver: zodResolver(schema) });

  const onSubmit = async (values: Values) => {
    setFormError(null);
    try {
      await signup(values.name, values.email, values.password);
      navigate("/app/dashboard", { replace: true });
    } catch (e) {
      setFormError(e instanceof ApiError ? e.message : "Could not create your account.");
    }
  };

  return (
    <div className="relative grid min-h-svh place-items-center overflow-hidden px-6 py-10">
      {/* warm ambient background */}
      <div className="absolute inset-0 -z-10 bg-[radial-gradient(90%_70%_at_50%_-10%,oklch(0.93_0.05_70)_0%,var(--background)_60%)]" />

      <div className="w-full max-w-sm">
        <div className="mb-8 flex flex-col items-center text-center">
          <Brand mark="h-10 w-10" text="text-xl" />
          <h2 className="mt-6 text-2xl font-semibold tracking-tight">Create your account</h2>
          <p className="mt-1.5 text-sm text-muted-foreground">Start with $100,000 in simulated cash.</p>
        </div>

        <div className="rounded-2xl border bg-card p-6 shadow-sm">
          {formError && (
            <div className="mb-5 rounded-lg border border-loss/30 bg-loss/10 px-3 py-2 text-sm text-loss">{formError}</div>
          )}
          <GoogleButton label="Sign up with Google" />
          <div className="my-6 flex items-center gap-3 text-xs text-muted-foreground">
            <span className="h-px flex-1 bg-border" /> or with email <span className="h-px flex-1 bg-border" />
          </div>
          <form className="space-y-4" onSubmit={handleSubmit(onSubmit)} noValidate>
            <div className="space-y-2">
              <Label htmlFor="name">Name</Label>
              <Input id="name" className="h-11" placeholder="Ada Lovelace" autoComplete="name" {...register("name")} />
              {errors.name && <p className="text-xs text-loss">{errors.name.message}</p>}
            </div>
            <div className="space-y-2">
              <Label htmlFor="email">Email</Label>
              <Input id="email" type="email" className="h-11" placeholder="you@example.com" autoComplete="email" {...register("email")} />
              {errors.email && <p className="text-xs text-loss">{errors.email.message}</p>}
            </div>
            <div className="space-y-2">
              <Label htmlFor="password">Password</Label>
              <div className="relative">
                <Input id="password" type={showPw ? "text" : "password"} className="h-11 pr-10" placeholder="At least 8 characters" autoComplete="new-password" {...register("password")} />
                <button type="button" onClick={() => setShowPw((v) => !v)} className="absolute right-3 top-1/2 -translate-y-1/2 text-muted-foreground hover:text-foreground" aria-label={showPw ? "Hide password" : "Show password"}>
                  {showPw ? <EyeOff className="h-4 w-4" /> : <Eye className="h-4 w-4" />}
                </button>
              </div>
              {errors.password && <p className="text-xs text-loss">{errors.password.message}</p>}
            </div>
            <Button type="submit" className="h-11 w-full" disabled={isSubmitting}>
              {isSubmitting ? <Spinner /> : <>Create account <ArrowRight className="h-4 w-4" /></>}
            </Button>
          </form>
        </div>

        <p className="mt-6 text-center text-sm text-muted-foreground">
          Already have an account?{" "}
          <Link to="/login" className="font-medium text-primary hover:underline">Sign in</Link>
        </p>
      </div>
    </div>
  );
}
