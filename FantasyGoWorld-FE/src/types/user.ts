export interface User {
    id: number;
    username: string;
    nickname: string;
    rank: string;
    elo: number;
    avatar_url?: string;
    dnd?: boolean;
}
