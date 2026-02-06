import { api } from "@/shared/api/client";
import type {
  User,
  AuthResponse,
  LoginRequest,
  RegisterRequest,
} from "../model/types";

export async function login(req: LoginRequest): Promise<AuthResponse> {
  const { data } = await api.post<AuthResponse>("/auth/login", req);
  return data;
}

export async function register(req: RegisterRequest): Promise<AuthResponse> {
  const { data } = await api.post<AuthResponse>("/auth/register", req);
  return data;
}

export async function refreshToken(): Promise<AuthResponse> {
  const { data } = await api.post<AuthResponse>("/auth/refresh");
  return data;
}

export async function fetchUser(): Promise<User> {
  const { data } = await api.get<User>("/user/me");
  return data;
}
