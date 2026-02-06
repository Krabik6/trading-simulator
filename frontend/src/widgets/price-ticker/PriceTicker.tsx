"use client";

import { usePriceStore } from "@/entities/price/model/price-store";
import { PriceDisplay } from "@/entities/price/ui/PriceDisplay";
import { useRef, useEffect, useState } from "react";

export function PriceTicker() {
  const prices = usePriceStore((s) => s.prices);
  const prevPrices = useRef<Record<string, number>>({});
  const [prev, setPrev] = useState<Record<string, number>>({});

  useEffect(() => {
    setPrev({ ...prevPrices.current });
    for (const [symbol, p] of Object.entries(prices)) {
      prevPrices.current[symbol] = p.mid;
    }
  }, [prices]);

  const symbols = Object.values(prices);

  if (symbols.length === 0) {
    return (
      <div className="text-muted-foreground flex h-10 items-center justify-center text-sm">
        Waiting for price data...
      </div>
    );
  }

  return (
    <div className="flex items-center gap-6 overflow-x-auto px-4 py-2">
      {symbols.map((p) => (
        <PriceDisplay
          key={p.symbol}
          symbol={p.symbol}
          bid={p.bid}
          ask={p.ask}
          mid={p.mid}
          prevMid={prev[p.symbol]}
        />
      ))}
    </div>
  );
}
