import { api } from "./client";
import type { ApiResponse, Room } from "@/types";

export function getRooms() {
  return api.get<ApiResponse<Room[]>>("/rooms");
}

export function createRoom(name: string) {
  return api.post<ApiResponse<Room>>("/rooms", { name });
}
