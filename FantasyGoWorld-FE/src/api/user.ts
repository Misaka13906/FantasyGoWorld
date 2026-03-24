import http from './http';

export interface User {
    id: number;
    username: string;
    nickname: string;
    rank: string;
    elo: number;
    avatar_url: string;
}

export interface PaginatedResponse<T> {
    items: T[];
    total: number;
    page: number;
    page_size: number;
}

export const getOnlineUsers = async (page = 1, pageSize = 20) => {
    return http.get<PaginatedResponse<User>>('/user/list', {
        params: { page, page_size: pageSize }
    });
};
