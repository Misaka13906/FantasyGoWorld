export type WSMessageType =
  | 'HEARTBEAT'
  | 'JOIN_ROOM'
  | 'LEAVE_ROOM'
  | 'INVITE'
  | 'INVITE_REPLY'
  | 'PROPOSAL'
  | 'MOVE'
  | 'PASS'
  | 'RESIGN'
  | 'LOBBY_UPDATE'; // Added for Phase 5

export interface WSEnvelope<T = any> {
  type: WSMessageType;
  seq: number;
  room_id: string; // "hall" or room UUID/ID string
  payload: T;
  timestamp?: number;
}
