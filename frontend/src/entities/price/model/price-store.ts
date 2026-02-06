"use client";

import { create } from "zustand";

interface PriceData {
  symbol: string;
  bid: number;
  ask: number;
  mid: number;
}

interface PriceState {
  prices: Record<string, PriceData>;
  updatePrices: (updates: PriceData[]) => void;
}

export const usePriceStore = create<PriceState>((set) => ({
  prices: {},
  updatePrices: (updates) =>
    set((state) => {
      const next = { ...state.prices };
      for (const p of updates) {
        next[p.symbol] = p;
      }
      return { prices: next };
    }),
}));
