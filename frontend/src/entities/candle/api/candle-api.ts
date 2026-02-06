import { api } from "@/shared/api/client";
import type { Candle, CandleInterval } from "../model/types";

export async function fetchCandles(
  symbol: string,
  interval: CandleInterval,
  limit: number = 300,
): Promise<Candle[]> {
  const { data } = await api.get<Candle[]>("/candles", {
    params: { symbol, interval, limit },
  });
  return data;
}
