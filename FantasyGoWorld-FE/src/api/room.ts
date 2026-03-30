import http from './http';
import type { Room } from '../types/room';
import type { PaginatedResponse } from '../types/api';

export interface CreateRoomReq {
    description: string;
    is_public: boolean;
    password?: string;
}

export const createRoom = async (req: CreateRoomReq) => {
    return http.post<Room>('/room', req);
};

export const getPublicRooms = async (page = 1, pageSize = 20) => {
    return http.get<PaginatedResponse<Room>>('/room/list', {
        params: { page, page_size: pageSize }
    });
};

export const closeRoom = async (roomId: number) => {
    return http.delete(`/room/${roomId}`);
};
