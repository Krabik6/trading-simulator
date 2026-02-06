"use client";

import { useQuery } from "@tanstack/react-query";
import { fetchTrades } from "../api/trade-api";

export const tradeKeys = {
  all: ["trades"] as const,
  list: (limit: number, offset: number) =>
    ["trades", "list", limit, offset] as const,
};

export function useTrades(limit = 50, offset = 0) {
  return useQuery({
    queryKey: tradeKeys.list(limit, offset),
    queryFn: () => fetchTrades(limit, offset),
  });
}
