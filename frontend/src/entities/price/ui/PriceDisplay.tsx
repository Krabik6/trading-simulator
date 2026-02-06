"use client";

import { formatCrypto } from "@/shared/lib/format";
import { cn } from "@/lib/utils";

interface PriceDisplayProps {
  symbol: string;
  bid: number;
  ask: number;
  mid: number;
  prevMid?: number;
  className?: string;
}

export function PriceDisplay({
  symbol,
  bid,
  ask,
  mid,
  prevMid,
  className,
}: PriceDisplayProps) {
  const direction =
    prevMid !== undefined
      ? mid > prevMid
        ? "up"
        : mid < prevMid
          ? "down"
          : "flat"
      : "flat";

  return (
    <div className={cn("flex items-center gap-3", className)}>
      <span className="text-sm font-semibold">{symbol}</span>
      <span
        className={cn(
          "font-mono text-sm font-medium transition-colors",
          direction === "up" && "text-profit",
          direction === "down" && "text-loss",
        )}
      >
        {formatCrypto(mid, symbol.startsWith("BTC") ? 2 : 2)}
      </span>
      <span className="text-muted-foreground text-xs">
        {formatCrypto(bid, 2)} / {formatCrypto(ask, 2)}
      </span>
    </div>
  );
}
