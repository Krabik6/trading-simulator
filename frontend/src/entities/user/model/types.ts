export interface User {
  id: number;
  email: string;
  created_at: string;
}

export interface AuthResponse {
  user_id: number;
  token: string;
}

export interface LoginRequest {
  email: string;
  password: string;
}

export interface RegisterRequest {
  email: string;
  password: string;
}
