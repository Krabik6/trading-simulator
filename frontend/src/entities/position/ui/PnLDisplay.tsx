"use client";

import { formatPnL } from "@/shared/lib/format";
import { cn } from "@/lib/utils";

export function PnLDisplay({
  value,
  className,
}: {
  value: string | number;
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
    </span>
  );
}
