"use client";

import { formatPnL } from "@/shared/lib/format";
import { cn } from "@/lib/utils";

export function PnLDisplay({
  value,
  percent,
  className,
}: {
  value: string | number;
  percent?: number;
  className?: string;
}) {
  const n = typeof value === "string" ? parseFloat(value) : value;
  return (
    <span
      className={cn(
        "font-mono text-sm font-medium",
        n > 0 && "text-profit",
        n < 0 && "text-loss",
        n === 0 && "text-muted-foreground",
        className,
      )}
    >
      {formatPnL(n)}
      {percent !== undefined && (
        <span className="ml-1 text-xs opacity-70">
          ({percent >= 0 ? "+" : ""}
          {percent.toFixed(2)}%)
        </span>
      )}
    </span>
  );
}
