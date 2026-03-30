import http from './http';
import type { User } from '../types/user';
import type { PaginatedResponse } from '../types/api';

export const getOnlineUsers = async (page = 1, pageSize = 20) => {
    return http.get<PaginatedResponse<User>>('/user/list', {
        params: { page, page_size: pageSize }
    });
};
