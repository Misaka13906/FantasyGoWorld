import { describe, it, expect, beforeEach } from 'vitest';
import { useAuthStore } from '../../src/store/authStore';

describe('AuthStore', () => {
  beforeEach(() => {
    useAuthStore.getState().clearAuth();
  });

  it('initially has no token and user', () => {
    const state = useAuthStore.getState();
    expect(state.token).toBeNull();
    expect(state.currentUser).toBeNull();
    expect(state.status).toBe('idle');
  });

  it('can set token and user', () => {
    const mockUser = {
      id: 1,
      username: 'test',
      nickname: 'tester',
      elo: 1500,
      rank: '18K',
      dnd: false,
    };

    useAuthStore.getState().setAuth('mock_token', mockUser);

    const state = useAuthStore.getState();
    expect(state.token).toBe('mock_token');
    expect(state.currentUser).toEqual(mockUser);
  });

  it('can set specific status', () => {
    useAuthStore.getState().setStatus('gaming');
    expect(useAuthStore.getState().status).toBe('gaming');
  });

  it('can clear auth properly', () => {
    useAuthStore.getState().setAuth('mock_token', { id: 1 } as any);
    useAuthStore.getState().setStatus('gaming');
    
    useAuthStore.getState().clearAuth();
    
    const state = useAuthStore.getState();
    expect(state.token).toBeNull();
    expect(state.currentUser).toBeNull();
    expect(state.status).toBe('idle');
  });
});
