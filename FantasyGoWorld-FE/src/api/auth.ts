import http from './http';
import type { User } from '../types/user';

export interface RegisterReq {
  username: string;
  password:  string;
  nickname:  string;
  rank:      string;
}

export interface LoginReq {
  username: string;
  password: string;
}

export interface AuthRes {
  code: number;
  msg:  string;
  data: {
    access_token: string;
    user: User;
  };
}

export const register = (data: RegisterReq) => http.post('/auth/register', data);

export const login = (data: LoginReq): Promise<AuthRes> => http.post('/auth/login', data);

export const refresh = (): Promise<{ data: { access_token: string } }> => http.post('/auth/refresh');

export const logout = () => http.post('/auth/logout');
