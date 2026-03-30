import { create } from 'zustand';
import type { User } from '../types/user';
import type { Room } from '../types/room';
import { getOnlineUsers } from '../api/user';
import { getPublicRooms, createRoom, type CreateRoomReq, closeRoom } from '../api/room';

interface LobbyState {
    onlineUsers: User[];
    publicRooms: Room[];
    totalUsers: number;
    totalRooms: number;
    loading: boolean;
    error: string | null;

    fetchLobbyData: () => Promise<void>;
    createNewRoom: (req: CreateRoomReq) => Promise<void>;
    closeExistingRoom: (roomId: number) => Promise<void>;
}

export const useLobbyStore = create<LobbyState>((set) => ({
    onlineUsers: [],
    publicRooms: [],
    totalUsers: 0,
    totalRooms: 0,
    loading: false,
    error: null,

    fetchLobbyData: async () => {
        set({ loading: true, error: null });
        try {
            // Initiate parallel fetching
            const [usersRes, roomsRes] = await Promise.all([
                getOnlineUsers(1, 50),
                getPublicRooms(1, 50)
            ]);

            set({
                onlineUsers: usersRes.data.items,
                totalUsers: usersRes.data.total,
                publicRooms: roomsRes.data.items,
                totalRooms: roomsRes.data.total,
                loading: false
            });
        } catch (err: any) {
            set({ error: err.message || '获取大厅数据失败', loading: false });
        }
    },

    createNewRoom: async (req: CreateRoomReq) => {
        set({ loading: true, error: null });
        try {
            await createRoom(req);
            // Re-fetch rooms to update the list
            const roomsRes = await getPublicRooms(1, 50);
            set({ 
                publicRooms: roomsRes.data.items, 
                totalRooms: roomsRes.data.total,
                loading: false 
            });
        } catch (err: any) {
            set({ error: err.message || '创建房间失败', loading: false });
            throw err;
        }
    },

    closeExistingRoom: async (roomId: number) => {
        set({ loading: true, error: null });
        try {
            await closeRoom(roomId);
            // Re-fetch
            const roomsRes = await getPublicRooms(1, 50);
            set({ 
                publicRooms: roomsRes.data.items, 
                totalRooms: roomsRes.data.total,
                loading: false 
            });
        } catch (err: any) {
            set({ error: err.message || '删除房间失败', loading: false });
            throw err;
        }
    }
}));
