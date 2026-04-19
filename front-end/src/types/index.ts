export interface User {
  id: string;
  username: string;
  email: string;
  created_at: string;
}

export interface Room {
  id: string;
  name: string;
  created_by: string;
  created_at: string;
}

export interface ChatMessage {
  id: string;
  room_id: string;
  user_id: string;
  username: string;
  content: string;
  timestamp: number;
}

export interface OnlineUser {
  user_id: string;
  username: string;
}

// ─── WebSocket event types ────────────────────────────────────────────────────

export type EventType =
  | "message"
  | "typing"
  | "join"
  | "leave"
  | "online_users"
  | "history";

export interface IncomingEvent {
  type: EventType;
  content: string;
}

export interface OutgoingEvent {
  type: EventType;
  content?: ChatMessage | ChatMessage[] | OnlineUser[];
  user_id?: string;
  username?: string;
  timestamp?: number;
}

// ─── API response wrapper ─────────────────────────────────────────────────────

export interface ApiResponse<T> {
  success: boolean;
  message?: string;
  data?: T;
  error?: string;
}

export interface AuthData {
  user: User;
  access_token: string;
}
