import { useEffect, useRef, useState } from "react";
import { useNavigate } from "react-router-dom";
import { Spinner } from "@/components/common/states";
import { Brand } from "@/components/Brand";
import { useAuth } from "@/auth/useAuth";

// OAuthCallback is where the backend redirects after a provider sign-in. The
// refresh token arrives in the URL fragment (kept out of logs/history); we
// exchange it for a session, then land in the app.
export default function OAuthCallback() {
  const { completeOAuth } = useAuth();
  const navigate = useNavigate();
  const [error, setError] = useState(false);
  const ran = useRef(false);

  useEffect(() => {
    if (ran.current) return; // StrictMode double-invoke guard
    ran.current = true;

    const params = new URLSearchParams(window.location.hash.slice(1));
    const rt = params.get("refresh_token");
    // Scrub the token from the address bar immediately.
    window.history.replaceState(null, "", window.location.pathname);

    if (!rt) {
      setError(true);
      return;
    }
    completeOAuth(rt)
      .then(() => navigate("/app/dashboard", { replace: true }))
      .catch(() => setError(true));
  }, [completeOAuth, navigate]);

  return (
    <div className="flex min-h-svh flex-col items-center justify-center gap-4 bg-background px-6 text-center">
      <Brand />
      {error ? (
        <>
          <p className="text-sm text-loss">Sign-in didn’t complete. Please try again.</p>
          <button onClick={() => navigate("/login", { replace: true })} className="text-sm font-medium text-primary hover:underline">
            Back to sign in
          </button>
        </>
      ) : (
        <p className="flex items-center gap-2 text-sm text-muted-foreground">
          <Spinner /> Finishing sign-in…
        </p>
      )}
    </div>
  );
}
