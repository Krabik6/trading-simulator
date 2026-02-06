"use client";

import { useQuery } from "@tanstack/react-query";
import { fetchSymbols } from "../api/symbol-api";

export const symbolKeys = {
  all: ["symbols"] as const,
};

export function useSymbols() {
  return useQuery({
    queryKey: symbolKeys.all,
    queryFn: fetchSymbols,
    staleTime: 5 * 60 * 1000,
  });
}
