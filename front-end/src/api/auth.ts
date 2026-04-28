import { api } from "./client";
import type { ApiResponse, AuthData } from "@/types";

export function register(username: string, email: string, password: string) {
  return api.post<ApiResponse<AuthData>>("/auth/register", {
    username,
    email,
    password,
  });
}

// hello gongskie 

export function login(email: string, password: string) {
  return api.post<ApiResponse<AuthData>>("/auth/login", { email, password });
}

export function logout() {
  return api.post<ApiResponse<null>>("/auth/logout");
}

export function refresh() {
  return api.post<ApiResponse<string>>("/auth/refresh");
}
