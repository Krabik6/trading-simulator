"use client";

import { useQuery } from "@tanstack/react-query";
import { fetchAccount } from "../api/account-api";

export const accountKeys = {
  all: ["account"] as const,
};

export function useAccount() {
  return useQuery({
    queryKey: accountKeys.all,
    queryFn: fetchAccount,
  });
}
