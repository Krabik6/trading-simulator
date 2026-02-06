"use client";

import { usePriceStore } from "@/entities/price/model/price-store";
import { useTicker24h } from "@/entities/price/model/use-ticker24h";
import { useActiveSymbol } from "@/features/symbol-selector/SymbolSelector";
import { useRef, useEffect, useState, useMemo } from "react";
import { cn } from "@/lib/utils";
import { formatCrypto } from "@/shared/lib/format";

export function PriceTicker() {
  const prices = usePriceStore((s) => s.prices);
  const activeSymbol = useActiveSymbol((s) => s.symbol);
  const setSymbol = useActiveSymbol((s) => s.setSymbol);
  const prevPrices = useRef<Record<string, number>>({});
  const [prev, setPrev] = useState<Record<string, number>>({});

  const { data: tickers } = useTicker24h();

  const tickerMap = useMemo(() => {
    if (!tickers) return {};
    const map: Record<string, { changePercent: number }> = {};
    for (const t of tickers) {
      map[t.symbol] = { changePercent: t.priceChangePercent };
    }
    return map;
  }, [tickers]);

  useEffect(() => {
    setPrev({ ...prevPrices.current });
    for (const [symbol, p] of Object.entries(prices)) {
      prevPrices.current[symbol] = p.mid;
    }
  }, [prices]);

  const symbols = Object.values(prices);

  if (symbols.length === 0) {
    return (
      <div className="text-muted-foreground flex h-10 items-center text-sm">
        Waiting for price data...
      </div>
    );
  }

  return (
    <div className="flex items-center gap-2">
      {symbols.map((p) => {
        const isActive = p.symbol === activeSymbol;
        const prevMid = prev[p.symbol];
        const direction =
          prevMid !== undefined
            ? p.mid > prevMid
              ? "up"
              : p.mid < prevMid
                ? "down"
                : "flat"
            : "flat";

        const ticker = tickerMap[p.symbol];
        const changePercent = ticker?.changePercent;

        return (
          <button
            key={p.symbol}
            onClick={() => setSymbol(p.symbol)}
            className={cn(
              "flex items-center gap-2 rounded-lg px-3 py-2 text-left transition-colors",
              isActive ? "bg-accent border" : "hover:bg-accent/50",
            )}
          >
            <span className="text-sm font-semibold">
              {p.symbol.replace("USDT", "")}
            </span>
            <span
              className={cn(
                "font-mono text-sm font-medium transition-colors",
                direction === "up" && "text-profit",
                direction === "down" && "text-loss",
              )}
            >
              {formatCrypto(p.mid, 2)}
            </span>
            {changePercent !== undefined && (
              <span
                className={cn(
                  "font-mono text-xs font-medium",
                  changePercent >= 0 ? "text-profit" : "text-loss",
                )}
              >
                {changePercent >= 0 ? "+" : ""}
                {changePercent.toFixed(2)}%
              </span>
            )}
          </button>
        );
      })}
    </div>
  );
}
