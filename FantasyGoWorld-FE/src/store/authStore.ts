import { create } from 'zustand';
import { persist } from 'zustand/middleware';
import type { User } from '../types/user';

interface AuthState {
  token:       string | null;
  currentUser: User | null;
  status:      'idle' | 'gaming' | 'dnd';
  setAuth:     (token: string, user: User) => void;
  setToken:    (token: string) => void;
  setStatus:   (status: 'idle' | 'gaming' | 'dnd') => void;
  clearAuth:   () => void;
}

export const useAuthStore = create<AuthState>()(
  persist(
    (set) => ({
      token:       null,
      currentUser: null,
      status:      'idle',
      setAuth: (token, user) => set({ token, currentUser: user }),
      setToken: (token) => set({ token }),
      setStatus: (status) => set({ status }),
      clearAuth: () => set({ token: null, currentUser: null, status: 'idle' }),
    }),
    {
      name: 'fgw-auth-storage', // 存储在 localStorage 的 key
    }
  )
);
