"use client";

import { useQuery } from "@tanstack/react-query";
import { fetchTicker24h } from "../api/price-api";
import type { Ticker24h } from "./types";

export function useTicker24h() {
  return useQuery<Ticker24h[]>({
    queryKey: ["ticker24h"],
    queryFn: fetchTicker24h,
    staleTime: 30_000,
    refetchInterval: 30_000,
  });
}
