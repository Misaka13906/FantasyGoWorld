import http from './http';

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
    user: {
      id:       number;
      username: string;
      nickname: string;
      elo:      number;
      rank:     string;
    };
  };
}

export const register = (data: RegisterReq) => http.post('/auth/register', data);

export const login = (data: LoginReq): Promise<AuthRes> => http.post('/auth/login', data);

export const refresh = (): Promise<{ data: { access_token: string } }> => http.post('/auth/refresh');

export const logout = () => http.post('/auth/logout');
