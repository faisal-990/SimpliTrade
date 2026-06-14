import { Navigate, Outlet, useLocation } from "react-router-dom";
import { useAuth } from "@/auth/useAuth";
import { Loading } from "@/components/common/states";

// Guards the authenticated area: waits for the initial session restore, then
// either renders the nested routes or redirects to login (preserving where the
// user was headed).
export function RequireAuth() {
  const { isAuthenticated, ready } = useAuth();
  const location = useLocation();

  if (!ready) {
    return (
      <div className="grid min-h-svh place-items-center">
        <Loading label="Restoring your session…" />
      </div>
    );
  }
  if (!isAuthenticated) {
    return <Navigate to="/login" replace state={{ from: location.pathname }} />;
  }
  return <Outlet />;
}
