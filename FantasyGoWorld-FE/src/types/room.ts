import type { User } from './user';

export interface Room {
    id: number;
    owner_id: number;
    description: string;
    is_public: boolean;
    status: number;
    owner?: User;
}
