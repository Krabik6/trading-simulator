"use client";

import { useEffect, useRef, useState, useCallback } from "react";
import {
  createChart,
  CandlestickSeries,
  type IChartApi,
  type ISeriesApi,
  type CandlestickData,
  type Time,
} from "lightweight-charts";
import { useQuery } from "@tanstack/react-query";
import { usePriceStore } from "@/entities/price/model/price-store";
import { useActiveSymbol } from "@/features/symbol-selector/SymbolSelector";
import { fetchCandles } from "@/entities/candle/api/candle-api";
import {
  CANDLE_INTERVALS,
  type CandleInterval,
} from "@/entities/candle/model/types";

function getIntervalSeconds(interval: CandleInterval): number {
  const map: Record<CandleInterval, number> = {
    "1s": 1,
    "1m": 60,
    "5m": 300,
    "15m": 900,
    "1h": 3600,
    "4h": 14400,
    "1d": 86400,
    "1w": 604800,
  };
  return map[interval];
}

export function TradingChart() {
  const containerRef = useRef<HTMLDivElement>(null);
  const chartRef = useRef<IChartApi | null>(null);
  const seriesRef = useRef<ISeriesApi<"Candlestick"> | null>(null);
  const currentCandleRef = useRef<CandlestickData | null>(null);

  const symbol = useActiveSymbol((s) => s.symbol);
  const price = usePriceStore((s) => s.prices[symbol]);
  const [interval, setInterval] = useState<CandleInterval>("1m");

  const {
    data: candles,
    isLoading,
    isError,
  } = useQuery({
    queryKey: ["candles", symbol, interval],
    queryFn: () => fetchCandles(symbol, interval, 300),
    staleTime: getIntervalSeconds(interval) * 1000,
    refetchOnWindowFocus: false,
  });

  // Create chart once per symbol/interval change
  useEffect(() => {
    if (!containerRef.current) return;

    const chart = createChart(containerRef.current, {
      layout: {
        background: { color: "transparent" },
        textColor: "#9ca3af",
      },
      grid: {
        vertLines: { color: "rgba(255,255,255,0.04)" },
        horzLines: { color: "rgba(255,255,255,0.04)" },
      },
      width: containerRef.current.clientWidth,
      height: containerRef.current.clientHeight,
      timeScale: { timeVisible: true, secondsVisible: interval === "1s" },
      crosshair: { mode: 0 },
    });

    const series = chart.addSeries(CandlestickSeries, {
      upColor: "#22c55e",
      downColor: "#ef4444",
      borderUpColor: "#22c55e",
      borderDownColor: "#ef4444",
      wickUpColor: "#22c55e",
      wickDownColor: "#ef4444",
    });

    chartRef.current = chart;
    seriesRef.current = series;
    currentCandleRef.current = null;

    const ro = new ResizeObserver((entries) => {
      const { width, height } = entries[0].contentRect;
      chart.resize(width, height);
    });
    ro.observe(containerRef.current);

    return () => {
      ro.disconnect();
      chart.remove();
      chartRef.current = null;
      seriesRef.current = null;
      currentCandleRef.current = null;
    };
  }, [symbol, interval]);

  // Load historical candles — depends on symbol+interval too
  // so it re-runs after chart recreation even if candles are cached
  useEffect(() => {
    if (!candles || !seriesRef.current) return;

    const data: CandlestickData[] = candles.map((c) => ({
      time: c.time as Time,
      open: c.open,
      high: c.high,
      low: c.low,
      close: c.close,
    }));

    seriesRef.current.setData(data);
    chartRef.current?.timeScale().fitContent();

    if (data.length > 0) {
      currentCandleRef.current = data[data.length - 1];
    }
  }, [candles, symbol, interval]);

  // Update current candle from live price ticks
  const updateCandle = useCallback(
    (mid: number) => {
      if (!seriesRef.current) return;

      const intervalSec = getIntervalSeconds(interval);
      const now = Math.floor(Date.now() / 1000);
      const candleTime = (Math.floor(now / intervalSec) * intervalSec) as Time;

      const current = currentCandleRef.current;

      if (current && current.time === candleTime) {
        const updated: CandlestickData = {
          time: candleTime,
          open: current.open,
          high: Math.max(current.high, mid),
          low: Math.min(current.low, mid),
          close: mid,
        };
        currentCandleRef.current = updated;
        seriesRef.current.update(updated);
      } else {
        const newCandle: CandlestickData = {
          time: candleTime,
          open: mid,
          high: mid,
          low: mid,
          close: mid,
        };
        currentCandleRef.current = newCandle;
        seriesRef.current.update(newCandle);
      }
    },
    [interval],
  );

  useEffect(() => {
    if (!price) return;
    updateCandle(price.mid);
  }, [price, updateCandle]);

  return (
    <div className="flex h-full flex-col">
      <div className="flex gap-1 px-2 py-1">
        {CANDLE_INTERVALS.map((iv) => (
          <button
            key={iv.value}
            onClick={() => setInterval(iv.value)}
            className={`rounded px-2 py-0.5 text-xs font-medium transition-colors ${
              interval === iv.value
                ? "bg-blue-600 text-white"
                : "text-muted-foreground hover:bg-muted"
            }`}
          >
            {iv.label}
          </button>
        ))}
      </div>
      <div className="relative min-h-[300px] flex-1">
        <div ref={containerRef} className="absolute inset-0" />
        {isLoading && (
          <div className="text-muted-foreground absolute inset-0 flex items-center justify-center text-sm">
            Loading chart...
          </div>
        )}
        {isError && (
          <div className="absolute inset-0 flex items-center justify-center text-sm text-red-400">
            Failed to load candles
          </div>
        )}
      </div>
    </div>
  );
}
