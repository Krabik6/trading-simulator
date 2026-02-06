"use client";

import { create } from "zustand";

interface AuthState {
  token: string | null;
  userId: number | null;
  isAuthenticated: boolean;
  login: (token: string, userId: number) => void;
  logout: () => void;
  hydrate: () => void;
}

export const useAuthStore = create<AuthState>((set) => ({
  token: null,
  userId: null,
  isAuthenticated: false,

  login: (token, userId) => {
    localStorage.setItem("token", token);
    localStorage.setItem("user_id", String(userId));
    set({ token, userId, isAuthenticated: true });
  },

  logout: () => {
    localStorage.removeItem("token");
    localStorage.removeItem("user_id");
    set({ token: null, userId: null, isAuthenticated: false });
  },

  hydrate: () => {
    const token = localStorage.getItem("token");
    const userId = localStorage.getItem("user_id");
    if (token && userId) {
      set({ token, userId: Number(userId), isAuthenticated: true });
    }
  },
}));
