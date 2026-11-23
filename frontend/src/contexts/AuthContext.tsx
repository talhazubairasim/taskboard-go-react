import { createContext, useCallback, useContext, useEffect, useMemo, useState, type ReactNode } from 'react';
import { useApolloClient } from '@apollo/client';
import type { User } from '../types';
import { tokens } from '../utils/tokens';
import { LOGOUT_EVENT, refreshAccessToken } from '../utils/apollo-client';

interface AuthContextType {
  user: User | null;
  isAuthenticated: boolean;
  login: (token: string, refreshToken: string, user: User) => void;
  logout: () => void;
  updateUser: (user: User) => void;
}

const AuthContext = createContext<AuthContextType | undefined>(undefined);

/** Access tokens live 15 minutes; refresh a little before that so WebSocket reconnects stay authorised. */
export const REFRESH_INTERVAL_MS = 14 * 60 * 1000;

export const AuthProvider = ({ children }: { children: ReactNode }) => {
  const client = useApolloClient();
  const [user, setUser] = useState<User | null>(() => (tokens.access() ? tokens.user() : null));

  const logout = useCallback(() => {
    tokens.clear();
    setUser(null);
    void client.clearStore();
  }, [client]);

  const login = useCallback((token: string, refreshToken: string, u: User) => {
    tokens.save(token, refreshToken, u);
    setUser(u);
  }, []);

  const updateUser = useCallback((u: User) => {
    tokens.saveUser(u);
    setUser(u);
  }, []);

  // The Apollo client fires LOGOUT_EVENT when a refresh fails.
  useEffect(() => {
    const onLogout = () => setUser(null);
    window.addEventListener(LOGOUT_EVENT, onLogout);
    return () => window.removeEventListener(LOGOUT_EVENT, onLogout);
  }, []);

  useEffect(() => {
    if (!user) return;
    const id = setInterval(() => void refreshAccessToken(), REFRESH_INTERVAL_MS);
    return () => clearInterval(id);
  }, [user]);

  const value = useMemo(
    () => ({ user, isAuthenticated: !!user, login, logout, updateUser }),
    [user, login, logout, updateUser],
  );
  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
};

// eslint-disable-next-line react-refresh/only-export-components
export const useAuth = () => {
  const ctx = useContext(AuthContext);
  if (!ctx) throw new Error('useAuth must be used within an AuthProvider');
  return ctx;
};
