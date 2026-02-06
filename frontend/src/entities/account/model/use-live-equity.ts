"use client";

import { useMemo } from "react";
import { useAccount } from "./use-account";
import { usePositionStore } from "@/entities/position/model/position-store";

/**
 * Returns live equity = balance + sum(unrealized PnL from WS).
 * Falls back to the REST-fetched equity when there are no live positions.
 */
export function useLiveEquity() {
  const { data: account } = useAccount();
  const livePositions = usePositionStore((s) => s.livePositions);

  return useMemo(() => {
    if (!account) return null;

    const balance = parseFloat(account.balance);
    const entries = Object.values(livePositions);

    if (entries.length === 0) return parseFloat(account.equity);

    const unrealizedPnl = entries.reduce(
      (sum, p) => sum + parseFloat(p.unrealized_pnl),
      0,
    );

    return balance + unrealizedPnl;
  }, [account, livePositions]);
}
