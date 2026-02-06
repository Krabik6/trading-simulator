"use client";

import { useEffect, useRef } from "react";
import { wsManager, type WsMessage } from "@/shared/lib/websocket-manager";
import { useAuthStore } from "@/shared/lib/auth-store";
import { usePriceStore } from "@/entities/price/model/price-store";
import { usePositionStore } from "@/entities/position/model/position-store";
import { getQueryClient } from "@/shared/api/query-client";
import { toast } from "sonner";
import { formatPnL } from "@/shared/lib/format";

interface PriceUpdate {
  symbol: string;
  bid: number;
  ask: number;
  mid: number;
}

interface PositionUpdate {
  id: number;
  symbol: string;
  side: string;
  quantity: string;
  entry_price: string;
  mark_price: string;
  unrealized_pnl: string;
  leverage: number;
}

interface PositionClose {
  position_id: number;
  realized_pnl: string;
}

export function WebSocketProvider({ children }: { children: React.ReactNode }) {
  const token = useAuthStore((s) => s.token);
  const priceBuffer = useRef<Record<string, PriceUpdate>>({});

  useEffect(() => {
    wsManager.connect(token);

    // Flush buffered prices to store once per second — all symbols at once
    const flushInterval = setInterval(() => {
      const buf = priceBuffer.current;
      const symbols = Object.keys(buf);
      if (symbols.length === 0) return;

      const updates = symbols.map((s) => buf[s]);
      priceBuffer.current = {};
      usePriceStore.getState().updatePrices(updates);
    }, 1000);

    const unsub = wsManager.subscribe((msg: WsMessage) => {
      switch (msg.type) {
        case "prices":
          // Buffer latest price per symbol instead of updating store directly
          for (const p of msg.data as PriceUpdate[]) {
            priceBuffer.current[p.symbol] = p;
          }
          break;
        case "position":
          usePositionStore
            .getState()
            .updatePosition(msg.data as PositionUpdate);
          break;
        case "position_close": {
          const close = msg.data as PositionClose;
          usePositionStore.getState().closePosition(close.position_id);
          const qc = getQueryClient();
          qc.invalidateQueries({ queryKey: ["positions"] });
          qc.invalidateQueries({ queryKey: ["account"] });
          qc.invalidateQueries({ queryKey: ["trades"] });
          const pnl = parseFloat(close.realized_pnl);
          toast(
            pnl >= 0
              ? "Position closed with profit"
              : "Position closed with loss",
            { description: `PnL: ${formatPnL(close.realized_pnl)} USDT` },
          );
          break;
        }
      }
    });

    return () => {
      clearInterval(flushInterval);
      unsub();
      wsManager.disconnect();
    };
  }, [token]);

  return <>{children}</>;
}
