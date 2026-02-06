"use client";

import { useEffect, useRef } from "react";
import {
  createChart,
  LineSeries,
  type IChartApi,
  type ISeriesApi,
  type LineData,
  type Time,
} from "lightweight-charts";
import { usePriceStore } from "@/entities/price/model/price-store";
import { useActiveSymbol } from "@/features/symbol-selector/SymbolSelector";

export function TradingChart() {
  const containerRef = useRef<HTMLDivElement>(null);
  const chartRef = useRef<IChartApi | null>(null);
  const seriesRef = useRef<ISeriesApi<"Line"> | null>(null);
  const symbol = useActiveSymbol((s) => s.symbol);
  const price = usePriceStore((s) => s.prices[symbol]);

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
      timeScale: { timeVisible: true, secondsVisible: true },
      crosshair: { mode: 0 },
    });

    const series = chart.addSeries(LineSeries, {
      color: "#3b82f6",
      lineWidth: 2,
    });

    chartRef.current = chart;
    seriesRef.current = series;

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
    };
  }, [symbol]);

  useEffect(() => {
    if (!price || !seriesRef.current) return;
    const now = Math.floor(Date.now() / 1000) as Time;
    seriesRef.current.update({ time: now, value: price.mid } as LineData);
  }, [price]);

  return <div ref={containerRef} className="h-full min-h-[300px] w-full" />;
}
