"use client";

import type { Position } from "../model/types";
import { PnLDisplay } from "./PnLDisplay";
import { formatUSD, formatCrypto } from "@/shared/lib/format";
import { cn } from "@/lib/utils";
import { usePositionStore } from "../model/position-store";

export function PositionRow({ position }: { position: Position }) {
  const live = usePositionStore((s) => s.livePositions[position.id]);
  const pnl = live?.unrealized_pnl ?? position.unrealized_pnl;
  const markPrice = live?.mark_price ?? position.mark_price;

  const pnlNum = typeof pnl === "string" ? parseFloat(pnl) : pnl;
  const margin = parseFloat(position.initial_margin);
  const pnlPercent = margin > 0 ? (pnlNum / margin) * 100 : 0;

  return (
    <>
      <td className="px-3 py-2 text-sm font-medium">{position.symbol}</td>
      <td
        className={cn(
          "px-3 py-2 text-sm font-medium",
          position.side === "LONG" ? "text-profit" : "text-loss",
        )}
      >
        {position.side}
      </td>
      <td className="px-3 py-2 text-sm font-mono">
        {formatCrypto(position.quantity)}
      </td>
      <td className="px-3 py-2 text-sm font-mono">
        {formatUSD(position.entry_price)}
      </td>
      <td className="px-3 py-2 text-sm font-mono">{formatUSD(markPrice)}</td>
      <td className="px-3 py-2 text-sm">{position.leverage}x</td>
      <td className="px-3 py-2">
        <PnLDisplay value={pnl} percent={pnlPercent} />
      </td>
      <td className="px-3 py-2 text-sm font-mono">
        {formatUSD(position.liquidation_price)}
      </td>
    </>
  );
}
