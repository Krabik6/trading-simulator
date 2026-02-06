"use client";

import { useEffect } from "react";
import { wsManager, type WsMessage } from "@/shared/lib/websocket-manager";
import { useAuthStore } from "@/shared/lib/auth-store";
import { usePriceStore } from "@/entities/price/model/price-store";
import { usePositionStore } from "@/entities/position/model/position-store";
import { getQueryClient } from "@/shared/api/query-client";
import { toast } from "sonner";
import { formatPnL } from "@/shared/lib/format";

export function WebSocketProvider({ children }: { children: React.ReactNode }) {
  const token = useAuthStore((s) => s.token);

  useEffect(() => {
    wsManager.connect(token);

    const unsub = wsManager.subscribe((msg: WsMessage) => {
      switch (msg.type) {
        case "prices":
          usePriceStore
            .getState()
            .updatePrices(msg.data as PriceUpdate[]);
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
            pnl >= 0 ? "Position closed with profit" : "Position closed with loss",
            { description: `PnL: ${formatPnL(close.realized_pnl)} USDT` },
          );
          break;
        }
      }
    });

    return () => {
      unsub();
      wsManager.disconnect();
    };
  }, [token]);

  return <>{children}</>;
}

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
