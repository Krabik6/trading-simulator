"use client";

import { useQuery } from "@tanstack/react-query";
import { fetchUser } from "../api/user-api";
import { useAuthStore } from "@/shared/lib/auth-store";

export const userKeys = {
  me: ["user", "me"] as const,
};

export function useUser() {
  const isAuthenticated = useAuthStore((s) => s.isAuthenticated);
  return useQuery({
    queryKey: userKeys.me,
    queryFn: fetchUser,
    enabled: isAuthenticated,
  });
}
