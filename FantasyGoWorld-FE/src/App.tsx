import React, { useEffect } from 'react';
import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom';
import { LoginPage } from './pages/LoginPage';
import { LobbyPage } from './pages/LobbyPage';
import { useAuthStore } from './store/authStore';
import axios from 'axios';
import { wsClient } from './ws/wsClient';

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

// Removed import { logout } from './api/auth'; as LobbyPlaceholder is removed
// Removed LobbyPlaceholder component

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

  // WS lifecycle
  const token = useAuthStore((s) => s.token);
  useEffect(() => {
    if (token) {
      wsClient.connect();
    } else {
      wsClient.disconnect();
    }
    return () => wsClient.disconnect();
  }, [token]);

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
              <LobbyPage />
            </ProtectedRoute>
          }
        />
        <Route path="/" element={<Navigate to="/lobby" replace />} />
      </Routes>
    </BrowserRouter>
  );
};
