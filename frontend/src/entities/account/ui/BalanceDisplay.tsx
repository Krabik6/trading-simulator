"use client";

import { formatUSD } from "@/shared/lib/format";

export function BalanceDisplay({
  balance,
  label,
}: {
  balance: string;
  label: string;
}) {
  return (
    <div className="flex flex-col gap-0.5">
      <span className="text-muted-foreground text-xs">{label}</span>
      <span className="font-mono text-sm font-medium">
        {formatUSD(balance)} USDT
      </span>
    </div>
  );
}
