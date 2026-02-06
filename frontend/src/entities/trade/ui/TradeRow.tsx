"use client";

import type { Trade } from "../model/types";
import { Badge } from "@/components/ui/badge";
import { PnLDisplay } from "@/entities/position/ui/PnLDisplay";
import { formatUSD, formatCrypto, formatDateTime } from "@/shared/lib/format";
import { cn } from "@/lib/utils";

const typeVariant: Record<string, "default" | "secondary" | "destructive" | "outline"> = {
  OPEN: "default",
  ADD: "outline",
  CLOSE: "secondary",
  LIQUIDATE: "destructive",
};

export function TradeRow({ trade }: { trade: Trade }) {
  return (
    <>
      <td className="px-3 py-2 text-sm">{trade.symbol}</td>
      <td
        className={cn(
          "px-3 py-2 text-sm font-medium",
          trade.side === "LONG" ? "text-profit" : "text-loss",
        )}
      >
        {trade.side}
      </td>
      <td className="px-3 py-2">
        <Badge variant={typeVariant[trade.type] ?? "secondary"}>
          {trade.type}
        </Badge>
      </td>
      <td className="px-3 py-2 text-sm font-mono">
        {formatCrypto(trade.quantity)}
      </td>
      <td className="px-3 py-2 text-sm font-mono">{formatUSD(trade.price)}</td>
      <td className="px-3 py-2">
        <PnLDisplay value={trade.pnl} />
      </td>
      <td className="px-3 py-2 text-sm font-mono">{formatUSD(trade.fee)}</td>
      <td className="text-muted-foreground px-3 py-2 text-sm">
        {formatDateTime(trade.created_at)}
      </td>
    </>
  );
}
