import React, { useEffect } from 'react';
import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom';
import { LoginPage } from './pages/LoginPage';
import { useAuthStore } from './store/authStore';
import axios from 'axios';

// ProtectedRoute: 拦截未登录请求
const ProtectedRoute = ({ children }: { children: React.ReactNode }) => {
  const token = useAuthStore((s) => s.token);
  if (!token) {
    return <Navigate to="/login" replace />;
  }
  return <>{children}</>;
};

// PublicRoute: 防止已登录用户重复登录
const PublicRoute = ({ children }: { children: React.ReactNode }) => {
  const token = useAuthStore((s) => s.token);
  if (token) {
    return <Navigate to="/lobby" replace />;
  }
  return <>{children}</>;
};

import { logout } from './api/auth';

const LobbyPlaceholder = () => {
  const user = useAuthStore((s) => s.currentUser);
  const clearAuth = useAuthStore((s) => s.clearAuth);

  const handleLogout = async () => {
    try {
      await logout();
    } catch(e) {
      console.error(e);
    } finally {
      clearAuth();
    }
  };

  return (
    <div style={{ padding: '40px', color: 'white', background: '#1a1a2e', minHeight: '100vh' }}>
      <h1>大厅 (施工中...)</h1>
      {user && (
        <div>
          <p>欢迎, {user.nickname} ({user.rank})</p>
          <button onClick={handleLogout} style={{ padding: '8px 16px', background: '#ff4b2b', border: 'none', color: 'white', borderRadius: '4px', cursor: 'pointer' }}>
            退出登录
          </button>
        </div>
      )}
    </div>
  );
};

export const App: React.FC = () => {
  const setToken = useAuthStore((s) => s.setToken);

  // App Mount: 尝试静默续期
  useEffect(() => {
    const silentRefresh = async () => {
      try {
        const baseURL = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080/api/v1';
        const res = await axios.post(`${baseURL}/auth/refresh`, {}, { withCredentials: true });
        if (res.data.code === 20000) {
          setToken(res.data.data.access_token);
        }
      } catch (err) {
        // 静默续期失败很正常（未登录或 Cookie 过期）
        console.log('Silent refresh skipped');
      }
    };
    silentRefresh();
  }, [setToken]);

  return (
    <BrowserRouter>
      <Routes>
        <Route
          path="/login"
          element={
            <PublicRoute>
              <LoginPage />
            </PublicRoute>
          }
        />
        <Route
          path="/lobby"
          element={
            <ProtectedRoute>
              <LobbyPlaceholder />
            </ProtectedRoute>
          }
        />
        <Route path="/" element={<Navigate to="/lobby" replace />} />
      </Routes>
    </BrowserRouter>
  );
};
