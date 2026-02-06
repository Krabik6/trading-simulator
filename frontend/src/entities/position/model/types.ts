export type PositionSide = "LONG" | "SHORT";
export type PositionStatus = "OPEN" | "CLOSED" | "LIQUIDATED";

export interface Position {
  id: number;
  symbol: string;
  side: PositionSide;
  status: PositionStatus;
  quantity: string;
  entry_price: string;
  mark_price: string;
  leverage: number;
  initial_margin: string;
  unrealized_pnl: string;
  realized_pnl: string;
  liquidation_price: string;
  stop_loss: string | null;
  take_profit: string | null;
  created_at: string;
}

export interface UpdateTPSLRequest {
  stop_loss?: string | null;
  take_profit?: string | null;
}
