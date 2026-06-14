// Central API client. Every network call goes through here so envelope parsing,
// auth headers, and 401-refresh live in exactly one place.
//
// Backend contract:
//   success -> { "data": ... }
//   error   -> { "error": { "code", "message" } }

const BASE = "/api";

/** ApiError carries the backend's machine code + human message + HTTP status. */
export class ApiError extends Error {
  status: number;
  code?: string;
  constructor(status: number, code?: string, message?: string) {
    super(message || code || `HTTP ${status}`);
    this.name = "ApiError";
    this.status = status;
    this.code = code;
  }
}

// Access token is held in memory only (never localStorage) to limit XSS exposure.
let accessToken: string | null = null;
export function setAccessToken(token: string | null): void {
  accessToken = token;
}
export function getAccessToken(): string | null {
  return accessToken;
}

// The auth layer installs this to refresh the access token on a 401.
type RefreshHandler = () => Promise<string | null>;
let refreshHandler: RefreshHandler | null = null;
export function setRefreshHandler(fn: RefreshHandler): void {
  refreshHandler = fn;
}

interface RequestOptions {
  body?: unknown;
  auth?: boolean;
  _retried?: boolean;
}

async function parse<T>(res: Response): Promise<T> {
  let body: { data?: T; error?: { code?: string; message?: string } } | null = null;
  if (res.status !== 204) {
    try {
      body = await res.json();
    } catch {
      body = null;
    }
  }
  if (!res.ok) {
    throw new ApiError(res.status, body?.error?.code, body?.error?.message);
  }
  return body?.data as T;
}

async function raw<T>(method: string, path: string, opts: RequestOptions = {}): Promise<T> {
  const { body, auth = true, _retried = false } = opts;
  const headers: Record<string, string> = { "Content-Type": "application/json" };
  if (auth && accessToken) headers.Authorization = `Bearer ${accessToken}`;

  const res = await fetch(BASE + path, {
    method,
    headers,
    body: body !== undefined ? JSON.stringify(body) : undefined,
  });

  // On 401, attempt one transparent refresh, then replay the request once.
  if (res.status === 401 && auth && !_retried && refreshHandler) {
    const fresh = await refreshHandler();
    if (fresh) return raw<T>(method, path, { body, auth, _retried: true });
  }
  return parse<T>(res);
}

export const api = {
  get: <T = unknown>(path: string, opts?: RequestOptions) => raw<T>("GET", path, opts),
  post: <T = unknown>(path: string, body?: unknown, opts?: RequestOptions) =>
    raw<T>("POST", path, { ...opts, body }),
  put: <T = unknown>(path: string, body?: unknown, opts?: RequestOptions) =>
    raw<T>("PUT", path, { ...opts, body }),
  del: <T = unknown>(path: string, opts?: RequestOptions) => raw<T>("DELETE", path, opts),
};
