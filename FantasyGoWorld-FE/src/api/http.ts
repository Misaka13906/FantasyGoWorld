import axios, { type InternalAxiosRequestConfig, type AxiosResponse, type AxiosError } from 'axios';

const http = axios.create({
  baseURL: import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080/api/v1',
  timeout: 10000,
  withCredentials: true, // 允许携带 Cookie
});

// 请求拦截器
http.interceptors.request.use(
  (config: InternalAxiosRequestConfig) => {
    const token = localStorage.getItem('access_token');
    if (token && config.headers) {
      config.headers.Authorization = `Bearer ${token}`;
    }
    return config;
  },
  (error: any) => {
    return Promise.reject(error);
  }
);

// 响应拦截器
http.interceptors.response.use(
  (response: AxiosResponse) => {
    // 只要是 2xx 都在此处理（Success）
    if (response.status >= 200 && response.status < 300) {
      return response.data;
    }
    // 实际上 axios 默认会将 2xx 以外的抛入下一个 error 闭包，
    // 这里做显式拦截以防万一。
    return Promise.reject(new Error(`Unexpected status ${response.status}`));
  },
  async (error: AxiosError) => {
    const originalRequest = error.config as InternalAxiosRequestConfig & { _retry?: boolean };
    
    // 如果是 401 且未重试过，尝试刷新 Token
    if (error.response?.status === 401 && !originalRequest._retry) {
      originalRequest._retry = true;
      try {
        // TODO: 调用刷新 Token 接口 (Phase 2 实现)
        // const { access_token } = await refreshAuthToken();
        // localStorage.setItem('access_token', access_token);
        // return http(originalRequest);
        console.warn('Unauthorized. Token refresh not yet implemented in Phase 1.');
      } catch (refreshError) {
        // 刷新失败，重定向到登录页
        // window.location.href = '/login';
        return Promise.reject(refreshError);
      }
    }
    return Promise.reject(error);
  }
);

export default http;
