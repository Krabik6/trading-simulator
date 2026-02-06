"use client";

import { create } from "zustand";

interface LivePosition {
  id: number;
  symbol: string;
  side: string;
  quantity: string;
  entry_price: string;
  mark_price: string;
  unrealized_pnl: string;
  leverage: number;
}

interface PositionState {
  livePositions: Record<number, LivePosition>;
  updatePosition: (p: LivePosition) => void;
  closePosition: (id: number) => void;
  clear: () => void;
}

export const usePositionStore = create<PositionState>((set) => ({
  livePositions: {},

  updatePosition: (p) =>
    set((state) => ({
      livePositions: { ...state.livePositions, [p.id]: p },
    })),

  closePosition: (id) =>
    set((state) => {
      const next = { ...state.livePositions };
      delete next[id];
      return { livePositions: next };
    }),

  clear: () => set({ livePositions: {} }),
}));
