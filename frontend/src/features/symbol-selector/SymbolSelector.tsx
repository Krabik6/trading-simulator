"use client";

import { create } from "zustand";
import { Tabs, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { useSymbols } from "@/entities/symbol/model/use-symbols";

interface ActiveSymbolState {
  symbol: string;
  setSymbol: (s: string) => void;
}

export const useActiveSymbol = create<ActiveSymbolState>((set) => ({
  symbol: "BTCUSDT",
  setSymbol: (symbol) => set({ symbol }),
}));

export function SymbolSelector() {
  const { data: symbols } = useSymbols();
  const active = useActiveSymbol((s) => s.symbol);
  const setSymbol = useActiveSymbol((s) => s.setSymbol);

  const list = symbols ?? [
    { symbol: "BTCUSDT" },
    { symbol: "ETHUSDT" },
    { symbol: "SOLUSDT" },
  ];

  return (
    <Tabs value={active} onValueChange={setSymbol}>
      <TabsList>
        {list.map((s) => (
          <TabsTrigger key={s.symbol} value={s.symbol}>
            {s.symbol.replace("USDT", "")}
          </TabsTrigger>
        ))}
      </TabsList>
    </Tabs>
  );
}
