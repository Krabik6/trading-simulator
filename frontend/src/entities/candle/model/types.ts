export interface Candle {
  time: number;
  open: number;
  high: number;
  low: number;
  close: number;
}

export type CandleInterval =
  | "1s"
  | "1m"
  | "5m"
  | "15m"
  | "1h"
  | "4h"
  | "1d"
  | "1w";

export const CANDLE_INTERVALS: { label: string; value: CandleInterval }[] = [
  { label: "1s", value: "1s" },
  { label: "1m", value: "1m" },
  { label: "5m", value: "5m" },
  { label: "15m", value: "15m" },
  { label: "1h", value: "1h" },
  { label: "4h", value: "4h" },
  { label: "1D", value: "1d" },
  { label: "1W", value: "1w" },
];
