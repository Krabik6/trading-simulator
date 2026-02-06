export type TradeSide = "LONG" | "SHORT";
export type TradeType = "OPEN" | "ADD" | "CLOSE" | "LIQUIDATE";

export interface Trade {
  id: number;
  position_id: number;
  order_id: number;
  symbol: string;
  side: TradeSide;
  type: TradeType;
  quantity: string;
  price: string;
  pnl: string;
  fee: string;
  created_at: string;
}
