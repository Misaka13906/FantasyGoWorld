import { create } from 'zustand';
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

export const useAuthStore = create<AuthState>((set) => ({
  token:       null,
  currentUser: null,
  status:      'idle',
  setAuth: (token, user) => set({ token, currentUser: user }),
  setToken: (token) => set({ token }),
  setStatus: (status) => set({ status }),
  clearAuth: () => set({ token: null, currentUser: null, status: 'idle' }),
}));
