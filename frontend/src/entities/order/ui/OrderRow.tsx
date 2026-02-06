"use client";

import type { Order } from "../model/types";
import { Badge } from "@/components/ui/badge";
import { formatUSD, formatCrypto, formatDateTime } from "@/shared/lib/format";
import { cn } from "@/lib/utils";

const statusVariant: Record<string, "default" | "secondary" | "destructive" | "outline"> = {
  PENDING: "outline",
  FILLED: "default",
  CANCELLED: "secondary",
  REJECTED: "destructive",
};

export function OrderRow({ order }: { order: Order }) {
  return (
    <>
      <td className="px-3 py-2 text-sm">{order.symbol}</td>
      <td
        className={cn(
          "px-3 py-2 text-sm font-medium",
          order.side === "BUY" ? "text-profit" : "text-loss",
        )}
      >
        {order.side}
      </td>
      <td className="px-3 py-2 text-sm">{order.type}</td>
      <td className="px-3 py-2 text-sm font-mono">{formatCrypto(order.quantity)}</td>
      <td className="px-3 py-2 text-sm font-mono">{formatUSD(order.price)}</td>
      <td className="px-3 py-2 text-sm">{order.leverage}x</td>
      <td className="px-3 py-2">
        <Badge variant={statusVariant[order.status] ?? "secondary"}>
          {order.status}
        </Badge>
      </td>
      <td className="text-muted-foreground px-3 py-2 text-sm">
        {formatDateTime(order.created_at)}
      </td>
    </>
  );
}
