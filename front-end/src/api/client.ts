const BASE = "/api";

// Access token is kept in memory (not localStorage) to reduce XSS risk.
// The refresh token lives in an HttpOnly cookie managed by the server.
let accessToken: string | null = null;

export function setAccessToken(token: string | null) {
  accessToken = token;
}

export function getAccessToken() {
  return accessToken;
}

async function request<T>(
  path: string,
  options: RequestInit = {},
): Promise<T> {
  const headers: Record<string, string> = {
    "Content-Type": "application/json",
    ...(options.headers as Record<string, string>),
  };

  if (accessToken) {
    headers["Authorization"] = `Bearer ${accessToken}`;
  }

  const res = await fetch(`${BASE}${path}`, {
    ...options,
    credentials: "include", // send HttpOnly refresh_token cookie
    headers,
  });

  const body = await res.json();

  if (!res.ok) {
    throw new Error(body.error ?? "Request failed");
  }

  return body;
}

export const api = {
  get: <T>(path: string) => request<T>(path),
  post: <T>(path: string, data?: unknown) =>
    request<T>(path, {
      method: "POST",
      body: data ? JSON.stringify(data) : undefined,
    }),
};
