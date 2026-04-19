import { login, logout, refresh, register } from "@/api/auth";
import { getAccessToken, setAccessToken } from "@/api/client";
import type { User } from "@/types";
import {
  createContext,
  useCallback,
  useContext,
  useEffect,
  useState,
  type ReactNode,
} from "react";

interface AuthContextValue {
  user: User | null;
  isLoading: boolean;
  register: (username: string, email: string, password: string) => Promise<void>;
  login: (email: string, password: string) => Promise<void>;
  logout: () => Promise<void>;
}

const AuthContext = createContext<AuthContextValue | null>(null);

export function AuthProvider({ children }: { children: ReactNode }) {
  const [user, setUser] = useState<User | null>(null);
  const [isLoading, setIsLoading] = useState(true);

  // On mount, try to restore the session via the refresh token cookie.
  useEffect(() => {
    refresh()
      .then((res) => {
        if (res.data) setAccessToken(res.data);
        // access token restored but we don't have user info here —
        // fetch it from a /me endpoint if you add one, or decode the JWT.
        // For now we rely on login/register to populate user.
      })
      .catch(() => {
        /* no valid session — stay logged out */
      })
      .finally(() => setIsLoading(false));
  }, []);

  const handleRegister = useCallback(
    async (username: string, email: string, password: string) => {
      const res = await register(username, email, password);
      if (res.data) {
        setAccessToken(res.data.access_token);
        setUser(res.data.user);
      }
    },
    [],
  );

  const handleLogin = useCallback(async (email: string, password: string) => {
    const res = await login(email, password);
    if (res.data) {
      setAccessToken(res.data.access_token);
      setUser(res.data.user);
    }
  }, []);

  const handleLogout = useCallback(async () => {
    try {
      await logout();
    } finally {
      setAccessToken(null);
      setUser(null);
    }
  }, []);

  return (
    <AuthContext.Provider
      value={{
        user,
        isLoading,
        register: handleRegister,
        login: handleLogin,
        logout: handleLogout,
      }}
    >
      {children}
    </AuthContext.Provider>
  );
}

export function useAuth() {
  const ctx = useContext(AuthContext);
  if (!ctx) throw new Error("useAuth must be used inside <AuthProvider>");
  return ctx;
}

// Convenience: is there a token in memory right now?
export function useAccessToken() {
  return getAccessToken();
}
