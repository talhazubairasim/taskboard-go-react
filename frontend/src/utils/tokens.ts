import type { User } from '../types';

const ACCESS = 'accessToken';
const REFRESH = 'refreshToken';
const USER = 'user';

export const tokens = {
  access: () => localStorage.getItem(ACCESS),
  refresh: () => localStorage.getItem(REFRESH),
  user: (): User | null => {
    const raw = localStorage.getItem(USER);
    if (!raw) return null;
    try {
      return JSON.parse(raw) as User;
    } catch {
      return null;
    }
  },
  save: (access: string, refresh: string, user?: User) => {
    localStorage.setItem(ACCESS, access);
    localStorage.setItem(REFRESH, refresh);
    if (user) localStorage.setItem(USER, JSON.stringify(user));
  },
  saveUser: (user: User) => localStorage.setItem(USER, JSON.stringify(user)),
  clear: () => {
    localStorage.removeItem(ACCESS);
    localStorage.removeItem(REFRESH);
    localStorage.removeItem(USER);
  },
};
