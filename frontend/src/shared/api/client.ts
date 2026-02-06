import axios from "axios";
import { env } from "@/shared/config/env";

export const api = axios.create({
  baseURL: env.apiUrl,
  headers: { "Content-Type": "application/json" },
});

api.interceptors.request.use((config) => {
  if (typeof window === "undefined") return config;
  const token = localStorage.getItem("token");
  if (token) {
    config.headers.Authorization = `Bearer ${token}`;
  }
  return config;
});

api.interceptors.response.use(
  (response) => response,
  async (error) => {
    const original = error.config;
    if (error.response?.status === 401 && !original._retry) {
      original._retry = true;
      try {
        const token = localStorage.getItem("token");
        if (!token) throw new Error("no token");
        const { data } = await axios.post(
          `${env.apiUrl}/auth/refresh`,
          {},
          { headers: { Authorization: `Bearer ${token}` } },
        );
        localStorage.setItem("token", data.token);
        original.headers.Authorization = `Bearer ${data.token}`;
        return api(original);
      } catch {
        localStorage.removeItem("token");
        localStorage.removeItem("user_id");
        window.location.href = "/en/login";
        return Promise.reject(error);
      }
    }
    return Promise.reject(error);
  },
);
