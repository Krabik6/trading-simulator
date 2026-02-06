export type OrderSide = "BUY" | "SELL";
export type OrderType = "MARKET" | "LIMIT";
export type OrderStatus = "PENDING" | "FILLED" | "CANCELLED" | "REJECTED";

export interface Order {
  id: number;
  symbol: string;
  side: OrderSide;
  type: OrderType;
  status: OrderStatus;
  quantity: string;
  price: string;
  leverage: number;
  stop_loss: string | null;
  take_profit: string | null;
  created_at: string;
}

export interface CreateOrderRequest {
  symbol: string;
  side: OrderSide;
  type: OrderType;
  quantity: string;
  price?: string;
  leverage: number;
  stop_loss?: string;
  take_profit?: string;
}

export interface UpdateOrderRequest {
  price?: string;
  quantity?: string;
  stop_loss?: string | null;
  take_profit?: string | null;
}
