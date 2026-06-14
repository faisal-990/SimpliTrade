import {
  createContext,
  useCallback,
  useEffect,
  useMemo,
  useRef,
  useState,
  type ReactNode,
} from "react";
import { api, setAccessToken, setRefreshHandler } from "@/lib/api";
import type { AuthResponse, User } from "@/types/api";

// Security model:
//  - access token: in memory only (api.ts) — short-lived, not in storage.
//  - refresh token: persisted so sessions survive reloads; exchanged for a fresh
//    access token on boot.
//  Hardening path: move refresh tokens to httpOnly cookies server-side.
const REFRESH_KEY = "simplitrade.refresh";

interface AuthValue {
  user: User | null;
  ready: boolean;
  isAuthenticated: boolean;
  login: (email: string, password: string) => Promise<AuthResponse>;
  signup: (name: string, email: string, password: string) => Promise<AuthResponse>;
  logout: () => Promise<void>;
  // completeOAuth ingests a refresh token from an OAuth callback and exchanges it
  // for a live session (access token + profile).
  completeOAuth: (refreshToken: string) => Promise<void>;
  // setUser replaces the cached profile (e.g. after editing the About-me section).
  setUser: (user: User) => void;
}

export const AuthContext = createContext<AuthValue | null>(null);

function loadRefresh(): string | null {
  try {
    return localStorage.getItem(REFRESH_KEY);
  } catch {
    return null;
  }
}
function storeRefresh(token: string | null): void {
  try {
    if (token) localStorage.setItem(REFRESH_KEY, token);
    else localStorage.removeItem(REFRESH_KEY);
  } catch {
    /* storage unavailable — session just won't persist */
  }
}

export function AuthProvider({ children }: { children: ReactNode }) {
  const [user, setUser] = useState<User | null>(null);
  const [ready, setReady] = useState(false);
  const refreshToken = useRef<string | null>(loadRefresh());
  // Single-flight guard: refresh tokens are single-use/rotating, so concurrent
  // refreshes (boot restore + OAuth callback + StrictMode) must share ONE call —
  // otherwise the racing rotations invalidate each other and force a logout.
  const refreshing = useRef<Promise<string | null> | null>(null);

  const applySession = useCallback((resp: AuthResponse): AuthResponse => {
    setAccessToken(resp.access_token);
    refreshToken.current = resp.refresh_token;
    storeRefresh(resp.refresh_token);
    setUser(resp.user);
    return resp;
  }, []);

  const clearSession = useCallback(() => {
    setAccessToken(null);
    refreshToken.current = null;
    storeRefresh(null);
    setUser(null);
  }, []);

  // Exchange the refresh token for a fresh access token; returns it (for api.ts's
  // 401 retry) or null on failure.
  const refresh = useCallback((): Promise<string | null> => {
    if (refreshing.current) return refreshing.current; // join the in-flight rotation
    const p = (async (): Promise<string | null> => {
      const rt = refreshToken.current;
      if (!rt) return null;
      try {
        const resp = await api.post<AuthResponse>("/auth/refresh", { refresh_token: rt }, { auth: false });
        applySession(resp);
        return resp.access_token;
      } catch {
        clearSession();
        return null;
      }
    })();
    refreshing.current = p;
    void p.finally(() => {
      refreshing.current = null;
    });
    return p;
  }, [applySession, clearSession]);

  useEffect(() => {
    setRefreshHandler(refresh);
  }, [refresh]);

  // Restore the session on first load.
  useEffect(() => {
    let active = true;
    (async () => {
      if (refreshToken.current) {
        const token = await refresh();
        if (token && active) {
          try {
            const me = await api.get<User>("/auth/me");
            if (active) setUser(me);
          } catch {
            /* keep user from the refresh response */
          }
        }
      }
      if (active) setReady(true);
    })();
    return () => {
      active = false;
    };
  }, [refresh]);

  const login = useCallback(
    (email: string, password: string) =>
      api.post<AuthResponse>("/auth/login", { email, password }, { auth: false }).then(applySession),
    [applySession]
  );

  const signup = useCallback(
    (name: string, email: string, password: string) =>
      api.post<AuthResponse>("/auth/signup", { name, email, password }, { auth: false }).then(applySession),
    [applySession]
  );

  const logout = useCallback(async () => {
    const rt = refreshToken.current;
    clearSession();
    if (rt) {
      try {
        await api.post("/auth/logout", { refresh_token: rt }, { auth: false });
      } catch {
        /* best-effort */
      }
    }
  }, [clearSession]);

  const completeOAuth = useCallback(
    async (rt: string) => {
      refreshToken.current = rt;
      storeRefresh(rt);
      const token = await refresh(); // exchanges the token, applies the session
      if (!token) throw new Error("Could not complete sign-in");
    },
    [refresh]
  );

  const value = useMemo<AuthValue>(
    () => ({ user, ready, isAuthenticated: !!user, login, signup, logout, completeOAuth, setUser }),
    [user, ready, login, signup, logout, completeOAuth]
  );

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
}
